package main

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
	bilibili "github.com/shynome/openapi-bilibili"
	"github.com/shynome/openapi-bilibili/live"
	"github.com/shynome/openapi-bilibili/live/cmd"
	"github.com/spf13/viper"
	"github.com/tidwall/buntdb"
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

	db := try.To1(buntdb.Open("uid.buntdb"))
	defer db.Close()
	go db.Shrink()

	srv := initOAuth2Server(clientStore, key)
	registerOAuth2Server(db, e.Group("/oauth"), key, srv)

	var biliApp = oc.Bilibili
	bclient := bilibili.NewClient(biliApp.Key, biliApp.Secret)

	danmuCh := make(chan cmd.Danmu, 1024)
	ctx := context.Background()
	ctx, exit := context.WithCancel(ctx)

	faces := try.To1(buntdb.Open(":memory:"))
	{
		linkDanmu := func(data []byte) {
			var err error
			defer err0.Then(&err, nil, func() {
				log.Println("parse danmu msg failed:", err)
			})
			var danmu cmd.Danmu
			try.To(json.Unmarshal(data, &danmu))
			var uid string
			db.View(func(tx *buntdb.Tx) (err error) {
				uid, err = tx.Get(danmu.OpenID)
				return err
			})
			if uid == "" {
				face := danmu.Uface
				err := faces.View(func(tx *buntdb.Tx) error {
					_, err := tx.Get(face)
					return err
				})
				if errors.Is(err, buntdb.ErrNotFound) {
					faces.Update(func(tx *buntdb.Tx) error {
						_, _, err := tx.Set(face, danmu.OpenID, &buntdb.SetOptions{
							Expires: true,
							TTL:     10 * time.Minute,
						})
						return err
					})
				}
			}
			danmuCh <- danmu
		}
		wctx, casue := context.WithCancelCause(ctx)
		go func() {
			connect := func(ctx context.Context) (err error) {
				defer err0.Then(&err, nil, nil)
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()
				app := try.To1(bclient.Open(ctx, biliApp.App, biliApp.Code))
				go app.KeepAlive(ctx)
				defer app.Close()
				info := app.Info().WebsocketInfo
				room := live.RoomWith(info)
				ch, err := room.Connect(ctx)
				casue(err)
				if err != nil {
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
				return nil
			}
			for {
				connect(ctx)
				time.Sleep(time.Second) // 等待1s后再重试, 避免重试过快导致占满资源
			}
		}()
		<-wctx.Done()
		if err := context.Cause(wctx); !errors.Is(err, context.Canceled) {
			try.To(err)
		}
	}

	if beer := os.Getenv("BEER"); beer != "" {
		go func() (err error) {
			connect := func(ctx context.Context) (err error) {
				defer err0.Then(&err, nil, nil)
				info := try.To1(getBeerConnectInfo(ctx, beer))
				room := live.RoomWith(info)
				ch := try.To1(room.Connect(ctx))
				log.Println("野生", "connected")
				for msg := range ch {
					if msg.Cmd != "DANMU_MSG" {
						continue
					}
					go func() (err error) {
						defer err0.Then(&err, nil, func() {
							slog.Error("解析野生弹幕出错", "err", err)
						})
						var list []json.RawMessage
						try.To(json.Unmarshal(msg.Info, &list))
						if len(list) == 0 {
							return
						}
						var list2 []json.RawMessage
						try.To(json.Unmarshal(list[0], &list2))
						if len(list2) < 16 {
							return
						}
						var user struct {
							YUser `json:"user"`
						}
						try.To(json.Unmarshal(list2[15], &user))

						return faces.View(func(tx *buntdb.Tx) error {
							openid, err := tx.Get(user.Face)
							if err != nil {
								return err
							}
							var uid string
							err = db.View(func(tx *buntdb.Tx) (err error) {
								uid, err = tx.Get(openid)
								return err
							})
							if uid != "" {
								return nil
							}
							return db.Update(func(tx *buntdb.Tx) error {
								uid := fmt.Sprintf("%d", user.UID)
								_, _, err := tx.Set(openid, uid, nil)
								slog.Info("link", "openid", openid, "uid", uid)
								return err
							})
						})
					}()
				}
				return nil
			}
			for {
				connect(ctx)
				time.Sleep(1 * time.Second) // 等待1s后再重试, 避免重试过快导致占满资源
			}
		}()
	}

	registerBiliveServer(e.Group("/bilive"), privateKey, biliApp.Room, danmuCh)
	registerBilibiliApi(e.Group("/bilibili"), privateKey, bclient, biliApp.App)

	quit := make(chan os.Signal)
	go func() {
		log.Println(f.Name(), "start")
		err := e.Start(args.addr)
		log.Println("server start failed", err)
		quit <- os.Interrupt
	}()

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	exit()

}

func getBeerConnectInfo(ctx context.Context, beer string) (info bilibili.WebsocketInfo, err error) {
	defer err0.Then(&err, nil, nil)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var uid, buvid string
	{
		link := fmt.Sprintf("http://%s/live-cookie", beer)
		req := try.To1(http.NewRequestWithContext(ctx, http.MethodGet, link, nil))
		resp := try.To1(http.DefaultClient.Do(req))
		defer resp.Body.Close()
		if code := resp.StatusCode; code != 200 {
			err := fmt.Errorf("resp status code expect 200, but got %d", code)
			try.To(err)
		}
		var data []string
		try.To(json.NewDecoder(resp.Body).Decode(&data))
		uid, buvid = data[0], data[1]
	}

	proxy := try.To1(url.Parse(fmt.Sprintf("http://%s:1080", beer)))
	hc := &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	var data Data
	{
		link := fmt.Sprintf("https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?id=%d&type=0", oc.Bilibili.Room)
		req := try.To1(http.NewRequestWithContext(ctx, http.MethodGet, link, nil))
		req.Header.Set("Js.fetch.credentials", "include")
		resp := try.To1(hc.Do(req))
		defer resp.Body.Close()
		if code := resp.StatusCode; code != 200 {
			err := fmt.Errorf("resp status code expect 200, but got %d", code)
			try.To(err)
		}
		var response bilibili.Response[Data]
		try.To(json.NewDecoder(resp.Body).Decode(&response))
		if response.Code != 0 {
			err := fmt.Errorf("code %d, err: %s", response.Code, response.Message)
			try.To(err)
		}
		data = response.Data
	}

	var auth = AuthBody{
		UID:      try.To1(strconv.Atoi(uid)),
		RoomID:   oc.Bilibili.Room,
		Ver:      2,
		BUVID:    buvid,
		Platform: "web",
		Type:     2,
		Token:    data.Token,
	}
	info.AuthBody = string(try.To1(json.Marshal(auth)))
	for _, h := range data.HostList {
		link := fmt.Sprintf("wss://%s:%d/sub", h.Host, h.WssPort)
		info.WssLink = append(info.WssLink, link)
	}
	return info, nil
}

type AuthBody struct {
	UID      int    `json:"uid"`
	RoomID   int    `json:"roomid"`
	Ver      int    `json:"protover"`
	BUVID    string `json:"buvid"`
	Platform string `json:"platform"`
	Type     int    `json:"type"`
	Token    string `json:"key"`
}

type Data struct {
	Token    string `json:"token"`
	HostList []Host `json:"host_list"`
}
type Host struct {
	Host    string `json:"host"`
	WssPort uint16 `json:"wss_port"`
}

// 野开用户结构
type YUser struct {
	YUserBase `json:"base"`

	UID int64 `json:"uid"`
}

type YUserBase struct {
	Face string `json:"face"`
}
