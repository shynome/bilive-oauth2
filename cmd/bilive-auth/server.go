package main

import (
	"flag"

	"github.com/labstack/echo/v4"
	"github.com/lainio/err2/try"
	"github.com/rs/xid"
)

var args struct {
	addr   string
	pg     string
	room   int
	secret string
}

func init() {
	flag.StringVar(&args.addr, "addr", ":9096", "http server listen addr")
	flag.StringVar(&args.pg, "pg", "postgres://postgres:postgres@localhost:5432/postgres", "token file db")
	flag.IntVar(&args.room, "room", 898286, "room id")
	flag.StringVar(&args.secret, "secret", xid.New().String(), "cookie secret")
}

func main() {

	e := echo.New()

	srv := initOAuth2Server(args.pg)
	registerOAuth2Server(e.Group("/oauth"), srv)
	registerBiliveServer(e.Group("/bilive"), args.room)

	try.To(e.Start(args.addr))
}
