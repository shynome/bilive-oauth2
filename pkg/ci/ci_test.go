package ci

import (
	"testing"
)

func TestCiLength(t *testing.T) {
	l := len(CiList)
	_ = l
	// assert.Equal(l, 263309)
}

func TestCiRandom(t *testing.T) {
	s := Random(3)
	t.Log(s)
}
