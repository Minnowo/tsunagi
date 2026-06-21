package api

import (
	"crypto/rand"
	"tsunagi/src/client"
	"tsunagi/src/database"
	"tsunagi/src/rpc"

	"github.com/minnowo/tsunagi/mod/tcrypto"
)

type TsunagiBase struct {
	DB database.DB

	// ClientClient *client.ClientRelayClient

	// always only talking to relays for now, clients have to talk to us.
	RelayClient *client.RelayRelayClient

	MacKey [tcrypto.MacKeySize]byte

	ClientConns *ClientConnManager
}

func (this *TsunagiBase) Init(db database.DB) {

	this.DB = db
	this.ClientConns = NewClientConnManager()
	this.RelayClient = client.NewRelayRelayClient(64, 64)
	rand.Read(this.MacKey[:])
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
