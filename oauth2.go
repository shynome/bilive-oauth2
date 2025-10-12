package main

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-session/session"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
	"github.com/shynome/bilive-oauth2/v2/db"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
	openbili "github.com/shynome/openapi-bilibili"
)

func initOAuth2(se *core.ServeEvent) (err error) {

	key := ed25519.NewKeyFromSeed(args.jwtKey)
	pubkey := key.Public()

	manager := manage.NewDefaultManager()

	tokenStore := &TokenStore{se.App}
	manager.MapTokenStorage(tokenStore)

	// 每小时清理下过期 token
	se.App.Cron().Add("clear-expired-tokens", "0 * * * *", func() {
		logger := se.App.Logger()
		q := dbx.NewExp("expired_at < {:now}", dbx.Params{"now": types.NowDateTime()})
		if _, err := se.App.DB().Delete(db.TableTokens, q).Execute(); err != nil {
			logger.Error("清理过期 token 出错了", "error", err)
		}
	})

	var clients oauth2.ClientStore = &ClientStore{se.App}
	manager.MapClientStorage(clients)

	manager.MapAccessGenerate(NewJWTAccessGenerate("bilive-auth", key, jwt.SigningMethodEdDSA))

	srv := server.NewDefaultServer(manager)
	srv.SetUserAuthorizationHandler(func(w http.ResponseWriter, r *http.Request) (userID string, err error) {
		ctx := r.Context()
		user := ctx.Value(UIDContenxtKey)
		uid, ok := user.(string)
		if !ok {
			return "", apis.NewBadRequestError("uid not found", nil)
		}
		return uid, nil
	})

	srv.SetClientInfoHandler(server.ClientFormHandler)

	eg := se.Router.Group("/oauth")

	eg.GET("/authorize", func(e *core.RequestEvent) error {
		q := e.Request.URL.Query()
		return e.Redirect(302, "/?"+q.Encode())
	})
	eg.POST("/authorize", func(e *core.RequestEvent) (err error) {
		defer err0.Then(&err, nil, nil)
		w, r := e.Response, e.Request
		token := r.FormValue("bilive-token")
		if token == "" {
			return apis.NewBadRequestError("token is required", nil)
		}
		claims := new(jwt.MapClaims)
		try.To1(jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
			return pubkey, nil
		}))
		ctx := r.Context()
		sub := try.To1(claims.GetSubject())
		ctx = context.WithValue(ctx, UIDContenxtKey, sub)
		r = r.WithContext(ctx)
		return srv.HandleAuthorizeRequest(w, r)
	})
	eg.POST("/id_code", func(e *core.RequestEvent) (err error) {
		defer err0.Then(&err, nil, nil)
		now := time.Now()
		w, r := e.Response, e.Request
		code := r.FormValue("code")
		if code == "" {
			return apis.NewBadRequestError("Code is required", nil)
		}

		ctx := r.Context()
		game, err := bclient.Open(ctx, args.App, code)
		if err != nil {
			var he *openbili.Response[json.RawMessage]
			if errors.As(err, &he) {
				msg := he.Error()
				return apis.NewBadRequestError(msg, err)
			}
			return err
		}
		game.Close() // 关闭它, 此时已经拿到主播信息了

		anchor := game.Info().AnchorInfo
		linkeds := try.To1(e.App.FindCachedCollectionByNameOrId(db.TableLinkeds))
		err = e.App.RunInTransaction(func(tx core.App) error {
			linked, err := tx.FindFirstRecordByData(db.TableLinkeds, "openid", anchor.OpenID)
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					return err
				}
				linked = core.NewRecord(linkeds)
				linked.Set("openid", anchor.OpenID)
			}
			// 更新信息
			linked.Set("room", anchor.RoomID)
			linked.Set("id_code", code)
			linked.Set("uid", fmt.Sprintf("%d", anchor.UID))
			linked.Set("uname", anchor.Username)
			return tx.Save(linked)
		})
		try.To(err)

		// 需要验证 Timestamp, 确认是否为用户本人操作的. 因为 Code 可能会被其他应用存储(是的,我也存了), 并不能代表是用户本人在操作
		if !e.App.IsDev() { // 在开发环境下不验证
			u := try.To1(url.Parse(r.FormValue("redirect_uri")))
			q := u.Query()
			if err := bclient.VerifyH5Params(q); err != nil {
				return apis.NewBadRequestError("参数验证失败", err)
			}
			tsInt := try.To1(strconv.ParseInt(q.Get("Timestamp"), 10, 64))
			t := time.Unix(tsInt, 0)
			if now.Sub(t) > 20*time.Second {
				return apis.NewBadRequestError("timestamp 已过期", nil)
			}
		}

		ctx = context.WithValue(ctx, UIDContenxtKey, anchor.OpenID)
		r = r.WithContext(ctx)
		if err := srv.HandleAuthorizeRequest(w, r); err != nil {
			return apis.NewBadRequestError("授权处理失败", err)
		}
		return nil
	})
	eg.Any("/token", func(e *core.RequestEvent) error {
		w, r := e.Response, e.Request
		err := srv.HandleTokenRequest(w, r)
		return err
	})
	eg.Any("/allow", func(e *core.RequestEvent) (err error) {
		defer err0.Then(&err, nil, nil)
		w, r := e.Response, e.Request
		ctx := r.Context()
		store := try.To1(session.Start(ctx, w, r))
		uid, ok := store.Get("uid")
		if !ok {
			return e.Redirect(302, "/")
		}
		store.Set("l-uid", uid.(string))
		try.To(store.Save())
		return e.Redirect(302, "/oauth/authorize")
	})
	eg.Any("/whoami", func(e *core.RequestEvent) (err error) {
		defer err0.Then(&err, nil, nil)
		r := e.Request
		toekn := try.To1(srv.ValidationBearerToken(r))
		openid := toekn.GetUserID()
		var uid, id_code, room string
		if record, err := e.App.FindFirstRecordByData(db.TableLinkeds, "openid", openid); err == nil {
			uid = record.GetString("uid")
			id_code = record.GetString("id_code")
			room = record.GetString("room")
		}
		info := UserInfo{
			OldUserCheck: OldUserCheck{ClientID: toekn.GetClientID(), UserID: uid},

			Id:       openid,
			Name:     uid,
			Username: openid,
			IDCode:   id_code,
			Room:     room,
		}
		if uid != "" {
			info.Email = fmt.Sprintf("%s@bilibili.com", uid)
			info.EmailVerified = true
		}
		return e.JSON(200, info)
	})

	return se.Next()
}

