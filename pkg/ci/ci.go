package ci

import (
	_ "embed"
	"math/rand"
	"strings"
)

//go:generate go run ./gen ci.txt

//go:embed ci.txt
var s string

var CiList []string
var CiListLen int

func init() {
	CiList = strings.Split(s, "|")
	CiListLen = len(CiList)
}

func Random(n int) (s string) {
	for i := 0; i < n; i++ {
		s += CiList[rand.Intn(CiListLen)]
	}
	return
}
