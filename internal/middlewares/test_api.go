package middlewares

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/karasunokami/chat-service/internal/types"
	"github.com/labstack/echo/v4"
)

type testApiClaims struct {
	Subject types.UserID `json:"sub,omitempty"`
}

func (c *testApiClaims) Valid() error {
	return nil
}

func (c *testApiClaims) UserID() types.UserID {
	return c.Subject
}

func SetToken(c echo.Context, uid types.UserID) {

	c.Set(tokenCtxKey, &jwt.Token{
		Claims: &testApiClaims{
			Subject: uid,
		},
	})
}
