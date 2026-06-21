package client

import (
	"tsunagi/src/rpc"

	"github.com/flynn/noise"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TsunagiConn struct {
	Addr     string
	Conn     *grpc.ClientConn
	Tsu      rpc.TsunagiClient
	Auth     rpc.AuthClient
	Identity noise.DHKey
}

func NewTsunagiConn(addr string, identity noise.DHKey) TsunagiConn {
	return TsunagiConn{Addr: addr, Identity: identity}
}

func (r *TsunagiConn) Close() error {

	if r.Conn != nil {

		if err := r.Conn.Close(); err != nil {
			return err
		}

		r.Conn = nil
	}
	return nil
}

func (r *TsunagiConn) Connect() error {

	if r.Conn == nil {

		conn, err := grpc.NewClient(
			r.Addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		if err != nil {
			return err
		}

		r.Conn = conn
		r.Auth = rpc.NewAuthClient(r.Conn)
		r.Tsu = rpc.NewTsunagiClient(r.Conn)
	}

	return nil
}
