package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/migrations"
	"github.com/shynome/bilive-oauth2/v2/db"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
)

func init() {
	migrations.Register(func(app core.App) (err error) {
		defer err0.Then(&err, nil, nil)

		clients := try.To1(app.FindCollectionByNameOrId(db.TableClients))
		clients.Fields.GetByName("secret").(*core.TextField).Hidden = true // 隐藏 secret
		clients.Fields.AddAt(getFieldNext(clients, "secret"),
			&core.NumberField{
				Name: "room", Id: ID("room"), System: true,
				Hidden:  true,
				OnlyInt: true,
			},
		)
		try.To(app.Save(clients))

		return nil
	}, func(app core.App) error {
		return fmt.Errorf("init no rollback")
	})
}
