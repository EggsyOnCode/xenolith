package core

import (
	"encoding/gob"
	"fmt"
	"math/rand"

	"github.com/EggsyOnCode/xenolith/core_types"
	"github.com/EggsyOnCode/xenolith/crypto_lib"
)

type CollectionTx struct {
	Fee      uint64
	MetaData []byte
	Quantity uint16
}

type MintTx struct {
	Fee             uint64
	MetaData        []byte
	CollectionOwner crypto_lib.PublicKey
	Signature       crypto_lib.Signature
	Collection      core_types.Hash
	NFT             core_types.Hash
}

type Transaction struct {
	//only used for native nft logic
	TxInner any
	//any arbitrary data for the VM
	Data []byte
	From crypto_lib.PublicKey
	To   crypto_lib.PublicKey
	//value of the native token being transferred
	Value     uint64
	Signature *crypto_lib.Signature
	timeStamp int64
	Nonce     int64

	//caches the hash of the tx
	hash core_types.Hash
}

func NewTransaction(data []byte) *Transaction {
	return &Transaction{
		Data:  data,
		Nonce: rand.Int63n(1000000),
	}
}


func (tx *Transaction) Revert() error {
	temp := tx.From
	tx.From = tx.To
	tx.To =  temp
	return nil
}

// making the Tx hasher implementation generic
func (t *Transaction) Hash(h Hasher[*Transaction]) core_types.Hash {
	if t.hash.IsZero() {
		t.hash = h.Hash(t)
	}

	return t.hash
}

func (t *Transaction) Sign(priv *crypto_lib.PrivateKey) error {
	// we sign the hash of the tx with our private key
	// hash := t.Hash(TxHasher{})

	// we hash it from the scratch; because the tx can only be signed once
	// therefore its the right place to calculate the hash of the tx
	hash := TxHasher{}.Hash(t)

	sig, err := priv.Sign(hash.ToSlice())
	if err != nil {
		return err
	}
	t.From = priv.PublicKey()
	t.Signature = sig

	return nil
}

func (t *Transaction) Verify() (bool, error) {
	if (t.Signature == nil) || (t.From == nil) {
		return false, fmt.Errorf("Transaction not signed")
	}

	//for verification we can't use the cached hash
	hash := TxHasher{}.Hash(t)
	return t.Signature.Verify(hash.ToSlice(), t.From), nil
}

//setters and getters for the timestamp

func (t *Transaction) SetTimeStamp(timeStamp int64) {
	t.timeStamp = timeStamp
}

func (t *Transaction) TimeStamp() int64 {
	return t.timeStamp
}

func (t *Transaction) Encode(enc Encoder[*Transaction]) error {
	return enc.Encode(t)
}
func (t *Transaction) Decode(dec Decoder[*Transaction]) error {
	return dec.Decode(t)
}

func init() {
	gob.Register(&CollectionTx{})
	gob.Register(&MintTx{})
}
