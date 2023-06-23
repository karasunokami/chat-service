package middlewares

import (
	"time"

	"github.com/karasunokami/chat-service/internal/types"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

func AuthWith(uid types.UserID) echo.MiddlewareFunc {
	return authWith(uid, time.Now().Add(time.Hour).Unix())
}

func AuthWithExp(uid types.UserID, exp int64) echo.MiddlewareFunc {
	return authWith(uid, exp)
}

func SetToken(c echo.Context, uid types.UserID) {
	c.Set(tokenCtxKey, &jwt.Token{Claims: &claimsMock{uid: uid, exp: time.Now().Add(time.Hour).Unix()}, Valid: true})
}

func setTokenWithExp(c echo.Context, uid types.UserID, exp int64) {
	c.Set(tokenCtxKey, &jwt.Token{Claims: &claimsMock{uid: uid, exp: exp}, Valid: true})
}

func authWith(uid types.UserID, exp int64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			setTokenWithExp(c, uid, exp)

			return next(c)
		}
	}
}

type claimsMock struct {
	uid types.UserID
	exp int64
}

func (m *claimsMock) Valid() error {
	return nil
}

func (m *claimsMock) UserID() types.UserID {
	return m.uid
}

func (m *claimsMock) ExpiresAt() int64 {
	return m.exp
}
