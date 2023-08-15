package adapter

import (
	"context"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/pocketbase/dbx"
	"github.com/shynome/bilive-oauth2/pkg/oauth2/model"
)

var _ oauth2.ClientStore = (*Adapter)(nil)

func (app *Adapter) GetByID(ctx context.Context, id string) (_ oauth2.ClientInfo, err error) {
	defer err2.Handle(&err)
	dao := app.Dao()
	var c = new(model.OAuth2Client)
	q := dao.ModelQuery(c).WithContext(ctx).
		Where(dbx.HashExp{"id": id}).
		Limit(1)
	try.To(q.One(c))
	return c, nil
}
