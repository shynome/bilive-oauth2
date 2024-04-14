package main

import (
	"context"
	"crypto"
	"fmt"
	"net/http"
	"os"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/go-session/session"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
	"github.com/tidwall/buntdb"
)

type OAuthClient struct {
	ID     string
	Secret string
	Domain string
}

func initOAuth2Server(clients oauth2.ClientStore, key []byte) *server.Server {

	manager := manage.NewDefaultManager()

	tokenStore := try.To1(store.NewMemoryTokenStore())
	manager.MapTokenStorage(tokenStore)

	manager.MapClientStorage(clients)

	manager.MapAccessGenerate(generates.NewJWTAccessGenerate("bilive-auth", key, jwt.SigningMethodEdDSA))

	srv := server.NewDefaultServer(manager)
	srv.SetUserAuthorizationHandler(func(w http.ResponseWriter, r *http.Request) (userID string, err error) {
		ctx := r.Context()
		user := ctx.Value(UIDContenxtKey)
		uid, ok := user.(string)
		if !ok {
			return "", echo.NewHTTPError(400, "uid not found")
		}
		return uid, nil
	})

	srv.SetClientInfoHandler(server.ClientFormHandler)

	return srv
}

type contextKey string

const UIDContenxtKey = contextKey("uid")

func registerOAuth2Server(db *buntdb.DB, e *echo.Group, key []byte, srv *server.Server) {

	var pubkey = func() crypto.PublicKey {
		key := try.To1(jwt.ParseEdPrivateKeyFromPEM(key))
		return key.(crypto.Signer).Public()
	}()

	e.Use(middleware.CORS())

	e.GET("/authorize", func(c echo.Context) (err error) {
		q := c.QueryString()
		return c.Redirect(302, "/?"+q)
	})
	e.POST("/authorize", func(c echo.Context) (err error) {
		w, r := c.Response(), c.Request()
		token := c.FormValue("bilive-token")
		if token == "" {
			return echo.NewHTTPError(400, "token is required")
		}
		claims := new(jwt.StandardClaims)
		try.To1(jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
			return pubkey, nil
		}))
		ctx := r.Context()
		ctx = context.WithValue(ctx, UIDContenxtKey, claims.Subject)
		r = r.WithContext(ctx)
		return srv.HandleAuthorizeRequest(w, r)
	})
	e.Any("/token", func(c echo.Context) (err error) {
		w, r := c.Response(), c.Request()
		return srv.HandleTokenRequest(w, r)
	})
	e.Any("/allow", func(c echo.Context) (err error) {
		defer err0.Then(&err, nil, nil)
		w, r := c.Response(), c.Request()
		store := try.To1(session.Start(r.Context(), w, r))
		uid, ok := store.Get("uid")
		if !ok {
			return c.Redirect(302, "/")
		}
		store.Set("l-uid", uid.(string))
		store.Save()
		return c.Redirect(302, "/oauth/authorize")
	})
	beer := os.Getenv("BEER")
	e.Any("/whoami", func(c echo.Context) (err error) {
		defer err0.Then(&err, nil, nil)
		r := c.Request()
		token := try.To1(srv.ValidationBearerToken(r))
		openid := token.GetUserID()
		var uid string
		if beer != "" {
			db.View(func(tx *buntdb.Tx) (err error) {
				uid, err = tx.Get(openid)
				return err
			})
		}
		info := UserInfo{
			OldUserCheck: OldUserCheck{ClientID: token.GetClientID(), UserID: uid},

			Id:       openid,
			Name:     uid,
			Username: openid,
		}
		if uid != "" {
			info.Email = fmt.Sprintf("%s@bilibili.com", uid)
			info.EmailVerified = true
		}
		return c.JSON(200, info)
	})
}

type UserInfo struct {
	OldUserCheck
	Id            string `json:"sub"`
	Name          string `json:"name"`
	Username      string `json:"preferred_username"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

// 遗留的兼容前端代码
type OldUserCheck struct {
	ClientID string `json:"client_id"`
	UserID   string `json:"user_id"`
}
