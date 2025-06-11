package db

import (
	"github.com/go-oauth2/oauth2/v4"
	"github.com/pocketbase/pocketbase/core"
)

type Client struct {
	core.BaseRecordProxy
}

var _ oauth2.ClientInfo = (*Client)(nil)

func (c *Client) GetID() string     { return c.GetString("application") }
func (c *Client) GetSecret() string { return c.GetString("secret") }
func (c *Client) GetDomain() string { return c.GetString("domain") }
func (c *Client) IsPublic() bool    { return c.GetBool("public") }
func (c *Client) GetUserID() string { return c.GetString("application") }
