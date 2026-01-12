package header

import "unsafe"

type Unsigned interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type IntPacker[T Unsigned] struct {
	v T
}

func NewIntPacker[T Unsigned]() *IntPacker[T] { return &IntPacker[T]{} }

func (p *IntPacker[T]) Value() T { return p.v }

// LSB: bit 0 ist das niederwertigste Bit (rechts)
func (p *IntPacker[T]) SetBitLSB(pos uint, on bool) {
	mask := T(1) << pos
	if on {
		p.v |= mask
	} else {
		p.v &^= mask
	}
}

func (p *IntPacker[T]) SetFieldLSB(offset, width uint, value T) {
	mask := ((T(1) << width) - 1) << offset
	p.v &^= mask
	p.v |= (value << offset) & mask
}

// MSB: bit 0 ist das höchstwertige Bit (links) – passt oft besser zu Specs
func (p *IntPacker[T]) SetFieldMSB(msbOffset, width uint, value T) {
	nbits := uint(unsafe.Sizeof(p.v) * 8)
	lsbOffset := nbits - (msbOffset + width) // MSB->LSB umrechnen
	p.SetFieldLSB(lsbOffset, width, value)
}

// Integer -> Bytes in Big Endian (höchstwertiges Byte zuerst)
func (p *IntPacker[T]) BytesBE() []byte {
	n := int(unsafe.Sizeof(p.v))
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		shift := uint((n - 1 - i) * 8)
		out[i] = byte(p.v >> shift)
	}
	return out
}
