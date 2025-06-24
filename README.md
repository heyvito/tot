# tot (tty-over-tcp)

A minimal, secure, SSH-authenticated remote PTY tunnel over TCP.

- **SSH key authentication** (OpenSSH compatible).
- **Encrypted** using ephemeral Curve25519 + AES-CTR.
- **Compressed** with DEFLATE (zlib-compatible).
- **Full terminal support** with window resizing.
- Single binary. No SSH server needed.

---

## Features
- Authentication with your SSH keys (`~/.ssh/id_ed25519` or `authorized_keys`).
- Encrypted using ephemeral ECDH (Curve25519) with AES-CTR.
- Compression (DEFLATE).
- Fully interactive terminal with window resizing (`SIGWINCH` supported).
- Runs as **client** or **server** from a single binary.

---

## Build

```bash
make
```

Produces:

```bash
./tot
```

---

## Usage

### Server

```bash
./tot server \
  -port 2222 \
  -shell /bin/zsh \
  -authorized-keys ~/.ssh/authorized_keys
```

- `-port`: TCP port to listen on.
- `-shell`: Command to run for the session (default `/bin/zsh`).
- `-authorized-keys`: File with authorized public keys (OpenSSH format).

---

### Client

```bash
./tot client \
  -addr 127.0.0.1:2222 \
  -key ~/.ssh/id_ed25519
```

- `-addr`: Address of the server (`host:port`).
- `-key`: Private SSH key to authenticate (OpenSSH PEM format).

---

## Authentication Flow
- Server sends a random challenge (`nonce`).
- Client signs with their SSH private key.
- Server checks the signature against its `authorized_keys`.

---

## Security Model
- Ephemeral Curve25519 key exchange (forward secrecy).
- AES-CTR symmetric encryption.
- SSH key authentication.
---

## ⚠️ Disclaimer
This project is a prototype-level tool for secure remote terminal access over TCP. Use with care. Contributions welcome.

---

## License

```
MIT License

Copyright (c) 2025 - Vito Sartori

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
