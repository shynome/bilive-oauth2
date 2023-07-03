package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/shynome/bilive-oauth2/danmu"
)

func TestRandomHex(t *testing.T) {
	s := try.To1(randomHex(8))
	l := len(s)
	t.Log(l)
}

func TestDanmu(t *testing.T) {
	dd := NewDisptacher[Danmu]()
	r, cmd := danmu.Connect("898286", "")
	try.To(cmd.Start())
	go func() {
		for {
			line, _ := try.To2(r.ReadLine())
			go func(line string) {
				arr := strings.SplitN(line, "|", 2)
				if len(arr) != 2 {
					return
				}
				dd.Dispatch(Danmu{UID: arr[0], Content: arr[1]})
			}(string(line))
		}
	}()
	for i := 0; i < 2; i++ {
		go func(i int) {
			defer err2.Catch()
			ctx := context.Background()
			if i == 0 {
				ctx, _ = context.WithTimeout(ctx, 10*time.Second)
			}
			vid := fmt.Sprintf("%d", i)
			done, l := ctx.Done(), dd.Listen(vid)
			defer dd.Free(vid)
			for {
				select {
				case d := <-l:
					log.Println("proc", i, d)
				case <-done:
					log.Println("proc", i, "exit")
					return
				}
			}
		}(i)
	}
	cmd.Wait()
}
