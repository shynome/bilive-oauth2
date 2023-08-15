package model

import (
	"github.com/go-oauth2/oauth2/v4"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/types"
)

type OAuth2Client struct {
	models.BaseModel
	Secret string `db:"secret"`
	Domain string `db:"domain"`
	Public bool   `db:"public"`
	UserID string `db:"user_id"`
}

var _ models.Model = (*OAuth2Client)(nil)

func (OAuth2Client) TableName() string { return "oauth_client" }

var _ oauth2.ClientInfo = (*OAuth2Client)(nil)

func (c *OAuth2Client) GetID() string     { return c.Id }
func (c *OAuth2Client) GetSecret() string { return c.Secret }
func (c *OAuth2Client) GetDomain() string { return c.Domain }
func (c *OAuth2Client) IsPublic() bool    { return c.Public }
func (c *OAuth2Client) GetUserID() string { return c.UserID }

type OAuth2Token struct {
	models.BaseModel
	Code      string         `db:"code"`
	Access    string         `db:"access"`
	Refresh   string         `db:"refresh"`
	ExpiresAt types.DateTime `db:"expires_at"`
	Info      types.JsonRaw  `db:"info"`
}

var _ models.Model = (*OAuth2Token)(nil)

func (OAuth2Token) TableName() string { return "oauth_token" }
