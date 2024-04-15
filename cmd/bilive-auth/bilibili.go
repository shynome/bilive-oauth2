package main

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
	bilibili "github.com/shynome/openapi-bilibili"
	"github.com/shynome/openapi-bilibili/live"
	"github.com/shynome/openapi-bilibili/live/cmd"
	"github.com/tidwall/buntdb"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func registerBilibiliApi(db *buntdb.DB, e *echo.Group, privateKey ed25519.PrivateKey, bclient *bilibili.Client, appid int64) {

	pubkey := privateKey.Public()
	e.Use(echojwt.WithConfig(echojwt.Config{
		TokenLookup: "header:Authorization:Bearer ,form:token",
		ParseTokenFunc: func(c echo.Context, auth string) (interface{}, error) {
			token, err := jwt.ParseWithClaims(
				auth, new(jwt.StandardClaims),
				func(t *jwt.Token) (interface{}, error) { return pubkey, nil },
			)
			if err != nil {
				return nil, err
			}
			claims := token.Claims.(*jwt.StandardClaims)
			if ok := claims.VerifyAudience("https://open-live.bilibili.com", true); !ok {
				return nil, fmt.Errorf("用途错误")
			}
			if claims.Subject != "root" {
				return nil, fmt.Errorf("该接口只允许 root 用户访问")
			}
			return token, nil
		},
	}))

	e.Any("/ws-info-keep", func(c echo.Context) (err error) {
		defer err0.Then(&err, nil, nil)
		IDCode := c.QueryParam("IDCode")
		if IDCode == "" {
			return echo.NewHTTPError(400, "require query param: IDCode")
		}
		r, w := c.Request(), c.Response()
		ctx := r.Context()
		app := try.To1(bclient.Open(ctx, appid, IDCode))
		defer app.Close()
		conn := try.To1(websocket.Accept(w, r, nil))
		defer conn.Close(websocket.StatusAbnormalClosure, "defer manual close")
		go func() {
			var closeMsg = "manual close"
			if err := app.KeepAlive(ctx); err != nil {
				// do nothing
				closeMsg = err.Error()
			}
			conn.Close(websocket.StatusAbnormalClosure, closeMsg)
		}()
		info := app.Info().WebsocketInfo
		try.To(wsjson.Write(ctx, conn, info))
		if beer != "" {
			go link(ctx, db, info)
		}
		for {
			var linked LinkedOpenID
			if err := wsjson.Read(ctx, conn, &linked); err != nil {
				return err
			}
			go func() {
				var uid string
				db.View(func(tx *buntdb.Tx) (err error) {
					uid, err = tx.Get(linked.OpenID)
					return err
				})
				if uid == "" {
					return
				}
				linked.UID = uid
				wsjson.Write(ctx, conn, linked)
			}()
		}
	})
}

var beer = os.Getenv("BEER")
var WSOpts *websocket.DialOptions
var faces = try.To1(buntdb.Open(":memory:"))

func init() {
	if proxy := os.Getenv("BEER_PROXY"); proxy != "" {
		proxy := try.To1(url.Parse(proxy))
		client := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxy),
			},
		}
		WSOpts = &websocket.DialOptions{
			HTTPClient: client,
		}
	}
}

