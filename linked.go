package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/shynome/bilive-oauth2/v2/db"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
)

var linkedClient *http.Client

func init() {
	if proxy := os.Getenv("BEER_PROXY"); proxy != "" {
		proxy := try.To1(url.Parse(proxy))
		linkedClient = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxy),
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	}
}

func FindUID(ctx context.Context, uname string) (_ string, err error) {
	defer err0.Then(&err, nil, nil)

	if linkedClient == nil {
		return "", fmt.Errorf("no env: BEER_PROXY")
	}

	req := try.To1(http.NewRequestWithContext(ctx, http.MethodGet, "https://api.bilibili.com/x/polymer/web-dynamic/v1/mention/search", nil))
	q := req.URL.Query()
	q.Set("keyword", uname)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("js.fetch.credentials", "include")
	resp := try.To1(linkedClient.Do(req))
	defer resp.Body.Close()
	if code := resp.StatusCode; code != http.StatusOK {
		return "", fmt.Errorf("request failed. got code: %d", code)
	}

	var br BilibiliResponse[MentionResult]
	try.To(json.NewDecoder(resp.Body).Decode(&br))
	if br.Code != 0 {
		return "", fmt.Errorf("status code %v, msg %s", br.Code, br.Message)
	}

	for _, group := range br.Data.Groups {
		for _, item := range group.Items {
			if item.Name == uname {
				return item.UID, nil
			}
		}
	}

	return "", fmt.Errorf("can't find uid by name %s", uname)
}

func TryLinkUnameUID(app core.App, openid, uname string) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	_, err := TryLinkUnameUIDWithCtx(ctx, app, openid, uname)
	return err
}

func TryLinkUnameUIDWithCtx(ctx context.Context, app core.App, openid, uname string) (linked *core.Record, err error) {
	logger := app.Logger()
	defer err0.Then(&err, nil, func() {
		logger.Error("link uname uid failed", "openid", openid, "uname", uname, "error", err)
	})

	linkeds := try.To1(app.FindCachedCollectionByNameOrId(db.TableLinkeds))
	linked, err = app.FindFirstRecordByData(linkeds, "openid", openid)
	if err == nil {
		return linked, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	uid := try.To1(FindUID(ctx, uname))
	linked = core.NewRecord(linkeds)
	linked.Set("openid", openid)
	linked.Set("uname", uname)
	linked.Set("uid", uid)
	try.To(app.Save(linked))
	return linked, nil
}

type BilibiliResponse[T any] struct {
	Code    int    `json:"code"`
	Data    T      `json:"data"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
}

type MentionResult struct {
	Groups []MentionGroup `json:"groups"`
}

type MentionGroup struct {
	Name  string             `json:"group_name"`
	Type  int                `json:"group_type"`
	Items []MentionGroupItem `json:"items"`
}
type MentionGroupItem struct {
	Face string `json:"face"`
	Fans int    `json:"fans"`
	Name string `json:"name"`
	UID  string `json:"uid"`

	OfficialVerifyType int `json:"official_verify_type"`
}
