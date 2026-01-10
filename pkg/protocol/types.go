package protocol

type PacketType uint8

const (
	// Non Payload Packets
	PacketTypeNone                  PacketType = 0x00
	PacketTypeClientNeedsDisconnect PacketType = 0x01 // Client needs to disconnect

	// Debug Packets
	PacketTypeDebugHello PacketType = 0x90 // Debug: Hello
	PacketTypeDebugAny   PacketType = 0x91 // Debug: Any
)

// IsValidPacketType checks if the packet type is valid
// It returns true if the packet type is valid, false otherwise
func IsValidPacketType(packetType PacketType) bool {
	return packetType != PacketTypeNone
}
