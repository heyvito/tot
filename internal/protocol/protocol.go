package protocol

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/creack/pty"
	"golang.org/x/term"
)

func HandleInput(conn net.Conn, ptmx *os.File) {
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		// Check for control message (resize)
		if n >= 2 && buf[0] == 0 {
			parts := strings.SplitN(string(buf[1:n]), ":", 2)
			if len(parts) == 2 {
				var rows, cols int
				_, err1 := fmt.Sscanf(parts[0], "%d", &rows)
				_, err2 := fmt.Sscanf(parts[1], "%d", &cols)
				if err1 == nil && err2 == nil {
					pty.Setsize(ptmx, &pty.Winsize{
						Rows: uint16(rows),
						Cols: uint16(cols),
					})
				}
			}
		} else {
			_, _ = ptmx.Write(buf[:n])
		}
	}
}

func CopyOutput(conn net.Conn, ptmx *os.File) (int64, error) {
	return io.Copy(conn, ptmx)
}

func SendWindowSize(conn net.Conn) {
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return
	}
	msg := fmt.Sprintf("\x00%d:%d\n", height, width)
	_, _ = conn.Write([]byte(msg))
}
