package relayapi

import (
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"google.golang.org/grpc"
)

type RelayApi struct {
	rpc.UnimplementedTsunagiServer
	inbox Inbox
}

func (this *RelayApi) Init() {

	this.inbox = Inbox{
		inbox: map[data.Identifier]Box{},
	}
}

func (this *RelayApi) Register(r *grpc.Server) {
	rpc.RegisterTsunagiServer(r, this)
}
