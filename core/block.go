package core

import (
	"crypto"
	"io"

	"github.com/EggsyOnCode/xenolith/core_types"
	"github.com/EggsyOnCode/xenolith/crypto_lib"
)

// Header is the header of a block
// TODO: Inclusion of MerkleRoot Hash in the header
type Header struct {
	//version specifies the version of the header config; if the version changes, it has to updated here
	Version uint32
	//hash of all hte tx data
	DataHash      core_types.Hash
	PrevBlockHash core_types.Hash
	Head          uint32
	//rep the unix timestamp
	Timestamp uint64
}

type Block struct {
	Header       *Header
	Transactions []Transaction
	//these two fields are for the validator/miner that would be proposing hte block to the network
	Validator crypto.PublicKey
	Signature *crypto_lib.Signature
	//cached hash of the block (so that if someone reqs it we don;t have to hash it agin n again)
	hash core_types.Hash
}

// implementing the Hasher interface for the Block type (Hasher[*Block])
// here we are passing block as a type to Hasher
func (b *Block) Hash(hasher Hasher[*Block]) core_types.Hash {
	if b.hash.IsZero() {
		b.hash = hasher.Hash(b)
	}

	return b.hash
}

func (b *Block) Encode(w io.Writer,enc Encoder[*Block]) error{
	return enc.Encode(w,b)
}
func (b *Block) Decode(r io.Reader,dec Decoder[*Block]) error{
	return dec.Decode(r,b)
}


