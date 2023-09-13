package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/shynome/bilive-oauth2/danmu"
	"github.com/tidwall/buntdb"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type Config struct {
	Room int    `json:"room"`
	Code string `json:"code"`
}
type Danmu struct {
	UID     string `json:"uid"`
	Content string `json:"content"`
}

type MsgType string

const (
	MsgInit    MsgType = "init"
	MsgDanmu   MsgType = "danmu"
	MsgVerfied MsgType = "verified"
)

type Msg[T any] struct {
	Type MsgType `json:"type"`
	Data T       `json:"data"`
}

type VerifiedMsg struct {
	Token string `json:"token"`
}

func registerBiliveServer(e *echo.Group, privkey []byte, roomid int, bilipage string) {
	key := try.To1(jwt.ParseEdPrivateKeyFromPEM(privkey))
	cache := try.To1(buntdb.Open(":memory:"))

	dd := NewDisptacher[Danmu]()

	r, cmd := danmu.Connect(fmt.Sprintf("%d", roomid), bilipage)
	try.To(cmd.Start())
	go func() {
		for {
			line, _ := try.To2(r.ReadLine())
			go func(line string) {
				arr := strings.SplitN(line, "|", 2)
				if len(arr) != 2 {
					return
				}
				dd.Dispatch(Danmu{
					UID:     arr[0],
					Content: arr[1],
				})
			}(string(line))
		}
	}()

	e.Any("/pair", func(c echo.Context) (err error) {
		defer err2.Handle(&err, func() {
			// log.Println(err)
		})
		w, r := c.Response(), c.Request()

		ctx := r.Context()

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

		done, l := ctx.Done(), dd.Listen(vid)
		defer dd.Free(vid)

		try.To(wsjson.Write(ctx, conn, Msg[Config]{
			Type: MsgInit,
			Data: Config{Room: roomid, Code: vid},
		}))

		for {
			select {
			case <-done:
				return
			case danmu := <-l:
				go wsjson.Write(ctx, conn, Msg[Danmu]{
					Type: MsgDanmu,
					Data: danmu,
				})
				if danmu.Content == vid {
					claims := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{
						"sub": danmu.UID,
					})
					token := try.To1(claims.SignedString(key))
					wsjson.Write(ctx, conn, Msg[VerifiedMsg]{
						Type: MsgVerfied,
						Data: VerifiedMsg{
							Token: token,
						},
					})
					return
				}
			}
		}
	})
}

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
