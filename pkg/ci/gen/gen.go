package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/lainio/err2/try"
)

const link = "https://github.com/pwxcoo/chinese-xinhua/raw/master/data/ci.json"

type Ci struct {
	Ci string `json:"ci"`
}

func main() {
	resp := try.To1(http.Get(link))
	var v []Ci
	try.To(json.NewDecoder(resp.Body).Decode(&v))
	var ss []string
	r := regexp.MustCompile(`\(|\s`)
	for _, ci := range v {
		w := ci.Ci
		if matched := r.MatchString(w); matched {
			continue
		}
		ss = append(ss, w)
	}
	s := strings.Join(ss, "|")
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "output file required")
		os.Exit(1)
		return
	}
	f := os.Args[1]
	try.To(os.WriteFile(f, []byte(s), os.ModePerm))
}
