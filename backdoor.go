package main

import (
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pocketbase/pocketbase/core"
	"github.com/shynome/bilive-oauth2/v2/db"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
)

func initBackdoor(app core.App) {
	app.OnRecordCreateRequest(db.TableBackdoor).BindFunc(func(e *core.RecordRequestEvent) (err error) {
		defer err0.Then(&err, nil, nil)
		r := e.Record
		linked := try.To1(e.App.FindRecordById(db.TableLinkeds, r.GetString("linked")))
		expired := time.Now().AddDate(0, 0, 1)
		r.Set("expired", expired)
		claims := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.RegisteredClaims{
			Subject:   linked.GetString("openid"),
			ExpiresAt: jwt.NewNumericDate(expired),
		})
		token := try.To1(claims.SignedString(privKey))
		u := try.To1(url.Parse(e.App.Settings().Meta.AppURL))
		u.Path += "/backdoor/reset-token/"
		q := u.Query()
		q.Set("token", token)
		u.RawQuery = q.Encode()
		b := u.String()
		r.Set("backdoor", b)
		return e.Next()
	})
}
