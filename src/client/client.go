package client

import (
	"fmt"
	"sync"
	"tsunagi/src/rpc"
)

var (
	ErrReqTimeout   = fmt.Errorf("request hit deadline trying to send")
	ErrClientClosed = fmt.Errorf("the client connection is closed")
)

type TsunagiClient struct {
	mu      sync.RWMutex
	clients map[string]*rpcClient

	respChanPool sync.Pool
}

func New() *TsunagiClient {

	return &TsunagiClient{
		clients: map[string]*rpcClient{},
		respChanPool: sync.Pool{
			New: func() any {
				// buffered 1 so send goroutine never blocks if caller isn't ready immediately
				return make(chan error)
			},
		},
	}
}

func (c *TsunagiClient) getRespChan() chan error {
	return c.respChanPool.Get().(chan error)
}

func (c *TsunagiClient) putRespChan(ch chan error) {
	c.respChanPool.Put(ch)
}

func (c *TsunagiClient) getClient(addr string) *rpcClient {

	client, ok := func() (*rpcClient, bool) {
		c.mu.RLock()
		defer c.mu.RUnlock()
		client, ok := c.clients[addr]
		return client, ok
	}()

	if !ok {

		client = &rpcClient{
			addr: addr,
			send: make(chan sendRequest),
			read: make(chan *rpc.Event),
			exit: make(chan bool),
		}
		go client.processSends()
		go client.processReads()

		func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			c.clients[addr] = client
		}()
	}

	return client
}

// Read returns a read and exit channel for the client.
// The read channel can be read from forever and will never close.
// If the exit channel is closed the read channel should stop being used.
func (c *TsunagiClient) Read(addr string) (chan *rpc.Event, chan bool) {

	client := c.getClient(addr)

	return client.read, client.exit
}

// Send sends the event and blocks until it was delivered.
// If the event could not be sent due to the client being closed an error is returned.
func (c *TsunagiClient) Send(addr string, event *rpc.Event) error {

	client := c.getClient(addr)

	resp := c.getRespChan()
	defer c.putRespChan(resp)

	req := sendRequest{
		resp:  resp,
		event: event,
	}

	select {
	case client.send <- req:
		break
	case <-client.exit:
		return ErrClientClosed
	}

	return <-resp
}

func (c *TsunagiClient) DeliverMsg(addr string, event *rpc.DeliverRequest) error {
	return c.Send(addr, &rpc.Event{
		Body: &rpc.Event_DeliverRequest{
			DeliverRequest: event,
		},
	})
}

func (c *TsunagiClient) ForwardMsg(addr string, event *rpc.ForwardRequest) error {
	return c.Send(addr, &rpc.Event{
		Body: &rpc.Event_ForwardRequest{
			ForwardRequest: event,
		},
	})
}
