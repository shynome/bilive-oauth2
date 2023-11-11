package main

import (
	"context"
	"crypto/ed25519"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	bilibili "github.com/shynome/openapi-bilibili"
	"github.com/shynome/openapi-bilibili/live"
	"github.com/shynome/openapi-bilibili/live/cmd"
	"github.com/spf13/viper"
)

//go:embed all:build
var frontendFiles embed.FS

var args struct {
	addr   string
	jwtKey string
}

var Version = "dev"
var f = flag.NewFlagSet("bilive-oauth2@"+Version, flag.ExitOnError)

type BiliveAuthConfig struct {
	Clients  []OAuthClient
	Bilibili BilibiliLiveConfig
}

type BilibiliLiveConfig struct {
	Key    string
	Secret string
	App    int64  // 应用 ID
	Code   string // 身份码
	Room   int    // 直播间, 虽然身份码可以拿到直播间号, 但还是直接写一下吧
}

var oc BiliveAuthConfig

func init() {
	viper.SetConfigName("bilive-auth")
	viper.AddConfigPath(".")

	f.StringVar(&args.addr, "addr", ":9096", "http server listen addr")
	f.StringVar(&args.jwtKey, "jwt-key", "./bilive-jwt-key", "jwt ed25519 private key")
}

func main() {
	f.Parse(os.Args[1:])

	e := echo.New()

	frontend := try.To1(fs.Sub(frontendFiles, "build"))
	assertHandler := http.FileServer(http.FS(frontend))
	e.GET("*", echo.WrapHandler(assertHandler))

	var key = try.To1(os.ReadFile(args.jwtKey))
	privateKey, ok := try.To1(jwt.ParseEdPrivateKeyFromPEM(key)).(ed25519.PrivateKey)
	if !ok {
		panic(fmt.Errorf("jwt-key must be ed25519 private key"))
	}

	clientStore := store.NewClientStore()
	loadClients := func() {
		for _, c := range oc.Clients {
			clientStore.Set(c.ID, &models.Client{
				ID:     c.ID,
				Secret: c.Secret,
				Domain: c.Domain,
			})
		}
	}
	viper.OnConfigChange(func(in fsnotify.Event) {
		if err := viper.Unmarshal(&oc); err != nil {
			return
		}
		loadClients()
		log.Println("clients reloaded")
	})
	viper.WatchConfig()
	// 加载配置
	try.To(viper.ReadInConfig())
	try.To(viper.Unmarshal(&oc))
	loadClients()

	srv := initOAuth2Server(clientStore, key)
	registerOAuth2Server(e.Group("/oauth"), key, srv)

	var biliApp = oc.Bilibili
	bclient := bilibili.NewClient(biliApp.Key, biliApp.Secret)

	danmuCh := make(chan cmd.Danmu, 1024)
	{
		ctx := context.Background()
		getInfo := func() (_ bilibili.WebsocketInfo, err error) {
			defer err2.Handle(&err)
			app := try.To1(bclient.Open(ctx, biliApp.App, biliApp.Code))
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

	registerBiliveServer(e.Group("/bilive"), privateKey, biliApp.Room, danmuCh)
	registerBilibiliApi(e.Group("/bilibili"), privateKey, bclient, biliApp.App)

	log.Println(f.Name(), "start")
	try.To(e.Start(args.addr))
}
