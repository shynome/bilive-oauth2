package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/SierraSoftworks/multicast/v2"
	"github.com/go-session/session"
	"github.com/labstack/echo/v4"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/shynome/bilive-oauth2/danmu"
	"github.com/tidwall/buntdb"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type BiliveDanmu struct {
	Info []any `json:"info"`
}
type Config struct {
	Room int    `json:"room"`
	Code string `json:"code"`
}
type Danmu struct {
	UID     string
	Content string
}

func registerBiliveServer(e *echo.Group, roomid int) {
	cache := try.To1(buntdb.Open(":memory:"))

	dd := multicast.New[Danmu]()

	r, cmd := danmu.Connect(fmt.Sprintf("%d", roomid))
	try.To(cmd.Start())
	go func() {
		for {
			line, _ := try.To2(r.ReadLine())
			go func(line string) {
				arr := strings.SplitN(line, "|", 2)
				if len(arr) != 2 {
					return
				}
				dd.C <- Danmu{UID: arr[0], Content: arr[1]}
			}(string(line))
		}
	}()

	e.Any("/pair", func(c echo.Context) (err error) {
		defer err2.Handle(&err, func() {
			// log.Println(err)
		})
		w, r := c.Response(), c.Request()

		ctx := r.Context()
		store := try.To1(session.Start(ctx, w, r))

		conn := try.To1(websocket.Accept(w, r, nil))
		defer conn.Close(websocket.StatusAbnormalClosure, "")

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

		done, l := ctx.Done(), dd.Listen()
		for {
			select {
			case <-done:
				return
			case danmu := <-l.C:
				go wsjson.Write(ctx, conn, BiliveDanmu{
					Info: []any{
						[]any{},
						danmu.Content,
						[]any{0, "danmu"},
					},
				})
				if danmu.Content == vid {
					store.Set("uid", danmu.UID)
					store.Save()
					return
				}
			}
		}
		return
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
