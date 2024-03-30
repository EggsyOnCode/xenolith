package core

import (
	"crypto"

	"github.com/EggsyOnCode/xenolith/crypto_lib"
)

type Transaction struct {
	Data      []byte
	Validator crypto.PublicKey
	Signature *crypto_lib.Signature
}
