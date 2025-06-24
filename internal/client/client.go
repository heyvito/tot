package client

import (
	"flag"
	"github.com/heyvito/tot/internal/compressconn"
	"github.com/heyvito/tot/internal/secureconn"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/heyvito/tot/internal/auth"
	"github.com/heyvito/tot/internal/protocol"
	"golang.org/x/term"
)

func Main() {
	addr := flag.String("addr", "127.0.0.1:2222", "Server address")
	keyFile := flag.String("key", os.Getenv("HOME")+"/.ssh/id_ed25519", "Private key file")
	flag.Parse()

	signer, err := auth.LoadPrivateKey(*keyFile)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatal("Dial error:", err)
	}
	defer conn.Close()
	connSecure, err := secureconn.Wrap(conn, false)
	if err != nil {
		log.Fatal(err)
	}

	conn, err = compressconn.Wrap(connSecure)
	if err != nil {
		log.Fatal(err)
	}

	if err := auth.PerformAuthClient(conn, signer); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)
	go func() {
		for range sig {
			protocol.SendWindowSize(conn)
		}
	}()
	protocol.SendWindowSize(conn)

	go io.Copy(conn, os.Stdin)
	io.Copy(os.Stdout, conn)
}
