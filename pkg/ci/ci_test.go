package ci

import (
	"testing"

	"github.com/lainio/err2/assert"
)

func TestCiLength(t *testing.T) {
	l := len(CiList)
	assert.Equal(l, 263309)
}

func TestCiRandom(t *testing.T) {
	s := Random(3)
	t.Log(s)
}
