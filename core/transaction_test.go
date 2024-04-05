package core

import (
	"bytes"
	"testing"

	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/stretchr/testify/assert"
)

func TestSignTX(t *testing.T) {
	tx := &Transaction{
		Data: []byte("Hello World"),
	}
	priv := crypto_lib.GeneratePrivateKey()
	err := tx.Sign(priv)
	assert.Nil(t, err)
	assert.NotNil(t, tx.Signature)
	assert.Equal(t, tx.From, priv.PublicKey())
}
func TestTxVerification(t *testing.T) {
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

func TestCodecTx(t *testing.T) {
	tx := &Transaction{
		Data: []byte("Hello World"),
	}

	buf := &bytes.Buffer{}

	assert.Nil(t, tx.Encode(NewGobTxEncoder(buf)))

	//the value of the encoded tx will be decoded into txDecoded
	txDecoded := new(Transaction)
	assert.Nil(t, txDecoded.Decode(NewGobTxDecoder(buf)))

	assert.Equal(t, tx.Data, txDecoded.Data)

}
