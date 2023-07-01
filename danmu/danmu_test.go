package danmu

import (
	"testing"
	"time"

	"github.com/lainio/err2/try"
)

func TestRun(t *testing.T) {
	r, cmd := Connect("24393")
	try.To(cmd.Start())
	go func() {
		for {
			line, _ := try.To2(r.ReadLine())
			t.Log(line)
		}
	}()
	var w = make(chan error)
	go func() {
		w <- cmd.Wait()
	}()
	select {
	case <-time.After(5 * time.Second):
		t.Log("pass")
	case err := <-w:
		t.Error(err)
	}
}
