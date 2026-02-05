package goaio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
)

type Transport interface {
	Connect(ctx context.Context) (Connection, error)
}

type StdioTransport struct{}

type Connection interface {
	Read()
	Write()
	Close() error
}

type nopCloserWriter struct {
	io.Writer
}

type msgOrErr struct {
	msg jsonrpc.RawMessage
	err error
}

func (nopCloserWriter) Close() error {
	return nil
}

type ioConn struct {
	protocolVersion string // protocol version for SAXS streams

	rwc io.ReadWriteCloser // stream for data

	incoming <-chan msgOrErr // incoming msg

	outgoingBatch []jsonrpc.Message // batching requests

	queue []jsonrpc.Message // unread from the last batch

	closedOnce sync.Once
	closed     chan struct{}
	closedErr  error
}

func (c *ioConn) Close() error {
	return nil
}

func (c *ioConn) Write() {
}

func (c *ioConn) Read() {
}

type rwc struct {
	rc io.ReadCloser
	wc io.WriteCloser
}

func (r rwc) Read(p []byte) (n int, err error) {
	return r.rc.Read(p)
}

func (r rwc) Write(p []byte) (n int, err error) {
	return r.wc.Write(p)
}

func (r rwc) Close() error {
	rcErr := r.rc.Close()

	var wcErr error
	if r.wc != nil { // we only allow a nil writer in unit tests
		wcErr = r.wc.Close()
	}

	return errors.Join(rcErr, wcErr)
}

// Connect implements the [Transport] interface.
func (*StdioTransport) Connect(context.Context) (Connection, error) {
	return newIOConn(rwc{os.Stdin, nopCloserWriter{os.Stdout}}), nil
}

func newIOConn(rwc io.ReadWriteCloser) *ioConn {
	var (
		incoming = make(chan msgOrErr)
		closed   = make(chan (struct{}))
	)

	go func() {
		var raw json.RawMessage
		err := dec.Decode(&raw)
		// If decoding was successful, check for trailing data at the end of the stream.
		if err == nil {
			// Read the next byte to check if there is trailing data.
			var tr [1]byte
			if n, readErr := dec.Buffered().Read(tr[:]); n > 0 {
				// If read byte is not a newline, it is an error.
				// Support both Unix (\n) and Windows (\r\n) line endings.
				if tr[0] != '\n' && tr[0] != '\r' {
					err = fmt.Errorf("invalid trailing data at the end of stream")
				}
			} else if readErr != nil && readErr != io.EOF {
				err = readErr
			}
		}
		select {
		case incoming <- msgOrErr{msg: raw, err: err}:
		case <-closed:
			return
		}
		if err != nil {
			return
		}
	}()

	return &ioConn{
		rwc:      rwc,
		incoming: incoming,
		closed:   closed,
	}
}
