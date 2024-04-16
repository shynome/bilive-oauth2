package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
	"github.com/shynome/openapi-bilibili/live/cmd"
	"github.com/tidwall/buntdb"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type Config struct {
	Room string `json:"room"`
	Code string `json:"code"`
}
type Danmu struct {
	OpenID   string `json:"open_id"`
	Content  string `json:"content"`
	Nickname string `json:"nickname"`
}

type BeerDanmu struct {
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

func registerBiliveServer(e *echo.Group, key ed25519.PrivateKey, roomid int, ch <-chan cmd.Danmu) {
	cache := try.To1(buntdb.Open(":memory:"))

	dd := NewDisptacher[Danmu]()

	go func() {
		for danmu := range ch {
			d := Danmu{
				OpenID:   danmu.OpenID,
				Content:  danmu.Msg,
				Nickname: danmu.Username,
			}
			dd.Dispatch(d)
		}
	}()

	e.Any("/pair", func(c echo.Context) (err error) {
		defer err0.Then(&err, nil, func() {
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
			defer err0.Then(&err, nil, nil)
			for i := 0; i < 5; i++ {
				vid = try.To1(randomHex(8))
				_, ierr := tx.Get(vid)
				if errors.Is(ierr, buntdb.ErrNotFound) {
					_, _, err := tx.Set(vid, "yes", &buntdb.SetOptions{Expires: true, TTL: ttl})
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
			Data: Config{Room: fmt.Sprintf("%d", roomid), Code: vid},
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
					now := time.Now()
					claims := jwt.NewWithClaims(jwt.SigningMethodEdDSA, CustomClaims{
						StandardClaims: jwt.StandardClaims{
							Subject:   danmu.OpenID,
							Issuer:    "https://bilive-auth.remoon.cn/",
							IssuedAt:  now.Unix(),
							NotBefore: now.Unix(),
							ExpiresAt: now.AddDate(0, 0, 7).Unix(),
						},
						Nickname: danmu.Nickname,
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

	e.Any("/pair2", func(c echo.Context) (err error) {
		defer err0.Then(&err, nil, func() {
			log.Println(err)
		})
		w, r := c.Response(), c.Request()

		ctx := r.Context()

		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		const ttl = 10 * time.Minute
		ctx, cancel := context.WithTimeout(ctx, ttl)
		defer cancel()

		reader, writer := io.Pipe()
		stream := StreamWriter{Writer: writer, Flusher: w}
		go func() {
			defer cancel()
			c.Stream(http.StatusOK, "text/event-stream", reader)
			return
		}()

		var vid string
		err = cache.Update(func(tx *buntdb.Tx) (err error) {
			defer err0.Then(&err, nil, nil)
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

		msg := Msg[Config]{
			Type: MsgInit,
			Data: Config{Room: fmt.Sprintf("%d", roomid), Code: vid},
		}
		try.To(msg.Write(stream))

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				try.To1(fmt.Fprintf(stream, ": hack for pass cdn \n"))
				stream.Flush()
			case danmu := <-l:
				msg := Msg[Danmu]{
					Type: MsgDanmu,
					Data: danmu,
				}
				try.To(msg.Write(stream))
				if danmu.Content == vid {
					now := time.Now()
					claims := jwt.NewWithClaims(jwt.SigningMethodEdDSA, CustomClaims{
						StandardClaims: jwt.StandardClaims{
							Subject:   danmu.OpenID,
							Issuer:    "https://bilive-auth.remoon.cn/",
							IssuedAt:  now.Unix(),
							NotBefore: now.Unix(),
							ExpiresAt: now.AddDate(0, 0, 7).Unix(),
						},
						Nickname: danmu.Nickname,
					})
					token := try.To1(claims.SignedString(key))
					msg := Msg[VerifiedMsg]{
						Type: MsgVerfied,
						Data: VerifiedMsg{
							Token: token,
						},
					}
					try.To(msg.Write(stream))
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

func (msg Msg[T]) Write(w StreamWriter) (err error) {
	defer w.Flush()
	id := time.Now().Unix()
	try.To1(fmt.Fprintf(w, "id:%d\n", id))
	try.To1(io.WriteString(w, "data:"))
	try.To(json.NewEncoder(w).Encode(msg))
	try.To1(io.WriteString(w, "\n"))
	// 结束该 Event
	try.To1(io.WriteString(w, "\n"))
	return nil
}

type StreamWriter struct {
	http.Flusher
	io.Writer
}

type CustomClaims struct {
	jwt.StandardClaims
	Nickname string `json:"nickname"`
}
