package core

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/stretchr/testify/assert"
)

func TestNFTTx(t *testing.T) {
	pk := crypto_lib.GeneratePrivateKey()
	collection := &CollectionTx{
		Fee:      100,
		MetaData: []byte("Hello World Token from China!"),
		Quantity: 20000,
	}

	tx := &Transaction{
		TxInner: collection,
	}

	tx.Sign(pk)
	buf := &bytes.Buffer{}
	assert.Nil(t, gob.NewEncoder(buf).Encode(tx))

	txDecoded := new(Transaction)
	assert.Nil(t, gob.NewDecoder(buf).Decode(txDecoded))

	assert.Equal(t, tx, txDecoded)
}

func TestNativeTokenTransfer(t *testing.T) {
	senderPk := crypto_lib.GeneratePrivateKey()
	receiverPk := crypto_lib.GeneratePrivateKey()

	tx := &Transaction{
		From:  senderPk.PublicKey(),
		To:    receiverPk.PublicKey(),
		Value: 100,
	}

	assert.Nil(t, tx.Sign(senderPk))
	assert.NotNil(t, tx.Signature)

}

func TestSignTX(t *testing.T) {
	tx := &Transaction{
		Data: []byte("Hello World"),
	}
	priv := crypto_lib.GeneratePrivateKey()
	err := tx.Sign(priv)
	assert.Nil(t, err)
	assert.NotNil(t, tx.Signature)
	assert.Equal(t, tx.From, crypto_lib.PublicKey(crypto_lib.PublicKey(priv.PublicKey())))
}
func TestTxVerification(t *testing.T) {
	tx := &Transaction{
		Data: []byte("Hello World"),
	}
	priv := crypto_lib.GeneratePrivateKey()
	err := tx.Sign(priv)
	assert.Nil(t, err)
	assert.NotNil(t, tx.Signature)
	assert.Equal(t, tx.From, crypto_lib.PublicKey(priv.PublicKey()))
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

func randomTxWithSignature(t *testing.T) *Transaction {
	privKey := crypto_lib.GeneratePrivateKey()
	tx := Transaction{
		Data: []byte("foo"),
	}
	assert.Nil(t, tx.Sign(privKey))

	return &tx
}
