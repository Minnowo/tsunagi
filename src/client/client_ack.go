package client

import (
	"tsunagi/src/data"
)

type ClientAck struct {
	ClientID  data.Identifier
	MessageID uint64
}
