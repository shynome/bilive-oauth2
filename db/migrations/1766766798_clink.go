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

		clinks := core.NewBaseCollection(db.TableCLinks, ID(db.TableCLinks))
		clinks.Fields.Add(
			&core.NumberField{
				Name: "app", Id: ID("app"), System: true,
				Required: true,
				OnlyInt:  true,
			},
			&core.TextField{
				Name: "openid", Id: ID("openid"), System: true,
				Required: true,
			},
			&core.RelationField{
				Name: "linked", Id: ID("linked"), System: true,
				Required:     true,
				CollectionId: linkeds.Id, MaxSelect: 1,
			},
		)
		addUpdatedFields(&clinks.Fields)
		try.To(app.Save(clinks))

		return nil
	}, func(app core.App) error {
		return fmt.Errorf("init no rollback")
	})
}
