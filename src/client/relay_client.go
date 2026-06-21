package client

import (
	"fmt"
	"sync"
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"github.com/flynn/noise"
)

var (
	ErrInvalidSendType = fmt.Errorf("invalid send type")
	ErrStreamClosed    = fmt.Errorf("the stream is closed")
	ErrClientShutdown  = fmt.Errorf("the client is shutdown")
)

type _ConnRelayStream struct {
	conn   *TsunagiConn
	stream *RelayRelayStream
}

type RelayRelayClient struct {
	mu           sync.RWMutex
	clients      map[string]_ConnRelayStream
	sendChanSize int
	ackChan      chan ClientAck
	identity     noise.DHKey
}

func NewRelayRelayClient(identity noise.DHKey, sendChanSize int, ackChanSize int) *RelayRelayClient {

	return &RelayRelayClient{
		clients:      map[string]_ConnRelayStream{},
		sendChanSize: sendChanSize,
		ackChan:      make(chan ClientAck, ackChanSize),
		identity:     identity,
	}
}

func (c *RelayRelayClient) getStream(addr string) (*RelayRelayStream, error) {

	client, ok := func() (_ConnRelayStream, bool) {
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

		client = _ConnRelayStream{
			conn:   &conn,
			stream: NewRelayRelayStream(&conn, c.ackChan, c.sendChanSize),
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
			client.stream = NewRelayRelayStream(client.conn, c.ackChan, c.sendChanSize)
		}()
	}

	client.stream.Start()

	return client.stream, nil
}

// ReadAck returns the chan which recieves all the ack messages.
func (c *RelayRelayClient) ReadAck() <-chan ClientAck {
	return c.ackChan
}

// Send puts the event into the send queue for which will later be sent.
// Returns nil if the event was put into the queue, otherwise an error.
// If the event has a MessageID, an Ack message will be generated on the RelayRelayClient.ackChan if the message is delivered to the other relay.
func (c *RelayRelayClient) Send(addr string, sendingClientID data.Identifier, event *rpc.RelayEvent) error {

	stream, err := c.getStream(addr)

	if err != nil {
		return err
	}

	req := RelayRelayClientEvent{
		Sender: sendingClientID,
		Event:  event,
	}

	select {
	case <-stream.exit:
		return ErrStreamClosed
	case stream.send <- req:
		return nil
	}
}
