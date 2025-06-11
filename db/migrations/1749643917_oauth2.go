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

		clients := core.NewBaseCollection(db.TableClients, ID(db.TableClients))
		clients.Fields.Add(
			&core.TextField{
				Name: "application", Id: ID("application"), System: true,
				Required: true, Presentable: true,
			},
			&core.TextField{
				Name: "secret", Id: ID("secret"), System: true,
				Required: true,
			},
			&core.URLField{
				Name: "domain", Id: ID("domain"), System: true,
			},
			&core.BoolField{
				Name: "public", Id: ID("public"), System: true,
			},
		)
		addUpdatedFields(&clients.Fields)
		try.To(app.Save(clients))

		tokens := core.NewBaseCollection(db.TableTokens, ID(db.TableTokens))
		tokens.Fields.Add(
			&core.TextField{
				Name: "code", Id: ID("code"), System: true,
			},
			&core.TextField{
				Name: "access", Id: ID("access"), System: true,
			},
			&core.TextField{
				Name: "refresh", Id: ID("refresh"), System: true,
			},
			&core.DateField{
				Name: "expired_at", Id: ID("expired_at"), System: true,
			},
			&core.JSONField{
				Name: "data", Id: ID("data"), System: true,
				Required: true,
			},
		)
		addUpdatedFields(&tokens.Fields)
		tokens.AddIndex("code", false, "code", "")
		tokens.AddIndex("access", false, "access", "")
		tokens.AddIndex("refresh", false, "refresh", "")
		try.To(app.Save(tokens))

		return nil
	}, func(app core.App) error {
		return fmt.Errorf("clients no rollback")
	})
}
