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
func (BlockHasher) Hash(header *Header) core_types.Hash {
	h := sha256.Sum256(header.Bytes())
	return core_types.Hash(h)
}

type TxHasher struct{}

// Data any
// value 8
// / from 32
// to 32
// nonce 8
// whole tx struct will be hashed
func (TxHasher) Hash(tx *Transaction) core_types.Hash {
	buf := new(bytes.Buffer)

	if err := gob.NewEncoder(buf).Encode(*tx); err != nil {
		panic(err)
	}
	h := sha256.Sum256(buf.Bytes())
	return core_types.Hash(h)
}
