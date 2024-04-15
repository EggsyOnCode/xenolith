package crypto_lib

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"math/big"

	"github.com/EggsyOnCode/xenolith/core_types"
)

type PrivateKey struct {
	key *ecdsa.PrivateKey
}

func NewPrivateKeyUsingReader(r io.Reader) *PrivateKey {
	key, err := ecdsa.GenerateKey(elliptic.P256(), r)
	if err != nil {
		panic(err)
	}

	return &PrivateKey{key: key}
}

// the SECP256k1 curve is used in generating the private,public key pairs
// the private key returned also has an embedded field for the public key
func GeneratePrivateKey() *PrivateKey {
	return NewPrivateKeyUsingReader(rand.Reader)
}

// msg are signed with PrivateKey
func (p *PrivateKey) Sign(data []byte) (*Signature, error) {

	r, s, err := ecdsa.Sign(rand.Reader, p.key, data)
	if err != nil {
		return nil, err
	}

	return &Signature{R: r, S: s}, nil
}

type PublicKey []byte

func (p *PrivateKey) PublicKey() PublicKey {
	return elliptic.MarshalCompressed(elliptic.P256(), p.key.X, p.key.Y)
}

// slice of bytes which would be sent over the network

func (p PublicKey) String() string {
	return hex.EncodeToString(p)
}
func (p PublicKey) Address() core_types.Address {
	hash := sha256.Sum256(p)

	//the last 20 bytes are the address
	return core_types.AddressFromBytes(hash[len(hash)-20:])
}

type Signature struct {
	R, S *big.Int
}

func (sig *Signature) String() string {
	return sig.R.String() + sig.S.String()
}

// msg can be verified with the public key
func (sig *Signature) Verify(data []byte, p []byte) bool {
	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), p)
	pk := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
	return ecdsa.Verify(pk, data, sig.R, sig.S)
}
