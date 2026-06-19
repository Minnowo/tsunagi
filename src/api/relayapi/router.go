package relayapi

import (
	"crypto/rand"
	"encoding/base64"
	"tsunagi/src/client"
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"github.com/minnowo/tsunagi/mod/tcrypto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type RelayApi struct {
	rpc.UnimplementedTsunagiServer
	rpc.UnimplementedAuthServer
	inbox       Inbox
	relayClient *client.RelayRelayClient
	macKey      []byte
}

func (this *RelayApi) GetAuthIdentity(md metadata.MD) ([]byte, error) {

	auth := md.Get("authorization")

	if len(auth) != 1 {
		return nil, tcrypto.ErrMacMismatch
	}

	token, err := base64.StdEncoding.DecodeString(auth[0])

	if err != nil { return nil, err }

	return tcrypto.ParseAuthToken(token, this.macKey)
}

func (this *RelayApi) Init() {

	this.inbox = Inbox{
		inbox: map[data.Identifier]Box{},
	}

	this.relayClient = client.NewRelayRelayClient(50)
	this.macKey = make([]byte, tcrypto.MacKeySize)
	rand.Read(this.macKey)
}

func (this *RelayApi) Register(r *grpc.Server) {
	rpc.RegisterAuthServer(r, this)
	rpc.RegisterTsunagiServer(r, this)
}
