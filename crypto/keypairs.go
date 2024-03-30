package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"

	"github.com/EggsyOnCode/xenolith/core_types"
)

type PrivateKey struct {
	key *ecdsa.PrivateKey
}

// the SECP256k1 curve is used in generating the private,public key pairs
// the private key returned also has an embedded field for the public key
func GeneratePrivateKey() *PrivateKey {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	return &PrivateKey{key: key}
}

func (p *PrivateKey) PublicKey() *PublicKey {
	return &PublicKey{key: &p.key.PublicKey}
}

type PublicKey struct {
	key *ecdsa.PublicKey
}

func (p *PublicKey) ToSlice() []byte {
	return elliptic.MarshalCompressed(elliptic.P256(), p.key.X, p.key.Y)
}

func (p *PublicKey) Address() core_types.Address {
	hash := sha256.Sum256(p.ToSlice())

	//the last 20 bytes are the address 
	return core_types.AddressFromBytes(hash[len(hash)-20:])
}

type Signature struct{}

