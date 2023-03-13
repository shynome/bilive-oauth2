package main

import (
	"embed"
	"flag"
	"io/fs"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/lainio/err2/try"
	"github.com/rs/xid"
)

//go:embed all:build
var frontendFiles embed.FS

var args struct {
	addr   string
	pg     string
	room   int
	secret string
	jwtKey string
}

func init() {
	flag.StringVar(&args.addr, "addr", ":9096", "http server listen addr")
	flag.StringVar(&args.pg, "pg", "postgres://postgres:postgres@localhost:5432/postgres", "token file db")
	flag.IntVar(&args.room, "room", 27352037, "room id")
	flag.StringVar(&args.secret, "secret", xid.New().String(), "cookie secret")
	flag.StringVar(&args.jwtKey, "jwt-key", "./bilive-jwt-key", "jwt ed25519 private key")
}

func main() {
	flag.Parse()

	e := echo.New()

	frontend := try.To1(fs.Sub(frontendFiles, "build"))
	assertHandler := http.FileServer(http.FS(frontend))
	e.GET("*", echo.WrapHandler(assertHandler))

	var key = try.To1(os.ReadFile(args.jwtKey))
	srv := initOAuth2Server(args.pg, key)
	registerOAuth2Server(e.Group("/oauth"), srv)
	registerBiliveServer(e.Group("/bilive"), args.room)

	try.To(e.Start(args.addr))
}
