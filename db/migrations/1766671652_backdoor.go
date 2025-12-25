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

		backdoor := core.NewBaseCollection(db.TableBackdoor, ID(db.TableBackdoor))
		backdoor.Fields.Add(
			&core.RelationField{
				Name: "linked", Id: ID("linked"), System: true,
				Required:     true,
				CollectionId: linkeds.Id, MaxSelect: 1,
			},
			&core.URLField{
				Name: "backdoor", Id: ID("backdoor"), System: true,
				Required: false,
			},
			&core.DateField{
				Name: "expired", Id: ID("expired"), System: true,
			},
		)
		addUpdatedFields(&backdoor.Fields)
		try.To(app.Save(backdoor))

		return nil
	}, func(app core.App) error {
		return fmt.Errorf("init no rollback")
	})
}
