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
			&core.TextField{
				Name: "bname", Id: ID("bname"), System: true,
			},
			&core.BoolField{
				Name: "verify_sign", Id: ID("verify_sign"), System: true, // 是否验证签名
			},
			&core.JSONField{
				Name: "bconfig", Id: ID("bconfig"), System: true,
				Hidden: true,
			},
		)
		try.To(app.Save(clients))

		return nil
	}, func(app core.App) error {
		return fmt.Errorf("init no rollback")
	})
}
