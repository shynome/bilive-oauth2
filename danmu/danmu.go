package danmu

import (
	"bufio"
	_ "embed"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/lainio/err2/try"
)

//go:generate bun build  --outdir dist --minify danmu.mjs

//go:embed dist/danmu.js
var danmujs []byte

var f string

func init() {
	dir := try.To1(os.MkdirTemp("", "bilive-oauth2"))
	f = filepath.Join(dir, "danmu.js")
	try.To(os.WriteFile(f, danmujs, os.ModePerm))
}

func Connect(room string) (r *bufio.Reader, cmd *exec.Cmd) {
	cmd = exec.Command("bun", "run", f, room)
	cmdReader, cmdOut := io.Pipe()
	cmd.Stdout = cmdOut
	cmd.Stderr = os.Stderr
	r = bufio.NewReader(cmdReader)
	return
}
