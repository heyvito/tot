package compressconn

import (
	"compress/flate"
	"io"
	"net"
)

type Conn struct {
	net.Conn
	writer *flate.Writer
	reader io.ReadCloser
}

func Wrap(conn net.Conn) (*Conn, error) {
	// Incoming data (decompression)
	reader := flate.NewReader(conn)

	// Outgoing data (compression)
	writer, err := flate.NewWriter(conn, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}

	return &Conn{
		Conn:   conn,
		writer: writer,
		reader: reader,
	}, nil
}

// Write compresses and writes data.
func (c *Conn) Write(p []byte) (int, error) {
	n, err := c.writer.Write(p)
	if err != nil {
		return n, err
	}
	// Important: flush to ensure data is sent immediately
	err = c.writer.Flush()
	return n, err
}

// Read decompresses incoming data.
func (c *Conn) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

// Close closes both compression streams and the underlying connection.
func (c *Conn) Close() error {
	_ = c.writer.Close()
	_ = c.reader.Close()
	return c.Conn.Close()
}
