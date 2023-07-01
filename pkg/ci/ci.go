package ci

import (
	_ "embed"
	"strings"
)

//go:embed ci.txt
var s string

var CiList []string

func init() {
	CiList = strings.Split(s, "|")
}
