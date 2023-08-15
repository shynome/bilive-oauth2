package migrations

import (
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	migrations.AppMigrations.Register(func(db dbx.Builder) (err error) {
		defer err2.Handle(&err)

		dao := daos.New(db)

		func() {
			c := &models.Collection{
				Name: "oauth_client",
				Type: models.CollectionTypeBase,
				Indexes: types.JsonArray[string]{
					"domain",
				},
				Schema: schema.NewSchema(
					&schema.SchemaField{Name: "secret", Type: schema.FieldTypeText, Required: true},
					&schema.SchemaField{Name: "user_id", Type: schema.FieldTypeText, Required: true},
					&schema.SchemaField{Name: "domain", Type: schema.FieldTypeText},
					&schema.SchemaField{Name: "public", Type: schema.FieldTypeBool},
				),
			}
			c.MarkAsNew()
			try.To(dao.SaveCollection(c))
		}()

		func() {
			c := &models.Collection{
				Name: "oauth_token",
				Type: models.CollectionTypeBase,
				Indexes: types.JsonArray[string]{
					"code", "access", "refresh",
				},
				Schema: schema.NewSchema(
					&schema.SchemaField{Name: "code", Type: schema.FieldTypeText},
					&schema.SchemaField{Name: "access", Type: schema.FieldTypeText},
					&schema.SchemaField{Name: "refresh", Type: schema.FieldTypeText},
					&schema.SchemaField{Name: "info", Type: schema.FieldTypeJson},
				),
			}
			c.MarkAsNew()
			try.To(dao.SaveCollection(c))
		}()

		return
	}, func(db dbx.Builder) (err error) {
		defer err2.Handle(&err)
		tables := []string{
			"oauth_client",
			"oauth_token",
		}
		for _, t := range tables {
			try.To1(db.DropTable(t).Execute())
		}
		return
	})
}
