package core

import (
	"fmt"

	"github.com/EggsyOnCode/xenolith/crypto_lib"
)

type Transaction struct {
	Data      []byte
	PublicKey *crypto_lib.PublicKey
	Signature *crypto_lib.Signature
}

func (t *Transaction) Sign(priv *crypto_lib.PrivateKey) error {
	sig, err := priv.Sign(t.Data)
	if err != nil {
		return err
	}
	t.PublicKey = priv.PublicKey()
	t.Signature = sig

	return nil
}

func (t *Transaction) Verify() (bool, error) {
	if (t.Signature == nil) || (t.PublicKey == nil) {
		return false, fmt.Errorf("Transaction not signed")
	}
	return t.Signature.Verify(t.Data, t.PublicKey), fmt.Errorf("invalid signature")
}
