package api

import (
	"crypto/rand"
	"tsunagi/src/client"
	"tsunagi/src/database"
	"tsunagi/src/rpc"

	"github.com/flynn/noise"
	"github.com/minnowo/tsunagi/mod/tcrypto"
	"github.com/rs/zerolog/log"
)

type TsunagiBase struct {
	DB database.DB

	// ClientClient *client.ClientRelayClient

	// always only talking to relays for now, clients have to talk to us.
	RelayClient *client.RelayRelayClient

	MacKey [tcrypto.MacKeySize]byte

	ClientConns *ClientConnManager

	RelayIdentity noise.DHKey
}

func (this *TsunagiBase) Init(db database.DB) {

	var err error

	this.RelayIdentity, err = tcrypto.GenerateNoiseKeypair()

	if err != nil {
		log.Panic().Err(err).Msg("could not generate relay identity")
	}

	this.DB = db
	this.ClientConns = NewClientConnManager()
	this.RelayClient = client.NewRelayRelayClient(this.RelayIdentity, 64, 64)
	rand.Read(this.MacKey[:])
	go this.processAckMessages()
}

func (this *TsunagiBase) processAckMessages() {

	ackStream := this.RelayClient.ReadAck()

	for {
		select {
		case ack, ok := <-ackStream:

			if !ok {
				return
			}

			this.ClientConns.PutRelayMsg(ack.ClientID, &rpc.RelayEvent{
				Body: &rpc.RelayEvent_RelayAck{
					RelayAck: &rpc.RelayAck{
						MessageID: ack.MessageID,
					},
				},
			})
		}
	}
}
