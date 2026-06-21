package httpapi

// EventType is the JSON "type" field used in WebSocket messages.
type EventType string

const (
	EventNoiseHandshake EventType = "noise_handshake"
	EventMessagePayload EventType = "message_payload"
	EventRelayAck       EventType = "relay_ack"
)
