package client

import (
	"sync"
	"tsunagi/src/rpc"
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
}

func NewClientRelayClient(sendChanSize int) *ClientRelayClient {

	return &ClientRelayClient{
		clients:      map[string]_ConnClientStream{},
		sendChanSize: sendChanSize,
	}
}

func (c *ClientRelayClient) getRespChan() chan error {
	return make(chan error, 1)
}

func (c *ClientRelayClient) getStream(addr string) (*ClientRelayStream, error) {

	client, ok := func() (_ConnClientStream, bool) {
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
func (c *ClientRelayClient) Send(addr string, event *rpc.Event) error {

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

func (c *ClientRelayClient) DeliverMsg(addr string, event *rpc.DeliverRequest) error {
	return c.Send(addr, &rpc.Event{
		Body: &rpc.Event_DeliverRequest{
			DeliverRequest: event,
		},
	})
}

func (c *ClientRelayClient) ForwardMsg(addr string, event *rpc.ForwardRequest) error {
	return c.Send(addr, &rpc.Event{
		Body: &rpc.Event_ForwardRequest{
			ForwardRequest: event,
		},
	})
}

func (c *ClientRelayClient) GetReadHandle(addr string) (<-chan *rpc.Event, <-chan struct{}, error) {

	stream, err := c.getStream(addr)

	if err != nil {
		return nil, nil, err
	}

	return stream.ReadChan(), stream.ExitChan(), nil
}
