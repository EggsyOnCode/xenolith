package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"io"

	"github.com/EggsyOnCode/xenolith/core_types"
)

// Header is the header of a block
// TODO: Inclusion of MerkleRoot Hash in the header
type Header struct {
	//version specifies the version of the header config; if the version changes, it has to updated here
	Version   uint32
	PrevBlock core_types.Hash
	Head      uint32
	Nonce     uint64
	//rep the unix timestamp
	Timestamp uint64
}

// Block header needs to be sent over  the network in the form of byte slice
// we need encoding and decoding functions for that
func (h *Header) EncodeBinary(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, &h.Version); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &h.PrevBlock); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &h.Head); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &h.Nonce); err != nil {
		return err
	}
	return binary.Write(w, binary.LittleEndian, &h.Timestamp)
}

func (h *Header) DecodeBinary(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &h.Version); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &h.PrevBlock); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &h.Head); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &h.Nonce); err != nil {
		return err
	}
	return binary.Read(r, binary.LittleEndian, &h.Timestamp)
}

type Block struct {
	Header       Header
	Transactions []Transaction

	//cached hash of the block (so that if someone reqs it we don;t have to hash it agin n again)
	hash core_types.Hash
}

func (b *Block) EncodeBlock(w io.Writer) error {
	if err := b.Header.EncodeBinary(w); err != nil {
		return err
	}
	for _, tx := range b.Transactions{
		if err := tx.EncodeBinary(w); err != nil {
			return err
		}
	}
	return nil
}
func (b *Block) DecodeBlock(r io.Reader) error {
	if err := b.Header.DecodeBinary(r); err != nil {
		return err
	}
	for _, tx := range b.Transactions{
		if err := tx.DecodeBinary(r); err != nil {
			return err
		}
	}
	return nil
}

//hashing the block
func (b *Block) Hash() core_types.Hash {
	buf := new(bytes.Buffer)
	if err := b.EncodeBlock(buf); err != nil {
		panic(err)
	}
	if (b.hash.IsZero()){
		//hashing hte block using sha256
		b.hash = core_types.Hash(sha256.Sum256(buf.Bytes()))
	}

	return b.hash
}