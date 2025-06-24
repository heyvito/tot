package main

import (
	"fmt"
	"os"

	"github.com/heyvito/tot/internal/client"
	"github.com/heyvito/tot/internal/server"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: tty-over-tcp <server|client> [flags]")
		os.Exit(1)
	}

	mode := os.Args[1]
	os.Args = append([]string{os.Args[0]}, os.Args[2:]...) // Shift args

	switch mode {
	case "server":
		server.Main()
	case "client":
		client.Main()
	default:
		fmt.Println("Usage: tty-over-tcp <server|client> [flags]")
		os.Exit(1)
	}
}
