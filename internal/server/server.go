package server

import (
	"flag"
	"fmt"
	"github.com/heyvito/tot/internal/compressconn"
	"github.com/heyvito/tot/internal/secureconn"
	"log"
	"net"
	"os"
	"os/exec"

	"github.com/heyvito/tot/internal/auth"
	"github.com/heyvito/tot/internal/protocol"
	"github.com/heyvito/tot/internal/pty"
)

func Main() {
	port := flag.Int("port", 2222, "Port to listen on")
	shell := flag.String("shell", "/bin/zsh", "Shell to launch")
	authFile := flag.String("authorized-keys", "authorized_keys", "Authorized public keys file")
	flag.Parse()

	authorizedKeys, err := auth.LoadAuthorizedKeys(*authFile)
	if err != nil {
		log.Fatalf("Failed to load authorized_keys: %v", err)
	}

	addr := fmt.Sprintf(":%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Listen error: %v", err)
	}
	log.Printf("Listening on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}

		connSecure, err := secureconn.Wrap(conn, true)
		if err != nil {
			log.Fatal(err)
		}

		conn, err = compressconn.Wrap(connSecure)
		if err != nil {
			log.Fatal(err)
		}

		go handleConn(conn, *shell, authorizedKeys)
	}
}

func handleConn(conn net.Conn, shell string, authorizedKeys map[string]auth.PublicKey) {
	defer conn.Close()

	clientKey, err := auth.PerformAuth(conn, authorizedKeys)
	if err != nil {
		log.Printf("Auth failed from %s: %v", conn.RemoteAddr(), err)
		conn.Write([]byte(fmt.Sprintf("Auth failed: %v\n", err)))
		return
	}
	log.Printf("Authenticated %s (%s)", conn.RemoteAddr(), clientKey)

	cmd := exec.Command(shell)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("TOT_CONNECTION=true"))
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Printf("PTY error: %v", err)
		return
	}
	defer func() {
		_ = ptmx.Close()
		_ = cmd.Process.Kill()
	}()

	go protocol.HandleInput(conn, ptmx)
	_, _ = protocol.CopyOutput(conn, ptmx)
}
