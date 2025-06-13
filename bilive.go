package main

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/shynome/bilive-oauth2/v2/db"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
)

func initBiliveServer(se *core.ServeEvent) (err error) {
	defer err0.Then(&err, nil, nil)

	key := ed25519.NewKeyFromSeed(args.jwtKey)

	vids := try.To1(se.App.FindCachedCollectionByNameOrId(db.TableTmpVIDs))

	eg := se.Router.Group("/bilive")

	eg.Any("/pair2", func(e *core.RequestEvent) (err error) {
		defer err0.Then(&err, nil, nil)

		w, r := e.Response, e.Request
		ctx := r.Context()

		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		const ttl = 10 * time.Minute
		ctx, cancel := context.WithTimeout(ctx, ttl)
		defer cancel()

		flusher, ok := w.(http.Flusher)
		if !ok {
			return apis.NewInternalServerError("http.ResponseWriter does not implement http.Flusher", nil)
		}
		stream := StreamWriter{Writer: w, Flusher: flusher}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		vid := core.NewRecord(vids)
		try.To(e.App.Save(vid))
		defer e.App.Delete(vid)

		l := ps.Sub("*")
		defer ps.Unsub(l)

		msg := Msg[Config]{
			Type: MsgInit,
			Data: Config{Room: fmt.Sprintf("%d", args.Room), Code: vid.Id},
		}
		msg.WriteTry(stream)

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				try.To1(fmt.Fprintf(stream, ": hack for pass cdn \n"))
				stream.Flush()
			case danmu := <-l:
				msg := Msg[Danmu]{
					Type: MsgDanmu,
					Data: danmu,
				}
				msg.WriteTry(stream)
				if danmu.Content == vid.Id {
					now := time.Now()
					claims := jwt.NewWithClaims(jwt.SigningMethodEdDSA, CustomClaims{
						RegisteredClaims: jwt.RegisteredClaims{
							Subject:   danmu.OpenID,
							Issuer:    "https://bilive-auth.remoon.cn/",
							IssuedAt:  jwt.NewNumericDate(now),
							NotBefore: jwt.NewNumericDate(now),
							ExpiresAt: jwt.NewNumericDate(now.AddDate(0, 0, 7)),
						},
						Nickname: danmu.Nickname,
					})
					token := try.To1(claims.SignedString(key))
					msg := Msg[VerifiedMsg]{
						Type: MsgVerfied,
						Data: VerifiedMsg{Token: token},
					}
					// 只有用户来直播间并验证通过的时候才尝试绑定UID, 被风控搞麻了, 尽量少请求几次接口
					if linkedClient != nil {
						go TryLinkUnameUID(e.App, danmu.OpenID, danmu.Nickname)
					}
					msg.WriteTry(stream)
					return
				}
			}
		}
	})
	return se.Next()
}

type StreamWriter struct {
	http.Flusher
	io.Writer
}

type Config struct {
	Room string `json:"room"`
	Code string `json:"code"`
}

type Danmu struct {
	OpenID   string `json:"open_id"`
	Content  string `json:"content"`
	Nickname string `json:"nickname"`
}

type CustomClaims struct {
	jwt.RegisteredClaims
	Nickname string `json:"nickname"`
}

type VerifiedMsg struct {
	Token string `json:"token"`
}
