package auth

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

type PublicKey = ssh.PublicKey

func LoadAuthorizedKeys(file string) (map[string]PublicKey, error) {
	out := map[string]PublicKey{}

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	for len(data) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(data)
		if err != nil {
			break
		}
		out[string(pubKey.Marshal())] = pubKey
		data = rest
	}
	return out, nil
}

func LoadPrivateKey(file string) (ssh.Signer, error) {
	keyBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(keyBytes)
}

func PerformAuth(conn net.Conn, authorized map[string]PublicKey) (string, error) {
	nonce := make([]byte, 32)
	rand.Read(nonce)
	_, _ = conn.Write([]byte(fmt.Sprintf("nonce:%s\n", base64.StdEncoding.EncodeToString(nonce))))

	reader := bufio.NewReader(conn)

	pubLine, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	pubLine = strings.TrimSpace(pubLine)

	sigLine, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	sigLine = strings.TrimSpace(sigLine)

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(pubLine))
	if err != nil {
		return "", errors.New("invalid public key")
	}

	if _, ok := authorized[string(pubKey.Marshal())]; !ok {
		return "", errors.New("unauthorized key")
	}

	sigData, err := base64.StdEncoding.DecodeString(sigLine)
	if err != nil {
		return "", errors.New("invalid signature encoding")
	}
	sig := &ssh.Signature{
		Format: pubKey.Type(),
		Blob:   sigData,
	}
	err = pubKey.Verify(nonce, sig)
	if err != nil {
		return "", errors.New("signature verification failed")
	}

	_, _ = conn.Write([]byte("OK\n"))
	return pubLine, nil
}

func PerformAuthClient(conn net.Conn, signer ssh.Signer) error {
	reader := bufio.NewReader(conn)

	nonceLine, err := reader.ReadString('\n')
	if err != nil || !strings.HasPrefix(nonceLine, "nonce:") {
		return errors.New("failed to read nonce")
	}
	nonceB64 := strings.TrimPrefix(strings.TrimSpace(nonceLine), "nonce:")
	nonce, err := base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		return errors.New("invalid nonce encoding")
	}

	sig, err := signer.Sign(rand.Reader, nonce)
	if err != nil {
		return err
	}

	pubKeyLine := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(signer.PublicKey())))
	_, _ = conn.Write([]byte(pubKeyLine + "\n"))
	_, _ = conn.Write([]byte(base64.StdEncoding.EncodeToString(sig.Blob) + "\n"))

	resp, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if strings.TrimSpace(resp) != "OK" {
		return errors.New("auth rejected by server: " + resp)
	}
	return nil
}
