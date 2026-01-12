// Package wire provides utility functions for reading and writing
// numeric values to/from byte slices in big-endian format.
// These functions are used for DTLS wire protocol encoding and decoding.
package wire

import "encoding/binary"

// PutU8 writes a uint8 value to the first byte of the given slice.
// The slice must have at least 1 byte of capacity.
//
// Example:
//
//	buf := make([]byte, 1)
//	PutU8(buf, 42)
//	// buf[0] == 42
func PutU8(b []byte, v uint8) {
	b[0] = v
}

// PutU8At writes a uint8 value to the given slice at the specified offset.
// The slice must have at least offset+1 bytes of capacity.
//
// Example:
//
//	buf := make([]byte, 5)
//	PutU8At(buf, 2, 42)
//	// buf[2] == 42
func PutU8At(b []byte, offset int, v uint8) {
	b[offset] = v
}

// PutU16 writes a uint16 value to the given slice in big-endian format.
// The slice must have at least 2 bytes of capacity.
//
// Example:
//
//	buf := make([]byte, 2)
//	PutU16(buf, 0x1234)
//	// buf[0] == 0x12, buf[1] == 0x34
func PutU16(b []byte, v uint16) {
	binary.BigEndian.PutUint16(b, v)
}

// PutU16At writes a uint16 value to the given slice at the specified offset in big-endian format.
// The slice must have at least offset+2 bytes of capacity.
//
// Example:
//
//	buf := make([]byte, 5)
//	PutU16At(buf, 1, 0x1234)
//	// buf[1] == 0x12, buf[2] == 0x34
func PutU16At(b []byte, offset int, v uint16) {
	binary.BigEndian.PutUint16(b[offset:], v)
}

// PutU24 writes a uint32 value as a 24-bit (3-byte) value to the given slice in big-endian format.
// Only the lower 24 bits of v are written. The slice must have at least 3 bytes of capacity.
//
// Example:
//
//	buf := make([]byte, 3)
//	PutU24(buf, 0x123456)
//	// buf[0] == 0x12, buf[1] == 0x34, buf[2] == 0x56
func PutU24(b []byte, v uint32) {
	b[0] = byte(v >> 16)
	b[1] = byte(v >> 8)
	b[2] = byte(v)
}

// PutU24At writes a uint32 value as a 24-bit (3-byte) value to the given slice
// at the specified offset in big-endian format.
// Only the lower 24 bits of v are written. The slice must have at least offset+3 bytes of capacity.
//
// Example:
//
//	buf := make([]byte, 6)
//	PutU24At(buf, 1, 0x123456)
//	// buf[1] == 0x12, buf[2] == 0x34, buf[3] == 0x56
func PutU24At(b []byte, offset int, v uint32) {
	b[offset] = byte(v >> 16)
	b[offset+1] = byte(v >> 8)
	b[offset+2] = byte(v)
}

// PutU32 writes a uint32 value to the given slice in big-endian format.
// The slice must have at least 4 bytes of capacity.
//
// Example:
//
//	buf := make([]byte, 4)
//	PutU32(buf, 0x12345678)
//	// buf[0] == 0x12, buf[1] == 0x34, buf[2] == 0x56, buf[3] == 0x78
func PutU32(b []byte, v uint32) {
	binary.BigEndian.PutUint32(b, v)
}

// PutU32At writes a uint32 value to the given slice at the specified offset in big-endian format.
// The slice must have at least offset+4 bytes of capacity.
//
// Example:
//
//	buf := make([]byte, 8)
//	PutU32At(buf, 2, 0x12345678)
//	// buf[2] == 0x12, buf[3] == 0x34, buf[4] == 0x56, buf[5] == 0x78
func PutU32At(b []byte, offset int, v uint32) {
	binary.BigEndian.PutUint32(b[offset:], v)
}

// PutU64 writes a uint64 value to the given slice in big-endian format.
// The slice must have at least 8 bytes of capacity.
//
// Example:
//
//	buf := make([]byte, 8)
//	PutU64(buf, 0x1234567890ABCDEF)
//	// buf[0] == 0x12, buf[1] == 0x34, ..., buf[7] == 0xEF
func PutU64(b []byte, v uint64) {
	binary.BigEndian.PutUint64(b, v)
}

// PutU64At writes a uint64 value to the given slice at the specified offset in big-endian format.
// The slice must have at least offset+8 bytes of capacity.
//
// Example:
//
//	buf := make([]byte, 12)
//	PutU64At(buf, 2, 0x1234567890ABCDEF)
//	// buf[2] == 0x12, buf[3] == 0x34, ..., buf[9] == 0xEF
func PutU64At(b []byte, offset int, v uint64) {
	binary.BigEndian.PutUint64(b[offset:], v)
}

// ReadU8 reads a uint8 value from the first byte of the given slice.
// The slice must have at least 1 byte.
//
// Example:
//
//	buf := []byte{42}
//	value := ReadU8(buf)
//	// value == 42
func ReadU8(b []byte) uint8 { return b[0] }

// ReadU16 reads a uint16 value from the given slice in big-endian format.
// The slice must have at least 2 bytes.
//
// Example:
//
//	buf := []byte{0x12, 0x34}
//	value := ReadU16(buf)
//	// value == 0x1234
func ReadU16(b []byte) uint16 { return binary.BigEndian.Uint16(b) }

// ReadU24 reads a 24-bit (3-byte) value from the given slice in big-endian format
// and returns it as a uint32. The slice must have at least 3 bytes.
//
// Example:
//
//	buf := []byte{0x12, 0x34, 0x56}
//	value := ReadU24(buf)
//	// value == 0x123456
func ReadU24(b []byte) uint32 {
	return uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])
}

// ReadU32 reads a uint32 value from the given slice in big-endian format.
// The slice must have at least 4 bytes.
//
// Example:
//
//	buf := []byte{0x12, 0x34, 0x56, 0x78}
//	value := ReadU32(buf)
//	// value == 0x12345678
func ReadU32(b []byte) uint32 { return binary.BigEndian.Uint32(b) }

// ReadU64 reads a uint64 value from the given slice in big-endian format.
// The slice must have at least 8 bytes.
//
// Example:
//
//	buf := []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF}
//	value := ReadU64(buf)
//	// value == 0x1234567890ABCDEF
func ReadU64(b []byte) uint64 { return binary.BigEndian.Uint64(b) }

// AppendU16 appends a uint16 value to the destination slice in big-endian format.
// Returns the updated slice.
//
// Example:
//
//	buf := []byte{0x01, 0x02}
//	buf = AppendU16(buf, 0x1234)
//	// buf == []byte{0x01, 0x02, 0x12, 0x34}
func AppendU16(dst []byte, v uint16) []byte {
	return append(dst, byte(v>>8), byte(v))
}

// AppendU24 appends a uint32 value as a 24-bit (3-byte) value to the destination slice
// in big-endian format. Only the lower 24 bits of v are appended.
// Returns the updated slice.
//
// Example:
//
//	buf := []byte{0x01, 0x02}
//	buf = AppendU24(buf, 0x123456)
//	// buf == []byte{0x01, 0x02, 0x12, 0x34, 0x56}
func AppendU24(dst []byte, v uint32) []byte {
	return append(dst, byte(v>>16), byte(v>>8), byte(v))
}
