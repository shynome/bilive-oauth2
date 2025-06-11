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

		users := try.To1(app.FindCollectionByNameOrId("users"))
		users.CreateRule = nil
		try.To(app.Save(users))

		linkeds := core.NewBaseCollection(db.TableLinkeds, ID(db.TableLinkeds))
		linkeds.Fields.Add(
			&core.TextField{
				Name: "openid", Id: ID("openid"), System: true,
				Required: true,
			},
			&core.TextField{
				Name: "uname", Id: ID("uname"), System: true,
				Presentable: true,
			},
			&core.TextField{
				Name: "uid", Id: ID("uid"), System: true,
				Required: true, Presentable: true,
			},
		)
		addUpdatedFields(&linkeds.Fields)
		try.To(app.Save(linkeds))

		vids := core.NewBaseCollection(db.TableTmpVIDs, ID(db.TableTmpVIDs))
		addUpdatedFields(&vids.Fields)
		try.To(app.Save(vids))

		return nil
	}, func(app core.App) error {
		return fmt.Errorf("init no rollback")
	})
}
