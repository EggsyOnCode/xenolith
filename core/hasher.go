package core

import (
	"crypto/sha256"
	"encoding/binary"

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

// TxHashser for sha256 implementatin has been used
func (TxHasher) Hash(tx *Transaction) core_types.Hash {
	//int64 is 8 bytes long
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(tx.Nonce))
	data := append(buf, tx.Data...)
	h := sha256.Sum256(data)
	return core_types.Hash(h)
}