func link(ctx context.Context, db *buntdb.DB, info bilibili.WebsocketInfo) (err error) {
	logger := slog.With()
	defer err0.Then(&err, nil, nil)
	var roomid int
	{
		var authBody AuthBody
		try.To(json.Unmarshal([]byte(info.AuthBody), &authBody))
		roomid = authBody.RoomID
		logger = logger.With("room", roomid)

		room := live.RoomWith(info)
		connect := func() (err error) {
			defer err0.Then(&err, nil, nil)
			ch := try.To1(room.Connect(ctx))
			logger.Info("开平连接成功")
			for msg := range ch {
				go func() {
					switch msg.Cmd {
					case cmd.CmdDanmu:
						var danmu cmd.Danmu
						try.To(json.Unmarshal(msg.Data, &danmu))
						setFaceOpenID(db, danmu.Uface, danmu.OpenID)
					case cmd.CmdGift:
						var gift cmd.Gift
						try.To(json.Unmarshal(msg.Data, &gift))
						setFaceOpenID(db, gift.Uface, gift.OpenID)
					}
				}()
			}
			return nil
		}
		go func() {
			for {
				if err := connect(); errors.Is(err, context.Canceled) {
					return
				} else if err != nil {
					logger.Info("开平连接出错", "err", err)
				}
				time.Sleep(time.Second)
			}
		}()
	}

	connect := func() (err error) {
		defer err0.Then(&err, nil, nil)
		info := try.To1(getBeerConnectInfo(ctx, roomid))
		room := live.RoomWith(info)
		room.WSDialOptioins = WSOpts
		ch := try.To1(room.Connect(ctx))
		logger.Info("野开连接成功")
		for msg := range ch {
			go func() (err error) {
				defer err0.Then(&err, nil, func() {
					slog.Error("解析野生弹幕出错", "err", err)
				})
				var linked *LinkedOpenID
				switch msg.Cmd {
				case "DANMU_MSG":
					user := try.To1(getYDanmuUser(msg))
					linked = try.To1(linkOpenID(db, user.Face, user.UID))
				case "SEND_GIFT":
					var gift YGift
					try.To(json.Unmarshal(msg.Data, &gift))
					linked = try.To1(linkOpenID(db, gift.Face, gift.UID))
				}
				if linked != nil {
					// do nothing
				}
				return nil
			}()
		}
		return nil
	}

	for {
		if err := connect(); errors.Is(err, context.Canceled) {
			return nil
		} else if err != nil {
			logger.Error("野开连接出错", "err", err)
		}
		time.Sleep(time.Second)
	}
}

func setFaceOpenID(db *buntdb.DB, face string, openid string) (err error) {
	var uid string
	db.View(func(tx *buntdb.Tx) (err error) {
		uid, err = tx.Get(openid)
		return err
	})
	if uid != "" {
		return
	}
	return faces.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(face, openid, &buntdb.SetOptions{
			Expires: true,
			TTL:     10 * time.Minute,
		})
		return err
	})
}

func linkOpenID(db *buntdb.DB, face string, uidNew int64) (linked *LinkedOpenID, err error) {
	defer err0.Then(&err, nil, nil)
	var openid string
	err = faces.View(func(tx *buntdb.Tx) (err error) {
		openid, err = tx.Get(face)
		return err
	})
	// 如果 uid 已设置, 就不会在 faces 中设置 openid
	if errors.Is(err, buntdb.ErrNotFound) {
		return nil, nil
	}
	try.To(err)

	var uid string
	db.View(func(tx *buntdb.Tx) (err error) {
		uid, err = tx.Get(openid)
		return err
	})

	// 如果 openid 已经绑定了 uid 跳过. (因为写锁是单发的)
	if uid != "" {
		return nil, nil
	}

	uid = fmt.Sprintf("%d", uidNew)
	err = db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(openid, uid, nil)
		slog.Info("link", "openid", openid, "uid", uid)
		return err
	})
	try.To(err)

	return &LinkedOpenID{OpenID: openid, UID: uid}, nil
}

type LinkedOpenID struct {
	roomIDCode string

	OpenID string `json:"openid"`
	UID    string `json:"uid"`
}

func getBeerConnectInfo(ctx context.Context, room int) (info bilibili.WebsocketInfo, err error) {
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
		link := fmt.Sprintf("https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?id=%d&type=0", room)
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
		RoomID:   room,
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

func getYDanmuUser(msg cmd.Cmd[json.RawMessage]) (_ *YUser, err error) {
	defer err0.Then(&err, nil, nil)
	var list []json.RawMessage
	try.To(json.Unmarshal(msg.Info, &list))
	if len(list) == 0 {
		return nil, fmt.Errorf("弹幕数据格式有误.1")
	}
	var list2 []json.RawMessage
	try.To(json.Unmarshal(list[0], &list2))
	if len(list2) < 16 {
		return nil, fmt.Errorf("弹幕数据格式有误.2")
	}
	var user struct {
		YUser `json:"user"`
	}
	try.To(json.Unmarshal(list2[15], &user))
	return &user.YUser, nil
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

type YGift struct {
	Face string `json:"face"`
	UID  int64  `json:"uid"`
}
