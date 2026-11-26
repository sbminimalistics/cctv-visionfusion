package edge_device

import "context"

type BasicParsedMessage interface {
	GetEventType() string
}

type PacketParser interface {
	Parse(ctx context.Context, packet []byte) (BasicParsedMessage, error)
}
