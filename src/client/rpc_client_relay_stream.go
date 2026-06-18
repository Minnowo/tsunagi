package client

import (
	"context"
	"sync"
	"tsunagi/src/rpc"

	"google.golang.org/grpc"
)

// ClientRelayStream is a bidirectional stream from client to relay.
type ClientRelayStream struct {
	conn *TsunagiConn

	// mu locks both the stream and didStart to ensure the processing of this stream only happens once
	mu sync.Mutex

	stream grpc.BidiStreamingClient[rpc.ClientEvent, rpc.RelayEvent]

	// didStart is true if the goroutine loop was started
	didStart bool

	// didExit is true f the goroutine loop has ended
	didExit bool

	// send events recieved on this channel are sent through RPC.
	// This is never to be closed.
	send chan SendEventRequest

	// read events recieved on rpc are sent through this channel
	// This is never to be closed.
	read chan *rpc.RelayEvent

	// this channel SHOULD BE closed to signal exit (shutdown stream, stop goroutines, etc)
	exit chan struct{}
}

func NewClientRelayStream(conn *TsunagiConn, sendChanSize int) *ClientRelayStream {
	return &ClientRelayStream{
		conn:     conn,
		didStart: false,
		didExit:  false,
		send:     make(chan SendEventRequest, sendChanSize),
		read:     make(chan *rpc.RelayEvent),
		exit:     make(chan struct{}),
	}
}

func (r *ClientRelayStream) SendChan() chan<- SendEventRequest {
	return r.send
}

func (r *ClientRelayStream) ReadChan() <-chan *rpc.RelayEvent {
	return r.read
}

func (r *ClientRelayStream) ExitChan() chan struct{} {
	return r.exit
}

func (r *ClientRelayStream) Close() error {
	close(r.exit)
	r.disconnect()
	return nil
}

func (r *ClientRelayStream) Start() {

	if r.didStart {
		return
	}

	r.mu.Lock()
	didStart := r.didStart
	r.didStart = true
	r.mu.Unlock()

	if didStart {
		return
	}

	go func() {
		defer func() {
			r.didExit = true
			r.Close() // if a panic happens we need to close this
		}()
		processSends(r.exit, r.send, r)
	}()

	go func() {
		defer func() {
			r.didExit = true
			r.Close() // if a panic happens we need to close this
		}()
		processReads(r)
	}()
}

func (r *ClientRelayStream) DidExit() bool {
	return r.didExit
}

func (r *ClientRelayStream) IsConnected() bool {
	return r.stream != nil
}

func (r *ClientRelayStream) rpcSend(req any) error {

	reqq, ok := req.(*rpc.ClientEvent)

	if !ok {
		return ErrInvalidSendType
	}

	return r.stream.Send(reqq)
}

func (r *ClientRelayStream) disconnect() {

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stream == nil {
		return
	}

	r.stream.CloseSend()
	r.stream = nil
}

func (r *ClientRelayStream) connect(ctx context.Context) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stream != nil {
		return nil
	}

	authCtx, err := GetAuthContext(r.conn, ctx)

	if err != nil {
		return err
	}

	stream, err := r.conn.Tsu.ConnectClient(authCtx)

	if err != nil {
		return err
	}

	r.stream = stream

	return nil
}
