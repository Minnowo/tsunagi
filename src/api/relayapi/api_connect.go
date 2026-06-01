package relayapi

import (
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"google.golang.org/grpc"
)

func (this *RelayApi) Connect(req *rpc.ConnectRequest, stream grpc.ServerStreamingServer[rpc.Event]) error {

	var deviceID data.Identifier

	if err := deviceID.FromBytes(req.DeviceID); err != nil {
		return err
	}

	pipe, err := this.inbox.GetReadPipe(deviceID)

	if err != nil {
		return err
	}

	for {
		select {

		case <-stream.Context().Done():
			// Client disconnected, network lost, deadline exceeded, etc.
			return stream.Context().Err()

		case evt := <-pipe:

			msg := rpc.Event{
				CipherText: evt,
			}

			if err := stream.Send(&msg); err != nil {
				return err
			}
		}
	}
}
