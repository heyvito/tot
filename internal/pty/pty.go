package pty

import (
	"os"
	"os/exec"

	"github.com/creack/pty"
)

func Start(cmd *exec.Cmd) (*os.File, error) {
	return pty.Start(cmd)
}
