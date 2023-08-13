package danmu

import (
	"bufio"
	"bytes"
	_ "embed"
	"io"
	"os"
	"os/exec"
)

//go:generate esbuild --bundle --platform=node --outdir=dist --minify danmu.mjs

//go:embed dist/danmu.js
var danmujs []byte

func Connect(room, bilipage string) (r *bufio.Reader, cmd *exec.Cmd) {
	cmd = exec.Command("node", "-", room, bilipage)
	cmd.Stdin = bytes.NewReader(danmujs)
	cmdReader, cmdOut := io.Pipe()
	cmd.Stdout = cmdOut
	cmd.Stderr = os.Stderr
	r = bufio.NewReader(cmdReader)
	return
}
