package client

import (
	"fmt"
	"sync"
	"tsunagi/src/rpc"
)

var (
	ErrInvalidSendType   = fmt.Errorf("invalid send type")
	ErrStreamClosed   = fmt.Errorf("the stream is closed")
	ErrClientShutdown = fmt.Errorf("the client is shutdown")
)

type _ConnRelayStream struct {
	conn   *TsunagiConn
	stream *RelayRelayStream
}
type RelayRelayClient struct {
	mu           sync.RWMutex
	clients      map[string]_ConnRelayStream
	sendChanSize int
}

func NewRelayRelayClient(sendChanSize int) *RelayRelayClient {

	return &RelayRelayClient{
		clients:      map[string]_ConnRelayStream{},
		sendChanSize: sendChanSize,
	}
}

func (c *RelayRelayClient) getRespChan() chan error {
	return make(chan error, 1)
}

func (c *RelayRelayClient) getStream(addr string) (*RelayRelayStream, error) {

	client, ok := func() (_ConnRelayStream, bool) {
		c.mu.RLock()
		defer c.mu.RUnlock()
		client, ok := c.clients[addr]
		return client, ok
	}()

	if !ok {

		conn := NewTsunagiConn(addr)

		if err := conn.Connect(); err != nil {
			return nil, err
		}

		client = _ConnRelayStream{
			conn:   &conn,
			stream: NewRelayRelayStream(&conn, c.sendChanSize),
		}

		func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			c.clients[addr] = client
		}()
	}

	if client.stream.DidExit() {
		func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			client.stream = NewRelayRelayStream(client.conn, c.sendChanSize)
		}()
	}

	client.stream.Start()

	return client.stream, nil
}

// Send sends the event and blocks until it was delivered or the stream exits.
// If the event could not be sent due to the client being closed, or the stream exiting, an error is returned.
func (c *RelayRelayClient) Send(addr string, event *rpc.RelayEvent) error {

	stream, err := c.getStream(addr)
	if err != nil {
		return err
	}

	resp := c.getRespChan()

	req := SendEventRequest{
		resp:  resp,
		event: event,
	}

	select {
	case <-stream.exit:
		return ErrStreamClosed
	case stream.send <- req:
		break
	}

	select {
	case <-stream.exit:
		select {
		case err := <-resp:
			return err
		default:
			return ErrStreamClosed
		}
	case err := <-resp:
		return err
	}
}

