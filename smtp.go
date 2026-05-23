package main

import (
	"os"

	"github.com/pocketbase/pocketbase/core"
	"github.com/shynome/bilireq"
)

func initSMTP(app core.App) {
	bc := bilireq.New(os.Getenv("BEER"))
	_ = bc
}
