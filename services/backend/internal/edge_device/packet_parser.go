package edge_device

type BasicParsedMessage interface {
	GetEventType() string
}

type PacketParser interface {
	Parse(packet []byte) (BasicParsedMessage, error)
}
