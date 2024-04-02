package core

import (
	"bytes"
	"encoding/gob"
	"fmt"
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
	Height          uint32
	//rep the unix timestamp
	Timestamp uint64
}

type Block struct {
	Header       *Header
	Transactions []Transaction
	//these two fields are for the validator/miner that would be proposing hte block to the network
	Validator *crypto_lib.PublicKey
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

func (b *Block) Encode(w io.Writer, enc Encoder[*Block]) error {
	return enc.Encode(w, b)
}
func (b *Block) Decode(r io.Reader, dec Decoder[*Block]) error {
	return dec.Decode(r, b)
}

func (b *Block) Sign(priv *crypto_lib.PrivateKey) error {
	sig, err := priv.Sign(b.HeaderData())
	if err != nil {
		return err
	}
	b.Validator = priv.PublicKey()
	b.Signature = sig

	return nil
}

func (b *Block) Verify() (bool, error) {
	if (b.Signature == nil) || (b.Validator == nil) {
		return false, fmt.Errorf("Block not signed")
	}
	return b.Signature.Verify(b.HeaderData(), b.Validator), fmt.Errorf("invalid signature")
}

func (b *Block) HeaderData() []byte {
	buf := &bytes.Buffer{}
	//buf is the io.Writer in which the encoded data will be written to
	enc := gob.NewEncoder(buf)
	fmt.Printf("Header Height: %v\n", b.Header.Height)
	if err := enc.Encode(b.Header); err != nil {
		panic(err)
	}

	return buf.Bytes()
}