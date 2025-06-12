package main

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/netip"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/cskr/pubsub/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/shynome/bilive-oauth2/v2/db"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
	bilibili "github.com/shynome/openapi-bilibili"
	"github.com/shynome/openapi-bilibili/live"
	"github.com/shynome/openapi-bilibili/live/cmd"
)

var bclient *bilibili.Client
var mainGameID string

var ps = pubsub.New[string, Danmu](1024)

func initBilibili(se *core.ServeEvent) (err error) {

	key := ed25519.NewKeyFromSeed(args.jwtKey)
	bclient = bilibili.NewClient(args.Key, args.Secret)

	eg := se.Router.Group("/bilibili")

	eg.Any("/health", func(e *core.RequestEvent) error {
		ctx := e.Request.Context()
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		heartbeat := bilibili.ApiCall[bilibili.AppKeepAlive, json.RawMessage](bclient, "/v2/app/heartbeat")
		_, err := heartbeat(ctx, bilibili.AppKeepAlive{GameId: mainGameID})
		if err != nil {
			return apis.NewInternalServerError(err.Error(), err)
		}
		return e.String(http.StatusOK, mainGameID)
	})

	defer err0.Then(&err, func() {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		se.App.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
			cancel()
			return e.Next()
		})

		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.RegisteredClaims{
			Subject:  "root",
			Audience: []string{"https://open-live.bilibili.com"},
		})
		tokenStr := try.To1(token.SignedString(key))

		addr := try.To1(netip.ParseAddrPort(se.Server.Addr))
		wslink := fmt.Sprintf("http://127.0.0.1:%d/bilibili/ws-info-keep?IDCode=%s&token=%s", addr.Port(), args.Code, tokenStr)

		go retry.Do(func() (err error) {
			defer err0.Then(&err, nil, func() {
				slog.Error("主连接出错", "error", err)
			})

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			conn, _ := try.To2(websocket.Dial(ctx, wslink, nil))
			var info WebsocketInfo
			try.To(wsjson.Read(ctx, conn, &info))

			mainGameID = info.GameID
			room := live.RoomWith(info.WebsocketInfo, info.GameID)
			ch := try.To1(room.Connect(ctx))
			slog.Info("danmu connected", "game id", info.GameID)
			for msg := range ch {
				switch msg.Cmd {
				case cmd.CmdDanmu:
					go func(data []byte) {
						var danmu cmd.Danmu
						if err := json.Unmarshal(data, &danmu); err != nil {
							slog.Error("parse danmu msg failed.", "error", err)
							return
						}
						d := Danmu{
							OpenID:   danmu.OpenID,
							Content:  danmu.Msg,
							Nickname: danmu.Username,
						}
						ps.TryPub(d, "*")
					}(msg.Data)
				}
			}

			return fmt.Errorf("websocket连接断开")
		},
			retry.Attempts(0),
			retry.MaxDelay(time.Second),
		)
	}, nil)

	eg.Any("/ws-info-keep", func(e *core.RequestEvent) (err error) {
		defer err0.Then(&err, nil, nil)
		w, r := e.Response, e.Request
		q := r.URL.Query()
		IDCode := q.Get("IDCode")
		if IDCode == "" {
			return apis.NewBadRequestError("require query param: IDCode", nil)
		}
		ctx := r.Context()
		now := time.Now().Add(24 * time.Hour).In(shanghai)
		deadline := time.Date(now.Year(), now.Month(), now.Day(), 4, 30, 0, 0, shanghai)
		ctx, cancel := context.WithDeadlineCause(ctx, deadline, errors.New("于预定的次日4:30断开, 请重连"))
		defer cancel()
		ctx, cause := context.WithCancelCause(ctx)
		defer cause(nil)

		app := try.To1(bclient.Open(ctx, args.App, IDCode))
		defer app.Close()
		conn := try.To1(websocket.Accept(w, r, nil))
		defer func() {
			closedMsg := "defer manual close"
			if err := context.Cause(ctx); !errors.Is(err, context.Canceled) {
				closedMsg = err.Error()
			}
			conn.Close(websocket.StatusAbnormalClosure, closedMsg)
		}()
		go func() {
			if err := app.KeepAlive(ctx); err != nil {
				cause(err)
			} else {
				cause(nil)
			}
		}()
		info := WebsocketInfo{
			WebsocketInfo: app.Info().WebsocketInfo,
			GameID:        app.Info().GameInfo.GameId,
		}
		try.To(wsjson.Write(ctx, conn, info))

		logger := e.App.Logger()
		for {
			var linked LinkedOpenID
			// 客户端询问 OpenID 对应的 UID
			if err := wsjson.Read(ctx, conn, &linked); err != nil {
				return err
			}
			go func() {
				record, err := e.App.FindFirstRecordByData(db.TableLinkeds, "openid", linked.OpenID)
				if err != nil {
					logger.Error("找不到对应的UID", "linked", linked, "error", err)
					return
				}
				linked.UID = record.GetString("uid")
				wsjson.Write(ctx, conn, linked)
			}()
		}
	})

	return se.Next()
}

var shanghai = try.To1(time.LoadLocation("Asia/Shanghai"))

type WebsocketInfo struct {
	bilibili.WebsocketInfo
	GameID string `json:"game_id"`
}

type LinkedOpenID struct {
	OpenID string `json:"openid"`
	UID    string `json:"uid"`
}
