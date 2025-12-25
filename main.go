package main

import (
	"github.com/pocketbase/pocketbase"
	_ "github.com/shynome/bilive-oauth2/v2/db/migrations"
	"github.com/shynome/err0/try"
)

var args struct {
	jwtKey []byte

	Code string // 身份码
	Room int    // 直播间, 虽然身份码可以拿到直播间号, 但还是直接写一下吧
}

type BConfig struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
	App    int64  `json:"app"`
}

var bconfig BConfig

var Version = "dev"

func main() {
	app := pocketbase.New()
	app.RootCmd.Version = Version

	{
		flags := app.RootCmd.PersistentFlags()
		flags.BytesBase64Var(&args.jwtKey, "jwt-key", nil, "jwt ed25519 private key seed")

		flags.StringVar(&bconfig.Key, "key", "", "bilibili App Key")
		flags.StringVar(&bconfig.Secret, "secret", "", "bilibili App Secret")
		flags.Int64Var(&bconfig.App, "app", 0, "bilibili App ID")
		flags.StringVar(&args.Code, "code", "", "bilibili Room IDCode")
		flags.IntVar(&args.Room, "room", 0, "bilibili Room number")
	}
	app.OnServe().BindFunc(initBilibili) // 必须先初始化此项, 才有 bclient
	app.OnServe().BindFunc(initBiliveServer)
	app.OnServe().BindFunc(initOAuth2)

	try.To(app.Start())
}