type contextKey string

const UIDContenxtKey = contextKey("uid")

type UserInfo struct {
	OldUserCheck
	Id            string `json:"sub"`
	Name          string `json:"name"`
	Username      string `json:"preferred_username"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	IDCode        string `json:"id_code"`
	Room          string `json:"room"`
}

// 遗留的兼容前端代码
type OldUserCheck struct {
	ClientID string `json:"client_id"`
	UserID   string `json:"user_id"`
}

type ClientStore struct{ core.App }

var _ oauth2.ClientStore = (*ClientStore)(nil)

func (app *ClientStore) GetByID(ctx context.Context, id string) (oauth2.ClientInfo, error) {
	record, err := app.FindFirstRecordByData(db.TableClients, "application", id)
	if err != nil {
		return nil, err
	}
	c := &db.Client{}
	c.SetProxyRecord(record)
	return c, nil
}

type TokenStore struct{ core.App }

var _ oauth2.TokenStore = (*TokenStore)(nil)

// create and store the new token information
func (app *TokenStore) Create(ctx context.Context, info oauth2.TokenInfo) (err error) {
	defer err0.Then(&err, nil, nil)
	var jv types.JSONRaw = try.To1(json.Marshal(info))

	tokens := try.To1(app.FindCachedCollectionByNameOrId(db.TableTokens))
	record := core.NewRecord(tokens)
	record.Set("data", jv)

	if code := info.GetCode(); code != "" {
		expiredAt := info.GetCodeCreateAt().Add(info.GetCodeExpiresIn())
		record.Set("code", code)
		record.Set("expired_at", try.To1(types.ParseDateTime(expiredAt)))
	} else {
		access := info.GetAccess()
		expiredAt := info.GetAccessCreateAt().Add(info.GetAccessExpiresIn())
		record.Set("access", access)
		record.Set("expired_at", try.To1(types.ParseDateTime(expiredAt)))
		if refresh := info.GetRefresh(); refresh != "" {
			expiredAt := info.GetRefreshCreateAt().Add(info.GetRefreshExpiresIn())
			record.Set("refresh", refresh)
			record.Set("expired_at", expiredAt)
		}
	}

	try.To(app.Save(record))

	return nil
}

func (app *TokenStore) RemoveByCode(ctx context.Context, code string) error {
	return app.removeByData("code", code)
}
func (app *TokenStore) RemoveByAccess(ctx context.Context, access string) error {
	return app.removeByData("access", access)
}
func (app *TokenStore) RemoveByRefresh(ctx context.Context, refresh string) error {
	return app.removeByData("refresh", refresh)
}

func (app *TokenStore) GetByCode(ctx context.Context, code string) (_ oauth2.TokenInfo, err error) {
	return app.getByData("code", code)
}
func (app *TokenStore) GetByAccess(ctx context.Context, access string) (oauth2.TokenInfo, error) {
	return app.getByData("access", access)
}
func (app *TokenStore) GetByRefresh(ctx context.Context, refresh string) (oauth2.TokenInfo, error) {
	return app.getByData("refresh", refresh)
}

func (app *TokenStore) removeByData(key, data string) (err error) {
	defer err0.Then(&err, nil, nil)
	tokens := try.To1(app.FindCachedCollectionByNameOrId(db.TableTokens))
	record, err := app.FindFirstRecordByData(tokens, key, data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	try.To(app.Delete(record))
	return nil
}

func (app *TokenStore) getByData(key, data string) (_ oauth2.TokenInfo, err error) {
	defer err0.Then(&err, nil, nil)
	tokens := try.To1(app.FindCachedCollectionByNameOrId(db.TableTokens))
	record := try.To1(app.FindFirstRecordByData(tokens, key, data))
	m := record.GetString("data")
	var tm models.Token
	try.To(json.Unmarshal([]byte(m), &tm))
	return &tm, nil
}

func NewJWTAccessGenerate(kid string, key []byte, method jwt.SigningMethod) *JWTAccessGenerate {
	return &JWTAccessGenerate{
		SignedKeyID:  kid,
		SignedKey:    key,
		SignedMethod: method,
	}
}

type JWTAccessGenerate struct {
	SignedKeyID  string
	SignedKey    ed25519.PrivateKey
	SignedMethod jwt.SigningMethod
}

func (a *JWTAccessGenerate) Token(ctx context.Context, data *oauth2.GenerateBasic, isGenRefresh bool) (string, string, error) {
	claims := &generates.JWTAccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{data.Client.GetID()},
			Subject:   data.UserID,
			ExpiresAt: jwt.NewNumericDate(data.TokenInfo.GetAccessCreateAt().Add(data.TokenInfo.GetAccessExpiresIn())),
		},
	}

	token := jwt.NewWithClaims(a.SignedMethod, claims)
	if a.SignedKeyID != "" {
		token.Header["kid"] = a.SignedKeyID
	}

	access, err := token.SignedString(a.SignedKey)
	if err != nil {
		return "", "", err
	}
	refresh := ""

	if isGenRefresh {
		t := uuid.NewSHA1(uuid.Must(uuid.NewRandom()), []byte(access)).String()
		refresh = base64.URLEncoding.EncodeToString([]byte(t))
		refresh = strings.ToUpper(strings.TrimRight(refresh, "="))
	}

	return access, refresh, nil
}
