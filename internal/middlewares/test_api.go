package middlewares

import (
	"github.com/karasunokami/chat-service/internal/types"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

type testAPIClaims struct {
	Subject types.UserID `json:"sub,omitempty"`
}

func (c *testAPIClaims) Valid() error {
	return nil
}

func (c *testAPIClaims) UserID() types.UserID {
	return c.Subject
}

func SetToken(c echo.Context, uid types.UserID) {
	c.Set(tokenCtxKey, &jwt.Token{
		Claims: &testAPIClaims{
			Subject: uid,
		},
	})
}
