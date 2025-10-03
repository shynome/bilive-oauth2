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

		linkeds := try.To1(app.FindCollectionByNameOrId(db.TableLinkeds))
		linkeds.Fields.AddAt(getFieldNext(linkeds, "uid"),
			&core.TextField{
				Name: "room", Id: ID("room"), System: true,
			},
			&core.TextField{
				Name: "id_code", Id: ID("id_code"), System: true,
				Hidden: true,
			},
		)
		addUpdatedFields(&linkeds.Fields)
		try.To(app.Save(linkeds))

		return nil
	}, func(app core.App) error {
		return fmt.Errorf("init no rollback")
	})
}
