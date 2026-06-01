package client

import (
	"context"
	"encoding/hex"
	"sync"
	"time"
	"tsunagi/src/rpc"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type sendRequest struct {
	event *rpc.Event
	resp  chan error      // channel to report success/failure
	ctx   context.Context // deadline
}

func (r *sendRequest) PutErr(err error) {
	if r.resp != nil {
		select {
		default:
			return
		case r.resp <- err:
			return
		}
	}
}
func (r *sendRequest) IsDeadline() bool {
	if r.ctx != nil {
		select {
		default:
			return false
		case <-r.ctx.Done():
			return true
		}
	}
	return false
}

type rpcClient struct {
	addr string

	conn *grpc.ClientConn
	tsu  rpc.TsunagiClient
	auth rpc.AuthClient

	smu    sync.Mutex
	stream grpc.BidiStreamingClient[rpc.Event, rpc.Event]

	send chan sendRequest
	read chan *rpc.Event
	exit chan bool
}

func (r *rpcClient) disconnect() {

	r.smu.Lock()
	defer r.smu.Unlock()

	r.stream.CloseSend()
	r.stream = nil
}

func (r *rpcClient) connect() error {

	r.smu.Lock()
	defer r.smu.Unlock()

	if r.stream != nil {
		return nil
	}

	if r.conn == nil {

		conn, err := grpc.NewClient(
			r.addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		if err != nil {
			return err
		}

		r.conn = conn
		r.auth = rpc.NewAuthClient(r.conn)
		r.tsu = rpc.NewTsunagiClient(r.conn)
	}

	ctx := context.Background()

	challenge, err := r.auth.GetChallenge(ctx, &rpc.AuthRequest{})

	if err != nil {
		return err
	}

	log.Info().Hex("nonce", challenge.Nonce).Msg("got challenge")

	authToken, err := r.auth.ProveChallenge(ctx, &rpc.AuthProof{
		Signature: challenge.Nonce,
	})

	if err != nil {
		return err
	}

	log.Info().Hex("token", authToken.Token).Msg("proved challenge")

	authCtx := metadata.NewOutgoingContext(
		context.Background(),
		metadata.Pairs("authorization", "Bearer "+hex.EncodeToString(authToken.GetToken())),
	)

	stream, err := r.tsu.Connect(authCtx)

	if err != nil {
		return err
	}

	r.stream = stream

	return nil
}

func (r *rpcClient) processSends() {

	var req *sendRequest
	for {

		if req != nil && req.IsDeadline() {
			req.PutErr(ErrReqTimeout)
			req = nil
		}

		select {
		case <-r.exit:
			return
		default:
			break
		}

		if r.stream == nil {

			if err := r.connect(); err != nil {

				log.Warn().Err(err).Msg("failed to connect, retrying in 1s")
				time.Sleep(time.Second)

				continue
			}
		}

		if req != nil {

			if err := r.stream.Send(req.event); err != nil {

				log.Warn().Err(err).Msg("stream send failed, reconnecting")

				r.disconnect()
				time.Sleep(time.Second)

				continue
			} else {
				req.PutErr(nil)
				req = nil
			}
		}

		select {
		case nextEvent, ok := <-r.send:

			if !ok {
				log.Debug().Msg("send channel closed")
				return
			}

			req = &nextEvent
		}
	}
}

func (r *rpcClient) processReads() {

	for {

		select {
		case <-r.exit:
			return
		default:
			break
		}

		if r.stream == nil {

			if err := r.connect(); err != nil {

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

		r.read <- event
	}
}
