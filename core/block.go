package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/EggsyOnCode/xenolith/core_types"
	"github.com/EggsyOnCode/xenolith/crypto_lib"
)

// Header is the header of a block
// TODO: Inclusion of MerkleRoot Hash in the header
type Header struct {
	//version specifies the version of the header config; if the version changes, it has to updated here
	Version uint32
	//hash of all hte tx data
	//i.e should be merkle root hash of all the tx in the block
	DataHash      core_types.Hash
	PrevBlockHash core_types.Hash
	Height        uint32
	//rep the unix timestamp
	Timestamp uint64
}

func (h *Header) Bytes() []byte {
	buf := &bytes.Buffer{}
	//buf is the io.Writer in which the encoded data will be written to
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(h); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

type Block struct {
	Header       *Header
	Transactions []*Transaction
	//these two fields are for the validator/miner that would be proposing hte block to the network
	Validator crypto_lib.PublicKey
	Signature *crypto_lib.Signature
	//cached hash of the block (so that if someone reqs it we don;t have to hash it agin n again)
	hash       core_types.Hash
	NextBlocks []*Block
	PrevBlock  *Block
}

func NewBlock(h *Header, txx []*Transaction) *Block {
	return &Block{
		Header:       h,
		Transactions: txx,
	}
}

func NewBlockFromPrevHeader(prevHeader *Header, tx []*Transaction) (*Block, error) {
	datahash, err := CalculateDataHash(tx)
	if err != nil {
		return nil, err
	}

	head := &Header{
		Version:       prevHeader.Version,
		PrevBlockHash: BlockHasher{}.Hash(prevHeader),
		DataHash:      datahash,
		Timestamp:     uint64(time.Now().UnixNano()),
		Height:        prevHeader.Height + 1,
	}

	return NewBlock(head, tx), nil
}

// implementing the Hasher interface for the Block type (Hasher[*Block])
// here we are passing block as a type to Hasher
func (b *Block) Hash(hasher Hasher[*Header]) core_types.Hash {
	if b.hash.IsZero() {
		b.hash = hasher.Hash(b.Header)
	}

	return b.hash
}

func (b *Block) Encode(enc Encoder[*Block]) error {
	return enc.Encode(b)
}
func (b *Block) Decode(dec Decoder[*Block]) error {
	return dec.Decode(b)
}

func (b *Block) Sign(priv *crypto_lib.PrivateKey) error {
	sig, err := priv.Sign(b.Header.Bytes())
	if err != nil {
		return err
	}
	b.Validator = priv.PublicKey()
	b.Signature = sig

	return nil
}

// block data hash being calculated when the block  is verified
func (b *Block) Verify() error {
	if (b.Signature == nil) || (b.Validator == nil) {
		return fmt.Errorf("Block not signed")
	}
	fmt.Printf("validator is before chaning %+v\n", b.Validator)
	if !b.Signature.Verify(b.Header.Bytes(), b.Validator) {
		return fmt.Errorf("invalid signature")
	}

	for _, tx := range b.Transactions {
		if ans, err := tx.Verify(); !ans {
			return fmt.Errorf("invalid transaction with ans %v : %v", ans, err)
		}
	}

	//also need to compare datahash of the slice of tx with the datahash of the proposed block
	dataHash, _ := CalculateDataHash(b.Transactions)
	if dataHash != b.Header.DataHash {
		return fmt.Errorf("invalid data hash")
	}

	return nil
}

func (b *Block) AddTx(tx *Transaction) error {
	ans, _ := tx.Verify()
	if !ans {
		return fmt.Errorf("Transaction not signed")
	}
	b.Transactions = append(b.Transactions, tx)
	hash, err := CalculateDataHash(b.Transactions)
	if err != nil {
		fmt.Printf("Error calculating data hash: %v", err)
		return err
	}
	b.Header.DataHash = hash
	return nil
}

func CalculateDataHash(tx []*Transaction) (hash core_types.Hash, err error) {
	//hashing all the tx data
	buf := &bytes.Buffer{}
	for _, tx := range tx {
		if err = tx.Encode(NewGobTxEncoder(buf)); err != nil {
			return
		}
	}

	hash = sha256.Sum256(buf.Bytes())

	return
}

// func (b *Block) HeaderData() []byte {
// 	buf := &bytes.Buffer{}
// 	//buf is the io.Writer in which the encoded data will be written to
// 	enc := gob.NewEncoder(buf)
// 	if err := enc.Encode(b.Header); err != nil {
// 		panic(err)
// 	}

// 	return buf.Bytes()
// }
