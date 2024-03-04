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
	"github.com/shynome/err0"
	bilibili "github.com/shynome/openapi-bilibili"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
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
			if ok := claims.VerifyAudience("https://open-live.bilibili.com", true); !ok {
				return nil, fmt.Errorf("用途错误")
			}
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

	e.Any("/ws-info-keep", func(c echo.Context) (err error) {
		defer err0.Then(&err, nil, nil)
		IDCode := c.QueryParam("IDCode")
		if IDCode == "" {
			return echo.NewHTTPError(400, "require query param: IDCode")
		}
		r, w := c.Request(), c.Response()
		ctx := r.Context()
		app := try.To1(bclient.Open(ctx, appid, IDCode))
		defer app.Close()
		conn := try.To1(websocket.Accept(w, r, nil))
		defer conn.Close(websocket.StatusAbnormalClosure, "defer manual close")
		go func() {
			var closeMsg = "manual close"
			if err := app.KeepAlive(ctx); err != nil {
				// do nothing
				closeMsg = err.Error()
			}
			conn.Close(websocket.StatusAbnormalClosure, closeMsg)
		}()
		info := app.Info().WebsocketInfo
		try.To(wsjson.Write(ctx, conn, info))
		for {
			if _, _, err := conn.Read(ctx); err != nil {
				return err
			}
		}
	})
}
