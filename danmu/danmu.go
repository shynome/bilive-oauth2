package danmu

import (
	"bufio"
	_ "embed"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

//go:generate bun build  --outdir dist --minify danmu.mjs

//go:embed dist/danmu.js
var danmujs []byte

func Run(room string) (r *bufio.Reader, cmd *exec.Cmd, err error) {
	defer err2.Handle(&err)
	f := try.To1(savejs2file())
	cmd = exec.Command("bun", "run", f, room)
	cmdReader, cmdOut := io.Pipe()
	cmd.Stdout = cmdOut
	cmd.Stderr = os.Stderr
	r = bufio.NewReader(cmdReader)
	return
}

func savejs2file() (p string, err error) {
	defer err2.Handle(&err)
	dir := try.To1(os.MkdirTemp("", "bilive-oauth2"))
	p = filepath.Join(dir, "danmu.js")
	try.To(os.WriteFile(p, danmujs, os.ModePerm))
	return
}
