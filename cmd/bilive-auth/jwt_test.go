package main

import (
	_ "embed"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/lainio/err2/try"
)

//go:embed bilive-jwt-key
var privkey []byte

func TestJWTSign(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.StandardClaims{
		Subject:   "root",
		ExpiresAt: time.Now().AddDate(0, 0, 1).Unix(),
	})
	key := try.To1(jwt.ParseEdPrivateKeyFromPEM(privkey))
	s := try.To1(token.SignedString(key))
	t.Log(s)
}
