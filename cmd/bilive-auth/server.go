package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
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

	bilipage string //用于获取b站用户信息
}

var Version = "dev"
var f = flag.NewFlagSet("bilive-oauth2@"+Version, flag.ExitOnError)

func init() {
	f.StringVar(&args.addr, "addr", ":9096", "http server listen addr")
	f.StringVar(&args.pg, "pg", "postgres://postgres:postgres@localhost:5432/postgres", "token file db")
	f.IntVar(&args.room, "room", 27352037, "room id")
	f.StringVar(&args.secret, "secret", xid.New().String(), "cookie secret")
	f.StringVar(&args.jwtKey, "jwt-key", "./bilive-jwt-key", "jwt ed25519 private key")
	f.StringVar(&args.bilipage, "bilipage", "", "用于获取b站用户信息")
}

func main() {
	f.Parse(os.Args[1:])

	e := echo.New()

	frontend := try.To1(fs.Sub(frontendFiles, "build"))
	assertHandler := http.FileServer(http.FS(frontend))
	e.GET("*", echo.WrapHandler(assertHandler))

	var key = try.To1(os.ReadFile(args.jwtKey))
	srv := initOAuth2Server(args.pg, key)
	registerOAuth2Server(e.Group("/oauth"), srv)
	registerBiliveServer(e.Group("/bilive"), args.room, args.bilipage)

	log.Println(f.Name(), "start")
	try.To(e.Start(args.addr))
}
