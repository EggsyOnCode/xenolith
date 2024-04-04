package core

import (
	"fmt"

	"github.com/EggsyOnCode/xenolith/core_types"
	"github.com/EggsyOnCode/xenolith/crypto_lib"
)

type Transaction struct {
	Data      []byte
	From      *crypto_lib.PublicKey
	Signature *crypto_lib.Signature

	//caches the hash of the tx
	hash core_types.Hash
}

func NewTransaction(data []byte) *Transaction {
	return &Transaction{
		Data: data,
	}
}

// making the Tx hasher implementation generic
func (t *Transaction) Hash(h Hasher[*Transaction]) core_types.Hash {
	if t.hash.IsZero() {
		t.hash = h.Hash(t)
	}

	return t.hash
}

func (t *Transaction) Sign(priv *crypto_lib.PrivateKey) error {
	sig, err := priv.Sign(t.Data)
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
	// return t.Signature.Verify(t.Data, t.From), fmt.Errorf("invalid signature")
	return t.Signature.Verify(t.Data, t.From), nil
}
