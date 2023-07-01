package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-session/session"
	"github.com/labstack/echo/v4"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/shynome/bilive"
	"github.com/tidwall/buntdb"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type BiliveMsg struct {
	CMD bilive.CMD `json:"cmd"`
}
type BiliveDanmu struct {
	Info []any `json:"info"`
}
type Config struct {
	Room int    `json:"room"`
	Code string `json:"code"`
}

func registerBiliveServer(e *echo.Group, roomid int) {
	cache := try.To1(buntdb.Open(":memory:"))
	e.Any("/pair", func(c echo.Context) (err error) {
		defer err2.Handle(&err, func() {
			// log.Println(err)
		})
		w, r := c.Response(), c.Request()

		ctx := r.Context()
		store := try.To1(session.Start(ctx, w, r))

		conn := try.To1(websocket.Accept(w, r, nil))
		defer conn.Close(websocket.StatusAbnormalClosure, "")

		lc := bilive.NewClient(roomid)
		try.To(lc.Connect())
		defer lc.Close()

		const ttl = 10 * time.Minute
		ctx, cancel := context.WithTimeout(ctx, ttl)
		defer cancel()

		go func() { // 修复ws连接一直不断开的问题
			defer cancel()
			for {
				if _, _, err := conn.Read(ctx); err != nil {
					break
				}
			}
		}()

		var vid string
		err = cache.Update(func(tx *buntdb.Tx) (err error) {
			defer err2.Handle(&err)
			for i := 0; i < 5; i++ {
				vid = try.To1(randomHex(8))
				_, ierr := tx.Get(vid)
				if errors.Is(ierr, buntdb.ErrNotFound) {
					_, _, err := tx.Set(vid, "yes", &buntdb.SetOptions{TTL: ttl})
					return err
				}
			}
			return fmt.Errorf("gen vid failed")
		})
		try.To(err)
		try.To(wsjson.Write(ctx, conn, Config{Room: roomid, Code: vid}))

		for {
			_, b := try.To2(lc.Conn.Read(ctx))
			// fmt.Println(7777)
			go func(b []byte) {
				defer err2.Catch(func(err error) {
					// log.Printf("decode packet %d err: %e", roomid, err)
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
						store.Save()
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
		store.Delete("uid")
		store.Delete("l-uid")
		store.Save()
		return
	})
}

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}