package main

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	bilibili "github.com/shynome/openapi-bilibili"
)

func registerBilibiliApi(e *echo.Group, privateKey ed25519.PrivateKey, bclient *bilibili.Client, appid int64) {

	pubkey := privateKey.Public()
	e.Use(echojwt.WithConfig(echojwt.Config{
		TokenLookup: "header:Authorization:Bearer ,form:token",
		ParseTokenFunc: func(c echo.Context, auth string) (interface{}, error) {
			token, err := jwt.ParseWithClaims(
				auth, new(jwt.StandardClaims),
				func(t *jwt.Token) (interface{}, error) { return pubkey, nil },
			)
			if err != nil {
				return nil, err
			}
			claims := token.Claims.(*jwt.StandardClaims)
			if claims.Subject != "root" {
				return nil, fmt.Errorf("该接口只允许 root 用户访问")
			}
			return token, nil
		},
	}))

	e.Any("/ws-info", func(c echo.Context) (err error) {
		defer err2.Handle(&err, func() {
			err = echo.NewHTTPError(400, err.Error())
		})
		ctx := c.Request().Context()
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		IDCode := c.QueryParam("IDCode")
		if IDCode == "" {
			return c.String(400, "require query param: IDCode")
		}
		app := try.To1(bclient.Open(ctx, appid, IDCode))
		defer app.Close()
		info := app.Info().WebsocketInfo
		return c.JSON(http.StatusOK, info)
	})
}
