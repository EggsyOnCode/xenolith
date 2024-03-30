package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"

	"github.com/EggsyOnCode/xenolith/core_types"
)

// the inteface that every kind of Hashing algorithm should implement
type Hasher[T any] interface {
	Hash(T) core_types.Hash
}

//Concrete Implementation/ Strategies for Hasher

type BlockHasher struct{}

// sha256 implementatin has been used
// since the type itself is never used in  the implementation, we can use a receiver of type BlockHaser
func (BlockHasher) Hash(b *Block) core_types.Hash {
	//hash the block
	buf := &bytes.Buffer{}
	//buf is the io.Writer in which the encoded data will be written to
	enc := gob.NewEncoder(buf)
	err := enc.Encode(b.Header)
	if err != nil {
		panic(err)
	}

	h := sha256.Sum256(buf.Bytes())
	return core_types.Hash(h)
}
