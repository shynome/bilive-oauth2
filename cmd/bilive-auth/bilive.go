package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-session/session"
	"github.com/labstack/echo/v4"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/rs/xid"
	"github.com/shynome/bilive"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type BiliveMsg struct {
	CMD bilive.CMD `json:"cmd"`
}
type BiliveDanmu struct {
	Info []any `json:"info"`
}

func registerBiliveServer(e *echo.Group, roomid int) {
	e.Any("/pair", func(c echo.Context) (err error) {
		defer err2.Handle(&err)
		w, r := c.Response(), c.Request()

		ctx := r.Context()
		store := try.To1(session.Start(ctx, w, r))
		defer store.Save()

		conn := try.To1(websocket.Accept(w, r, nil))
		defer conn.Close(websocket.StatusAbnormalClosure, "")

		lc := bilive.NewClient(roomid)
		try.To(lc.Connect())
		defer lc.Close()

		ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()

		vid := xid.New().String()
		conn.Write(ctx, websocket.MessageText, []byte(vid))

		for {
			_, b := try.To2(lc.Conn.Read(ctx))
			go func(b []byte) {
				defer err2.Catch(func(err error) {
					log.Printf("decode packet %d err: %e", roomid, err)
				})
				pkts := try.To1(bilive.DecodePackets(b))
				for _, pkt := range pkts {
					go func(pkt *bilive.Packet) {
						defer err2.Catch(func(err error) {
						})
						if pkt.Operation != bilive.OpreationMessage {
							return
						}
						var msg BiliveMsg
						try.To(json.Unmarshal(pkt.Body, &msg))
						if msg.CMD != bilive.CMD_DANMU_MSG {
							return
						}
						var danmu BiliveDanmu
						try.To(json.Unmarshal(pkt.Body, &danmu))
						go wsjson.Write(ctx, conn, danmu)

						infos := danmu.Info
						if len(infos) < 3 {
							return
						}
						d, ok := infos[1].(string)
						if !ok {
							return
						}

						if d != vid {
							return
						}

						userInfos, ok := infos[2].([]any)
						if !ok || len(userInfos) < 2 {
							return
						}
						user, ok := userInfos[0].(float64)
						if !ok {
							return
						}
						uid := fmt.Sprintf("%d", int(user))
						store.Set("uid", uid)
						cancel()
					}(pkt)
				}
			}(b)
		}
	})
	e.Any("/whoami", func(c echo.Context) (err error) {
		defer err2.Handle(&err)
		w, r := c.Response(), c.Request()
		ctx := r.Context()
		store := try.To1(session.Start(ctx, w, r))
		_uid, ok := store.Get("uid")
		if !ok {
			return c.NoContent(http.StatusNotFound)
		}
		uid, ok := _uid.(string)
		if !ok {
			return c.NoContent(http.StatusNotFound)
		}
		return c.String(http.StatusOK, uid)
	})
	e.Any("/logout", func(c echo.Context) (err error) {
		defer err2.Handle(&err)
		w, r := c.Response(), c.Request()
		ctx := r.Context()
		store := try.To1(session.Start(ctx, w, r))
		store.Flush()
		store.Save()
		return
	})
}
