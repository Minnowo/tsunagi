package client

import (
	"context"
	"sync"
	"time"
	"tsunagi/src/rpc"

	"github.com/rs/zerolog/log"
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
	send chan *rpc.ClientEvent

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
		send:     make(chan *rpc.ClientEvent, sendChanSize),
		read:     make(chan *rpc.RelayEvent),
		exit:     make(chan struct{}),
	}
}

func (r *ClientRelayStream) SendChan() chan<- *rpc.ClientEvent {
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
		r.processSends()
	}()

	go func() {
		defer func() {
			r.didExit = true
			r.Close() // if a panic happens we need to close this
		}()
		r.processReads()
	}()
}

func (r *ClientRelayStream) DidExit() bool {
	return r.didExit
}

func (r *ClientRelayStream) IsConnected() bool {
	return r.stream != nil
}

func (r *ClientRelayStream) rpcSend(req *rpc.ClientEvent) error {
	return r.stream.Send(req)
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

func (r *ClientRelayStream) processReads() {

	for {

		select {
		case <-r.exit:
			return
		default:
			break
		}

		if !r.IsConnected() {

			if err := r.connect(context.Background()); err != nil {

				log.Warn().Err(err).Msg("failed to connect, retrying in 1s")

				time.Sleep(time.Second)

				continue
			}
		}

		event, err := r.stream.Recv()

		if err != nil {

			log.Warn().Err(err).Msg("ClientRelayStream.processSend: stream recv failed")

			r.disconnect()

			continue
		}

		r.read <- event
	}
}

func (r *ClientRelayStream) processSends() {

	var req *rpc.ClientEvent
	for {
		select {
		case <-r.exit:
			r.disconnect()
			return
		default:
			break
		}

		if !r.IsConnected() {

			if err := r.connect(context.Background()); err != nil {

				log.Warn().Err(err).Msg("failed to connect, retrying in 1s")
				time.Sleep(time.Second)

				continue
			}
		}

		if req != nil {

			if err := r.rpcSend(req); err != nil {

				log.Warn().Err(err).Msg("stream send failed, reconnecting")

				r.disconnect()
				time.Sleep(time.Second)

				continue
			} else {
				req = nil
			}
		}

		select {
		case nextEvent, ok := <-r.send:

			if !ok {
				log.Debug().Msg("send channel closed")
				return
			}

			req = nextEvent
		}
	}
}
