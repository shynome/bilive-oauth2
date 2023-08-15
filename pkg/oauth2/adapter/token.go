package adapter

import (
	"context"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/lainio/err2"
)

var _ oauth2.TokenStore = (*Adapter)(nil)

func (app *Adapter) Create(ctx context.Context, info oauth2.TokenInfo) (err error) {
	defer err2.Handle(&err)
	return
}

// delete the authorization code
func (app *Adapter) RemoveByCode(ctx context.Context, code string) (err error) {
	defer err2.Handle(&err)
	return
}

// use the access token to delete the token information
func (app *Adapter) RemoveByAccess(ctx context.Context, access string) (err error) {
	defer err2.Handle(&err)
	return
}

// use the refresh token to delete the token information
func (app *Adapter) RemoveByRefresh(ctx context.Context, refresh string) (err error) {
	defer err2.Handle(&err)
	return
}

// use the authorization code for token information data
func (app *Adapter) GetByCode(ctx context.Context, code string) (_ oauth2.TokenInfo, err error) {
	defer err2.Handle(&err)
	return
}

// use the access token for token information data
func (app *Adapter) GetByAccess(ctx context.Context, access string) (_ oauth2.TokenInfo, err error) {
	defer err2.Handle(&err)
	return
}

// use the refresh token for token information data
func (app *Adapter) GetByRefresh(ctx context.Context, refresh string) (_ oauth2.TokenInfo, err error) {
	defer err2.Handle(&err)
	return
}

type TokenInfo struct {
}
