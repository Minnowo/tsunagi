package client

import (
	"context"
	"sync"
	"tsunagi/src/rpc"

	"google.golang.org/grpc"
)

// RelayRelayStream is a one directional stream from a relay to another relay.
type RelayRelayStream struct {
	conn *TsunagiConn

	didStart bool
	didExit  bool

	mu     sync.Mutex
	stream grpc.ClientStreamingClient[rpc.RelayEvent, rpc.Empty]

	// send events recieved on this channel are sent through RPC.
	// This is never to be closed.
	send chan SendEventRequest

	// this channel SHOULD BE closed to signal exit (shutdown stream, stop goroutines, etc)
	exit chan struct{}
}

func NewRelayRelayStream(conn *TsunagiConn, sendChanSize int) *RelayRelayStream {
	return &RelayRelayStream{
		conn:     conn,
		didStart: false,
		didExit:  false,
		send:     make(chan SendEventRequest, sendChanSize),
		exit:     make(chan struct{}),
	}
}

func (r *RelayRelayStream) SendChan() chan<- SendEventRequest {
	return r.send
}

func (r *RelayRelayStream) ExitChan() chan struct{} {
	return r.exit
}

func (r *RelayRelayStream) Close() error {
	close(r.exit)
	r.disconnect()
	return nil
}

func (r *RelayRelayStream) Start() {

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
}

func (r *RelayRelayStream) DidExit() bool {
	return r.didExit
}

func (r *RelayRelayStream) IsConnected() bool {
	return r.stream != nil
}

func (r *RelayRelayStream) rpcSend(req any) error {

	reqq, ok := req.(*rpc.RelayEvent)

	if !ok {
		return ErrInvalidSendType
	}

	return r.stream.Send(reqq)
}

func (r *RelayRelayStream) disconnect() {

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stream == nil {
		return
	}

	r.stream.CloseAndRecv()
	r.stream = nil
}

func (r *RelayRelayStream) connect(ctx context.Context) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stream != nil {
		return nil
	}

	authCtx, err := GetAuthContext(r.conn, ctx)

	if err != nil {
		return err
	}

	stream, err := r.conn.Tsu.ConnectRelay(authCtx)

	if err != nil {
		return err
	}

	r.stream = stream

	return nil
}
