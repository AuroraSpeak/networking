package records

import (
	"errors"
)

// OptFunc applies an option to Opts.
type OptFunc func(*Opts)

// DTLS_VERSION is kept as a project-level knob (not an on-wire Unified Header field in DTLS 1.3).
type DTLS_VERSION int

const (
	V13 DTLS_VERSION = iota
)

type Opts struct {
	// Defaults
	version DTLS_VERSION // defaults to V13

	// On-wire relevant for DTLS 1.3 Unified Header
	epoch uint16 // least significant 2 bytes of epoch (DTLS 1.3)

	// Truncated sequence number carried in the Unified Header:
	// - if S=0 => 1 byte (low 8 bits)
	// - if S=1 => 2 bytes (low 16 bits)
	truncatedSeq uint16

	// Optional fields
	cid []byte

	length        uint16
	lengthPresent bool

	// Derived / forced
	seqLenBytes uint8 // 1 or 2; if 0, auto-derive from truncatedSeq
}

type UnifiedHeader struct {
	// Parsed/constructed values (not all exist on-wire simultaneously; presence depends on flags).
	Version DTLS_VERSION // project knob

	Epoch uint16

	// If CIDPresent is true, CID contains the negotiated-length CID bytes.
	CIDPresent bool
	CID        []byte

	// Truncated sequence number and its on-wire length (1 or 2 bytes).
	SeqLenBytes  uint8
	TruncatedSeq uint16

	// If LengthPresent is true, Length is present on-wire (2 bytes).
	LengthPresent bool
	Length        uint16
}

func NewHeaderFromOpts(opts ...OptFunc) (*UnifiedHeader, error) {
	o := defaultOpts()
	for _, fn := range opts {
		fn(&o)
	}
	if err := o.Validate(); err != nil {
		return nil, err
	}

	seqLen := o.seqLenBytes
	if seqLen == 0 {
		// Auto-derive: fit in 1 byte => S=0; otherwise S=1.
		if o.truncatedSeq <= 0xFF {
			seqLen = 1
		} else {
			seqLen = 2
		}
	}

	h := &UnifiedHeader{
		Version:       o.version,
		Epoch:         o.epoch,
		SeqLenBytes:   seqLen,
		TruncatedSeq:  o.truncatedSeq,
		LengthPresent: o.lengthPresent,
		Length:        o.length,
	}

	if len(o.cid) > 0 {
		h.CIDPresent = true
		// Copy to avoid accidental mutation by caller.
		h.CID = append([]byte(nil), o.cid...)
	}

	return h, nil
}

// --- Options ---

func WithVersion(version DTLS_VERSION) OptFunc {
	return func(o *Opts) { o.version = version }
}

func WithEpoch(epoch uint16) OptFunc {
	return func(o *Opts) { o.epoch = epoch }
}

// WithTruncatedSeq sets the truncated sequence number as carried on-wire (low 8 or 16 bits).
// Note: 0 is valid; do NOT treat it as "unset".
func WithTruncatedSeq(seq uint16) OptFunc {
	return func(o *Opts) { o.truncatedSeq = seq }
}

// WithSeqLenBytes forces the on-wire truncated sequence length (1 or 2 bytes).
// Use this if you want explicit control over S-bit behavior.
func WithSeqLenBytes(n uint8) OptFunc {
	return func(o *Opts) { o.seqLenBytes = n }
}

// WithCID sets the CID bytes (variable length, negotiated).
// Note: CID presence must be determined by len(cid)>0, NOT by numeric value.
func WithCID(cid []byte) OptFunc {
	return func(o *Opts) { o.cid = append([]byte(nil), cid...) }
}

// WithLength marks the length field as present (L=1) and sets its value.
func WithLength(length uint16) OptFunc {
	return func(o *Opts) {
		o.length = length
		o.lengthPresent = true
	}
}

// WithoutLength explicitly clears the length field presence (L=0).
func WithoutLength() OptFunc {
	return func(o *Opts) {
		o.length = 0
		o.lengthPresent = false
	}
}

// --- Validation ---

func (o *Opts) Validate() error {
	// version: only V13 supported for now
	if o.version != V13 {
		return errors.New("unsupported DTLS version (only V13 supported)")
	}

	// seqLenBytes: if set explicitly, must be 1 or 2
	if o.seqLenBytes != 0 && o.seqLenBytes != 1 && o.seqLenBytes != 2 {
		return errors.New("seqLenBytes must be 1 or 2")
	}

	// If forced to 1 byte, seq must fit
	if o.seqLenBytes == 1 && o.truncatedSeq > 0xFF {
		return errors.New("truncatedSeq does not fit in 1 byte (seqLenBytes=1)")
	}

	// Epoch is allowed to be 0 (first epoch).
	// Length is allowed to be 0 even if present; presence is controlled by lengthPresent.

	return nil
}

func defaultOpts() Opts {
	return Opts{
		version: V13,
	}
}
