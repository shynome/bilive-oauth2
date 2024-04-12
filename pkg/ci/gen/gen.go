package main

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/shynome/err0/try"
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
		for _, b := range blocks {
			if strings.Index(w, b) == -1 {
				continue
			}
		}
		ss = append(ss, w)
	}
	s := strings.Join(ss, "|")
	f := "../ci.txt"
	if len(os.Args) >= 2 {
		f = os.Args[1]
	}
	try.To(os.WriteFile(f, []byte(s), os.ModePerm))
}

var blocks = []string{
	"狗", "鸡", "死", "干",
	"娼", "财", "雌", "奸",
	"公", "寡", "母", "兵",
	"不", "荡", "斧", "嬖",
	"奴", "女", "色", "童",
	"狱", "罂", "刀", "大",
	"共产", "主义", "精神",
	"乳",
}
