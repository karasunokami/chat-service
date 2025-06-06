package middlewares

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	keycloakclient "github.com/karasunokami/chat-service/internal/clients/keycloak"
	"github.com/karasunokami/chat-service/internal/types"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/introspector_mock.gen.go -package=middlewaresmocks Introspector

const tokenCtxKey = "user-token"

var (
	ErrNoRequiredResourceRole = errors.New("no required resource role")
	ErrTokenIsNotActive       = errors.New("token is not active")
)

type Introspector interface {
	IntrospectToken(ctx context.Context, token string) (*keycloakclient.IntrospectTokenResult, error)
}

// NewKeyCloakTokenAuth returns a middleware that implements "active" authentication:
// each request is verified by the Keycloak server.
func NewKeyCloakTokenAuth(introspector Introspector, resource, role string) echo.MiddlewareFunc {
	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		KeyLookup:  "header:Authorization,header:Sec-WebSocket-Protocol",
		AuthScheme: "Bearer",
		Validator: func(tokenStr string, eCtx echo.Context) (bool, error) {
			extractedTokenStr := extractToken(tokenStr)

			res, err := introspector.IntrospectToken(eCtx.Request().Context(), extractedTokenStr)
			if err != nil {
				return false, fmt.Errorf("introspect token, err=%w", err)
			}

			if !res.Active {
				return false, ErrTokenIsNotActive
			}

			cl := claims{}

			token, _, err := jwt.NewParser().ParseUnverified(extractedTokenStr, &cl)
			if err != nil {
				return false, fmt.Errorf("jwt parse with claims, err=%v", err)
			}

			err = cl.Valid()
			if err != nil {
				return false, fmt.Errorf("validate claims, err=%w", err)
			}

			if !cl.ResourcesAccess.HasResourceRole(resource, role) {
				return false, ErrNoRequiredResourceRole
			}

			// Copy standard exp field to custom exp claims field to prevent
			// overriding by json decode
			cl.Exp = cl.ExpiresAt

			eCtx.Set(tokenCtxKey, token)

			return true, nil
		},
	})
}

func MustUserID(eCtx echo.Context) types.UserID {
	uid, ok := userID(eCtx)
	if !ok {
		panic("no user token in request context")
	}
	return uid
}

func MustExpiresAt(eCtx echo.Context) time.Time {
	exp, ok := expiresAt(eCtx)
	if !ok {
		panic("no exp in request context")
	}

	return time.Unix(exp, 0)
}

func extractToken(tokenStr string) string {
	parts := strings.Split(tokenStr, ", ")

	if len(parts) >= 2 {
		return parts[1]
	}

	return parts[0]
}

func userID(eCtx echo.Context) (types.UserID, bool) {
	tt, ok := extractTokenFromContext(eCtx)
	if !ok {
		return types.UserIDNil, false
	}

	userIDProvider, ok := tt.Claims.(interface{ UserID() types.UserID })
	if !ok {
		return types.UserIDNil, false
	}
	return userIDProvider.UserID(), true
}

func expiresAt(eCtx echo.Context) (int64, bool) {
	tt, ok := extractTokenFromContext(eCtx)
	if !ok {
		return 0, false
	}

	if expProvider, ok := tt.Claims.(interface{ ExpiresAt() int64 }); ok {
		return expProvider.ExpiresAt(), true
	}

	if expProvider, ok := tt.Claims.(interface{ ExpiresAtUnix() int64 }); ok {
		return expProvider.ExpiresAtUnix(), true
	}

	return 0, false
}

func extractTokenFromContext(eCtx echo.Context) (*jwt.Token, bool) {
	t := eCtx.Get(tokenCtxKey)
	if t == nil {
		return nil, false
	}

	tt, ok := t.(*jwt.Token)
	if !ok {
		return nil, false
	}

	return tt, true
}
