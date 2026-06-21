package client

import (
	"context"
	"sync"
	"time"
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

// RelayRelayStream is a one directional stream from a relay to another relay.
type RelayRelayStream struct {
	conn *TsunagiConn

	didStart bool
	didExit  bool

	mu     sync.Mutex
	stream grpc.BidiStreamingClient[rpc.RelayEvent, rpc.RelayAck]

	pendingIdCounter uint64
	pendingAck       map[uint64]ClientAck

	// sendAckCh is where all ack messages are sent to.
	// This channel is never closed by the stream, and should never be closed.
	sendAckCh chan<- ClientAck

	// send sends events recieved on this channel are sent through RPC.
	// All objects sent through this stream are owned by the stream, meaning it should be assumed that
	// the contents at the pointer will be modified as needed.
	// This is never to be closed.
	send chan RelayRelayClientEvent

	// this channel SHOULD BE closed to signal exit (shutdown stream, stop goroutines, etc)
	exit chan struct{}
}

type RelayRelayClientEvent struct {
	Sender data.Identifier
	Event  *rpc.RelayEvent
}

func NewRelayRelayStream(conn *TsunagiConn, ackchan chan<- ClientAck, sendChanSize int) *RelayRelayStream {
	return &RelayRelayStream{
		conn:      conn,
		didStart:  false,
		didExit:   false,
		sendAckCh: ackchan,
		send:      make(chan RelayRelayClientEvent, sendChanSize),
		exit:      make(chan struct{}),
	}
}

func (r *RelayRelayStream) SendChan() chan<- RelayRelayClientEvent {
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
		r.processAckReads()
	}()

	go func() {
		defer func() {
			r.didExit = true
			r.Close() // if a panic happens we need to close this
		}()
		r.processRelaySends()
	}()
}

func (r *RelayRelayStream) DidExit() bool {
	return r.didExit
}

func (r *RelayRelayStream) IsConnected() bool {
	return r.stream != nil
}

func (r *RelayRelayStream) rpcSend(req *rpc.RelayEvent) error {
	return r.stream.Send(req)
}

func (r *RelayRelayStream) disconnect() {

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stream == nil {
		return
	}

	r.stream.CloseSend()
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

func (r *RelayRelayStream) putAck(msg ClientAck) uint64 {

	r.pendingIdCounter++
	r.pendingAck[r.pendingIdCounter] = msg

	return r.pendingIdCounter
}

func (r *RelayRelayStream) sendAck(id uint64) {

	ack, ok := r.pendingAck[id]

	if !ok {
		return
	}
	delete(r.pendingAck, id)

	r.sendAckCh <- ack
}

func (r *RelayRelayStream) processAckReads() {

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

			log.Warn().Err(err).Msg("stream recv failed")

			r.disconnect()

			continue
		}

		r.sendAck(event.MessageID)
	}
}

func (r *RelayRelayStream) processRelaySends() {

	var req RelayRelayClientEvent
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

		if req.Event != nil {

			if err := r.rpcSend(req.Event); err != nil {

				log.Warn().Err(err).Msg("stream send failed, reconnecting")

				r.disconnect()
				time.Sleep(time.Second)

				continue
			} else {
				req.Event = nil
			}
		}

		select {
		case nextEvent, ok := <-r.send:

			if !ok {
				log.Debug().Msg("send channel closed")
				return
			}

			// per-stream MessageID for the internal map / ack.
			// the relay that gets our message will send our message id back.
			_, _ = rpc.GetSetRelayMessageID(nextEvent.Event, func(id uint64) (uint64, error) {
				return r.putAck(ClientAck{
					MessageID: id,
				}), nil
			})

			req = nextEvent
		}
	}
}
