package core_types

import (
	"encoding/hex"
	"math/rand"
	"time"
)

type Hash [32]uint8

// HashFromBytes converts a byte slice to a Hash
func HashFromBytes(b []byte) Hash {
	if len(b) != 32 {
		panic("Binary Hash must be 32 bytes long")
	}
	var v [32]uint8
	for i := 0; i < 32; i++ {
		v[i] = b[i]
	}

	return Hash(v)
}

func (h *Hash) IsZero() bool {
	for i := 0; i < 32; i++ {
		if h[i] != 0 {
			return false
		}
	}
	return true
}

func (h *Hash) ToSlice() []byte {
	buf := make([]byte, 32)
	for i := 0; i < 32; i++ {
		buf[i] = h[i]
	}
	return buf
}

// hash is implementing the String interface meaning all its outputs will now be typecasted to  hex string
func (h Hash) String() string {
	return hex.EncodeToString(h.ToSlice())
}

// utils func
func GenerateRandomHash(length int) Hash {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return (HashFromBytes(b))
}
