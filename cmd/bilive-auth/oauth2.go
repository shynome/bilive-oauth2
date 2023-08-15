package main

import (
	"net/http"
	"net/url"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-session/session"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/pocketbase/pocketbase"
)

func initOAuth2Server(app *pocketbase.PocketBase, key []byte) *server.Server {

	manager := manage.NewDefaultManager()

	var tokenStore oauth2.TokenStore
	manager.MapTokenStorage(tokenStore)

	var clientStore oauth2.ClientStore
	manager.MapClientStorage(clientStore)

	manager.MapAccessGenerate(generates.NewJWTAccessGenerate("bilive-auth", key, jwt.SigningMethodEdDSA))

	srv := server.NewDefaultServer(manager)
	srv.SetUserAuthorizationHandler(func(w http.ResponseWriter, r *http.Request) (userID string, err error) {
		defer err2.Handle(&err)
		store := try.To1(session.Start(r.Context(), w, r))

		uid, ok := store.Get("l-uid")
		if !ok {
			if r.Form == nil {
				try.To(r.ParseForm())
			}
			store.Set("ReturnUri", r.Form)
			try.To(store.Save())
			http.Redirect(w, r, "/", 302)
			return
		}

		userID = uid.(string)
		store.Delete("l-uid")
		try.To(store.Save())
		return
	})

	srv.SetClientInfoHandler(server.ClientFormHandler)

	return srv
}

func registerOAuth2Server(e *echo.Group, srv *server.Server) {

	e.Use(middleware.CORS())

	e.Any("/authorize", func(c echo.Context) (err error) {
		defer err2.Handle(&err)

		w, r := c.Response(), c.Request()
		store := try.To1(session.Start(r.Context(), w, r))
		if form, ok := store.Get("ReturnUri"); ok {
			r.Form = form.(url.Values)
		}

		store.Delete("ReturnUri")
		try.To(store.Save())

		try.To(srv.HandleAuthorizeRequest(w, r))
		return
	})
	e.Any("/token", func(c echo.Context) (err error) {
		defer err2.Handle(&err)
		w, r := c.Response(), c.Request()
		try.To(srv.HandleTokenRequest(w, r))
		return
	})
	e.Any("/allow", func(c echo.Context) (err error) {
		defer err2.Handle(&err)
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
	e.Any("/whoami", func(c echo.Context) (err error) {
		defer err2.Handle(&err)
		r := c.Request()
		token := try.To1(srv.ValidationBearerToken(r))
		return c.JSON(200, map[string]any{
			"client_id": token.GetClientID(),
			"user_id":   token.GetUserID(),
		})
	})
}
