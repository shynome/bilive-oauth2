package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/shynome/err0/try"
)

type MsgType string

const (
	MsgInit    MsgType = "init"
	MsgDanmu   MsgType = "danmu"
	MsgVerfied MsgType = "verified"
)

type Msg[T any] struct {
	Type MsgType `json:"type"`
	Data T       `json:"data"`
}

func (msg Msg[T]) WriteTry(w StreamWriter) {
	defer w.Flush()
	id := time.Now().Unix()
	try.To1(fmt.Fprintf(w, "id:%d\n", id))
	try.To1(io.WriteString(w, "data:"))
	try.To(json.NewEncoder(w).Encode(msg))
	try.To1(io.WriteString(w, "\n"))
	// 结束该 Event
	try.To1(io.WriteString(w, "\n"))
}
