package main

import (
	"testing"

	"github.com/lainio/err2/try"
)

func TestRandomHex(t *testing.T) {
	s := try.To1(randomHex(8))
	l := len(s)
	t.Log(l)
}
