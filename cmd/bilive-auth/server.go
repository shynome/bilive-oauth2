package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/rs/xid"
	bilibili "github.com/shynome/openapi-bilibili"
	"github.com/shynome/openapi-bilibili/live"
	"github.com/shynome/openapi-bilibili/live/cmd"
)

//go:embed all:build
var frontendFiles embed.FS

var args struct {
	addr   string
	pg     string
	room   int
	secret string
	jwtKey string

	bilicode string // 主播身份码
	biliapp  string // B站 key, serect, appid
}

type BiliApp struct {
	ID     int64  `json:"appid"`
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

var Version = "dev"
var f = flag.NewFlagSet("bilive-oauth2@"+Version, flag.ExitOnError)

func init() {
	f.StringVar(&args.addr, "addr", ":9096", "http server listen addr")
	f.StringVar(&args.pg, "pg", "postgres://postgres:postgres@localhost:5432/postgres", "token file db")
	f.StringVar(&args.secret, "secret", xid.New().String(), "cookie secret")
	f.StringVar(&args.jwtKey, "jwt-key", "./bilive-jwt-key", "jwt ed25519 private key")
	f.IntVar(&args.room, "room", 27352037, "room id")
	f.StringVar(&args.bilicode, "bilicode", os.Getenv("BILI_CODE"), "身份码")
	f.StringVar(&args.biliapp, "biliapp", os.Getenv("BILI_APP"), "身份码")
}

func main() {
	f.Parse(os.Args[1:])

	e := echo.New()

	frontend := try.To1(fs.Sub(frontendFiles, "build"))
	assertHandler := http.FileServer(http.FS(frontend))
	e.GET("*", echo.WrapHandler(assertHandler))

	var key = try.To1(os.ReadFile(args.jwtKey))
	srv := initOAuth2Server(args.pg, key)
	registerOAuth2Server(e.Group("/oauth"), key, srv)

	danmuCh := make(chan cmd.Danmu, 1024)
	{
		var biliApp BiliApp
		try.To(json.Unmarshal([]byte(args.biliapp), &biliApp))
		bclient := bilibili.NewClient(biliApp.Key, biliApp.Secret)
		ctx := context.Background()
		getInfo := func() (_ bilibili.WebsocketInfo, err error) {
			defer err2.Handle(&err)
			app := try.To1(bclient.Open(ctx, biliApp.ID, args.bilicode))
			try.To(app.Close())
			info := app.Info().WebsocketInfo
			return info, nil
		}
		linkDanmu := func(data []byte) {
			defer err2.Catch(func(err error) {
				log.Println("parse danmu msg failed:", err)
			})
			var danmu cmd.Danmu
			try.To(json.Unmarshal(data, &danmu))
			danmuCh <- danmu
		}
		wctx, casue := context.WithCancelCause(ctx)
		go func() {
			info := try.To1(getInfo())
			room := live.RoomWith(info)
			connect := func() {
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()
				ch, err := room.Connect(ctx)
				casue(err)
				if err != nil {
					if errors.Is(err, live.ErrAuthFailed) {
						info, err1 := getInfo()
						if err1 != nil {
							err = errors.Join(err, err1)
						}
						room = live.RoomWith(info)
					}
					log.Println("danmu channel connect error", err)
					return
				}
				log.Println("danmu connected")
				for msg := range ch {
					switch msg.Cmd {
					case cmd.CmdDanmu:
						go linkDanmu(msg.Data)
					}
				}
			}
			for {
				connect()
				time.Sleep(time.Second) // 等待1s后再重试, 避免重试过快导致占满资源
			}
		}()
		<-wctx.Done()
		if err := context.Cause(wctx); !errors.Is(err, context.Canceled) {
			try.To(err)
		}
	}

	registerBiliveServer(e.Group("/bilive"), key, args.room, danmuCh)

	log.Println(f.Name(), "start")
	try.To(e.Start(args.addr))
}
