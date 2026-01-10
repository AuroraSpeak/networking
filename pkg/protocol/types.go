package protocol

type PacketType uint8

type PacketTypeMapping struct {
	PacketType PacketType
	String     string
}

const (
	// Non Payload Packets
	PacketTypeNone                  PacketType = 0x00
	PacketTypeClientNeedsDisconnect PacketType = 0x01 // Client needs to disconnect

	// Debug Packets
	PacketTypeDebugHello PacketType = 0x90 // Debug: Hello
	PacketTypeDebugAny   PacketType = 0x91 // Debug: Any
)

var (
	PacketTypeMap = []PacketTypeMapping{
		{PacketType: PacketTypeNone, String: "None"},
		{PacketType: PacketTypeClientNeedsDisconnect, String: "ClientNeedsDisconnect"},
		{PacketType: PacketTypeDebugHello, String: "DebugHello"},
		{PacketType: PacketTypeDebugAny, String: "DebugAny"},
	}

	PacketTypeMapType = func() map[PacketType]string {
		m := make(map[PacketType]string)
		for _, mapping := range PacketTypeMap {
			m[mapping.PacketType] = mapping.String
		}
		return m
	}()
)

// IsValidPacketType checks if the packet type is valid
// It returns true if the packet type is valid, false otherwise
func IsValidPacketType(packetType PacketType) bool {
	return packetType != PacketTypeNone
}
