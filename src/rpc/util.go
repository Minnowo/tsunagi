package rpc

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

func GetSetClientMessageID(evt *ClientEvent, getSetFunc func(id uint64) (uint64, error)) (uint64, error) {

	switch v := evt.Body.(type) {
	case *ClientEvent_MessagePayload:

		id, err := getSetFunc(v.MessagePayload.MessageID)

		if err == nil {
			v.MessagePayload.MessageID = id
		}

		return v.MessagePayload.MessageID, err

	case *ClientEvent_NoiseHandshake:

		id, err := getSetFunc(v.NoiseHandshake.MessageID)

		if err == nil {
			v.NoiseHandshake.MessageID = id
		}

		return v.NoiseHandshake.MessageID, err

	}

	log.Panic().Str("type", fmt.Sprintf("%T", evt.Body)).Msg("unknown event type")

	return 0, nil
}
func GetSetRelayMessageID(evt *RelayEvent, getSetFunc func(id uint64) (uint64, error)) (uint64, error) {

	switch v := evt.Body.(type) {
	case *RelayEvent_MessagePayload:

		id, err := getSetFunc(v.MessagePayload.MessageID)

		if err == nil {
			v.MessagePayload.MessageID = id
		}

		return v.MessagePayload.MessageID, err

	case *RelayEvent_NoiseHandshake:

		id, err := getSetFunc(v.NoiseHandshake.MessageID)

		if err == nil {
			v.NoiseHandshake.MessageID = id
		}

		return v.NoiseHandshake.MessageID, err

	case *RelayEvent_RelayAck:

		id, err := getSetFunc(v.RelayAck.MessageID)

		if err == nil {
			v.RelayAck.MessageID = id
		}

		return v.RelayAck.MessageID, err
	}

	log.Panic().Str("type", fmt.Sprintf("%T", evt.Body)).Msg("unknown event type")

	return 0, nil
}

func GetRelayMessageID(evt *RelayEvent) (uint64, bool) {

	switch v := evt.Body.(type) {
	case *RelayEvent_MessagePayload:
		return v.MessagePayload.MessageID, true
	case *RelayEvent_NoiseHandshake:
		return v.NoiseHandshake.MessageID, true
	case *RelayEvent_RelayAck:
		return v.RelayAck.MessageID, true
	}

	log.Panic().Str("type", fmt.Sprintf("%T", evt.Body)).Msg("unknown event type")

	return 0, false
}
