package middlewares

import (
	"context"
	"errors"
	"fmt"

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
		KeyLookup:  "header:Authorization",
		AuthScheme: "Bearer",
		Validator: func(tokenStr string, eCtx echo.Context) (bool, error) {
			ctx := context.Background()

			res, err := introspector.IntrospectToken(ctx, tokenStr)
			if err != nil {
				return false, fmt.Errorf("introspect token, err=%w", err)
			}

			if !res.Active {
				return false, ErrTokenIsNotActive
			}

			cl := claims{}

			token, _, err := jwt.NewParser().ParseUnverified(tokenStr, &cl)
			if err != nil {
				return false, fmt.Errorf("jwt parse with claims, err=%v", err)
			}

			err = cl.Valid()
			if err != nil {
				return false, fmt.Errorf("validate claims, err=%w", err)
			}

			if !cl.hasRoleForResource(role, resource) {
				return false, ErrNoRequiredResourceRole
			}

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

func userID(eCtx echo.Context) (types.UserID, bool) {
	t := eCtx.Get(tokenCtxKey)
	if t == nil {
		return types.UserIDNil, false
	}

	tt, ok := t.(*jwt.Token)
	if !ok {
		return types.UserIDNil, false
	}

	userIDProvider, ok := tt.Claims.(interface{ UserID() types.UserID })
	if !ok {
		return types.UserIDNil, false
	}
	return userIDProvider.UserID(), true
}
