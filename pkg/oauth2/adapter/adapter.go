package adapter

import "github.com/pocketbase/pocketbase"

type Adapter struct {
	*pocketbase.PocketBase
}

func New(app *pocketbase.PocketBase) *Adapter {
	return &Adapter{
		PocketBase: app,
	}
}
