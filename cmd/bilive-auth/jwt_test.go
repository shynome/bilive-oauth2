package main

import (
	_ "embed"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/lainio/err2/try"
)

//go:embed bilive-jwt-key
var privkey []byte

func TestJWTSign(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{
		"sub": "00000",
	})
	key := try.To1(jwt.ParseEdPrivateKeyFromPEM(privkey))
	s := try.To1(token.SignedString(key))
	t.Log(s)
}
