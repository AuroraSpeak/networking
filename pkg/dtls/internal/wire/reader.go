package wire

import (
	"io"
)

type PacketReader struct {
	buf []byte
	pos int
}

func NewPacketReader(buf []byte) *PacketReader {
	return &PacketReader{
		buf: buf,
		pos: 0,
	}
}

func (r *PacketReader) ReadU8() (uint8, error) {
	if r.pos >= len(r.buf) {
		return 0, io.ErrUnexpectedEOF
	}
	val := r.buf[r.pos]
	r.pos++
	return val, nil
}

func (r *PacketReader) ReadU16() (uint16, error) {
	if r.pos+2 > len(r.buf) {
		return 0, io.ErrUnexpectedEOF
	}
	val := ReadU16(r.buf[r.pos:])
	r.pos += 2
	return val, nil
}

func (r *PacketReader) ReadU24() (uint32, error) {
	if r.pos+3 > len(r.buf) {
		return 0, io.ErrUnexpectedEOF
	}
	val := ReadU24(r.buf[r.pos:])
	r.pos += 3
	return val, nil
}

func (r *PacketReader) ReadU32() (uint32, error) {
	if r.pos+4 > len(r.buf) {
		return 0, io.ErrUnexpectedEOF
	}
	val := ReadU32(r.buf[r.pos:])
	r.pos += 4
	return val, nil
}

func (r *PacketReader) ReadU64() (uint64, error) {
	if r.pos+8 > len(r.buf) {
		return 0, io.ErrUnexpectedEOF
	}
	val := ReadU64(r.buf[r.pos:])
	r.pos += 8
	return val, nil
}

func (r *PacketReader) ReadBytes(n int) ([]byte, error) {
	if r.pos+n > len(r.buf) {
		return nil, io.ErrUnexpectedEOF
	}
	val := r.buf[r.pos : r.pos+n]
	r.pos += n
	return val, nil
}

func (r *PacketReader) Skip(n int) error {
	if r.pos+n > len(r.buf) {
		return io.ErrUnexpectedEOF
	}
	r.pos += n
	return nil
}

func (r *PacketReader) PeekU8() (uint8, error) {
	if r.pos+1 > len(r.buf) {
		return 0, io.ErrUnexpectedEOF
	}
	val := r.buf[r.pos]
	return val, nil
}

func (r *PacketReader) PeekU16() (uint16, error) {
	if r.pos+2 > len(r.buf) {
		return 0, io.ErrUnexpectedEOF
	}
	val := ReadU16(r.buf[r.pos:])
	return val, nil
}

func (r *PacketReader) PeekU24() (uint32, error) {
	if r.pos+3 > len(r.buf) {
		return 0, io.ErrUnexpectedEOF
	}
	val := ReadU24(r.buf[r.pos:])
	return val, nil
}

func (r *PacketReader) PeekU32() (uint32, error) {
	if r.pos+4 > len(r.buf) {
		return 0, io.ErrUnexpectedEOF
	}
	val := ReadU32(r.buf[r.pos:])
	return val, nil
}

func (r *PacketReader) PeekU64() (uint64, error) {
	if r.pos+8 > len(r.buf) {
		return 0, io.ErrUnexpectedEOF
	}
	val := ReadU64(r.buf[r.pos:])
	return val, nil
}

func (r *PacketReader) PeekBytes(n int) ([]byte, error) {
	if r.pos+n > len(r.buf) {
		return nil, io.ErrUnexpectedEOF
	}
	val := r.buf[r.pos : r.pos+n]
	return val, nil
}

func (r *PacketReader) Remaining() int {
	return len(r.buf) - r.pos
}

func (r *PacketReader) SetPos(pos int) {
	r.pos = pos
}
