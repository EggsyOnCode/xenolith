package core

import (
	"testing"

	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/stretchr/testify/assert"
)

func TestSignTX(t *testing.T){
	tx := &Transaction{
		Data: []byte("Hello World"),
	}
	priv := crypto_lib.GeneratePrivateKey()
	err := tx.Sign(priv)
	assert.Nil(t, err)
	assert.NotNil(t, tx.Signature)
	assert.Equal(t, tx.From, priv.PublicKey())
}
func TestTxVerification(t *testing.T){
	tx := &Transaction{
		Data: []byte("Hello World"),
	}
	priv := crypto_lib.GeneratePrivateKey()
	err := tx.Sign(priv)
	assert.Nil(t, err)
	assert.NotNil(t, tx.Signature)
	assert.Equal(t, tx.From, priv.PublicKey())
	verfication, _ := tx.Verify()
	assert.True(t, verfication)

	otherPv := crypto_lib.GeneratePrivateKey()
	tx.From = otherPv.PublicKey()

	ver, _ := tx.Verify()
	assert.False(t, ver)

}
