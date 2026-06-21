package client

import (
	"sync"
	"tsunagi/src/rpc"

	"github.com/flynn/noise"
)

type _ConnClientStream struct {
	conn   *TsunagiConn
	stream *ClientRelayStream
}

type ClientRelayClient struct {
	mu           sync.RWMutex
	clients      map[string]_ConnClientStream
	respChanPool sync.Pool
	sendChanSize int
	identity     noise.DHKey
	AutoMsgID    bool
	autoMsgCount uint64
}

func NewClientRelayClient(identity noise.DHKey, sendChanSize int) *ClientRelayClient {

	return &ClientRelayClient{
		clients:      map[string]_ConnClientStream{},
		sendChanSize: sendChanSize,
		identity:     identity,
	}
}

func (c *ClientRelayClient) getStream(addr string) (*ClientRelayStream, error) {

	client, ok := func() (_ConnClientStream, bool) {
		c.mu.RLock()
		defer c.mu.RUnlock()
		client, ok := c.clients[addr]
		return client, ok
	}()

	if !ok {

		conn := NewTsunagiConn(addr, c.identity)

		if err := conn.Connect(); err != nil {
			return nil, err
		}

		client = _ConnClientStream{
			conn:   &conn,
			stream: NewClientRelayStream(&conn, c.sendChanSize),
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
			client.stream = NewClientRelayStream(client.conn, c.sendChanSize)
		}()
	}

	client.stream.Start()

	return client.stream, nil
}

// Send sends the event and blocks until it was delivered or the stream exits.
// If the event could not be sent due to the client being closed, or the stream exiting, an error is returned.
func (c *ClientRelayClient) Send(addr string, event *rpc.ClientEvent) error {

	stream, err := c.getStream(addr)

	if err != nil {
		return err
	}

	if c.AutoMsgID {
		// not perfect but good enough for debugging
		rpc.GetSetClientMessageID(event, func(id uint64) (uint64, error) {
			c.autoMsgCount++
			return c.autoMsgCount, nil
		})
	}

	select {
	case <-stream.exit:
		return ErrStreamClosed
	case stream.send <- event:
		return nil
	}
}

func (c *ClientRelayClient) GetReadHandle(addr string) (<-chan *rpc.RelayEvent, <-chan struct{}, error) {

	stream, err := c.getStream(addr)

	if err != nil {
		return nil, nil, err
	}

	return stream.ReadChan(), stream.ExitChan(), nil
}
