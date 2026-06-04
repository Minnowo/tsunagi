package relayapi

import (
	"tsunagi/src/client"
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type RelayApi struct {
	rpc.UnimplementedTsunagiServer
	rpc.UnimplementedAuthServer
	inbox       Inbox
	relayClient *client.RelayRelayClient
}

func (this *RelayApi) ValidAuth(md metadata.MD) bool {

	auth := md.Get("authorization")

	if len(auth) == 0 {
		return false
	}

	// TODO: check this

	return true
}
func (this *RelayApi) Init() {

	this.inbox = Inbox{
		inbox: map[data.Identifier]Box{},
	}

	this.relayClient = client.NewRelayRelayClient(50)
}

func (this *RelayApi) Register(r *grpc.Server) {
	rpc.RegisterAuthServer(r, this)
	rpc.RegisterTsunagiServer(r, this)
}
