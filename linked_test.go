package main

import (
	"context"
	"testing"
	"time"

	"github.com/shynome/err0/try"
)

func TestFindUID(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	uid := try.To1(FindUID(ctx, "shynome"))
	if uid != "6660627" {
		t.Error("find uid failed. got", uid)
	}
}
