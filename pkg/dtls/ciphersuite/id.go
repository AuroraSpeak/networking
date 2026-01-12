package ciphersuite

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"

	"golang.org/x/crypto/chacha20poly1305"
)

// TLS_AES_128_GCM_SHA256, TLS_AES_256_GCM_SHA384, TLS_CHACHA20_POLY1305_SHA256
type ID uint16

const (
	TLS_AES_128_GCM_SHA256       ID = 0x1301
	TLS_AES_256_GCM_SHA384       ID = 0x1302
	TLS_CHACHA20_POLY1305_SHA256 ID = 0x1303
)

type Suite struct {
	ID   ID
	Name string

	// TLS 1.3: Hash ist Teil der Suite (für HKDF, Transcript, Finished, PSK Binder)
	Hash crypto.Hash

	KeyLen int // 16 (AES-128), 32 (AES-256/ChaCha20)
	IVLen  int // 12 (TLS 1.3 fixed_iv)

	NewAEAD func(key []byte) (cipher.AEAD, error)
}

func (s Suite) HashFunc() func() crypto.Hash {
	return func() crypto.Hash { return s.Hash }
}

var Registry = map[ID]Suite{
	TLS_AES_128_GCM_SHA256: {
		ID:     TLS_AES_128_GCM_SHA256,
		Name:   "TLS_AES_128_GCM_SHA256",
		Hash:   crypto.SHA256,
		KeyLen: 16,
		IVLen:  12,
		NewAEAD: func(key []byte) (cipher.AEAD, error) {
			b, err := aes.NewCipher(key)
			if err != nil {
				return nil, err
			}
			return cipher.NewGCM(b)
		},
	},
	TLS_AES_256_GCM_SHA384: {
		ID:     TLS_AES_256_GCM_SHA384,
		Name:   "TLS_AES_256_GCM_SHA384",
		Hash:   crypto.SHA384,
		KeyLen: 32,
		IVLen:  12,
		NewAEAD: func(key []byte) (cipher.AEAD, error) {
			b, err := aes.NewCipher(key)
			if err != nil {
				return nil, err
			}
			return cipher.NewGCM(b)
		},
	},
	TLS_CHACHA20_POLY1305_SHA256: {
		ID:     TLS_CHACHA20_POLY1305_SHA256,
		Name:   "TLS_CHACHA20_POLY1305_SHA256",
		Hash:   crypto.SHA256,
		KeyLen: 32,
		IVLen:  12,
		NewAEAD: func(key []byte) (cipher.AEAD, error) {
			return chacha20poly1305.New(key)
		},
	},
}

func Get(id ID) (Suite, bool) {
	s, ok := Registry[id]
	return s, ok
}

func Negotiate(serverPref, clientOffered []ID) (Suite, bool) {
	offered := make(map[ID]struct{}, len(clientOffered))
	for _, id := range clientOffered {
		offered[id] = struct{}{}
	}
	for _, id := range serverPref {
		if _, ok := offered[id]; ok {
			s, ok2 := Get(id)
			return s, ok2
		}
	}
	return Suite{}, false
}
