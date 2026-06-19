package relayapi

import (
	"encoding/base64"
	"tsunagi/src/api"
	"tsunagi/src/rpc"

	"github.com/minnowo/tsunagi/mod/tcrypto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type RelayApi struct {
	rpc.UnimplementedTsunagiServer
	rpc.UnimplementedAuthServer
	api.TsunagiBase
}

func (this *RelayApi) GetAuthIdentity(md metadata.MD) ([]byte, error) {

	auth := md.Get("authorization")

	if len(auth) != 1 {
		return nil, tcrypto.ErrMacMismatch
	}

	token, err := base64.StdEncoding.DecodeString(auth[0])

	if err != nil {
		return nil, err
	}

	return tcrypto.ParseAuthToken(token, this.MacKey[:])
}

func (this *RelayApi) Register(r *grpc.Server) {
	rpc.RegisterAuthServer(r, this)
	rpc.RegisterTsunagiServer(r, this)
}
