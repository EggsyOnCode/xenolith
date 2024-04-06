package util

import (
	"math/rand"
	"testing"
	"time"

	"github.com/EggsyOnCode/xenolith/core"
	"github.com/EggsyOnCode/xenolith/core_types"
	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/stretchr/testify/assert"
)

func RandomBytes(size int) []byte {
	token := make([]byte, size)
	rand.Read(token)
	return token
}

func RandomHash() core_types.Hash {
	return core_types.HashFromBytes(RandomBytes(32))
}

// NewRandomTransaction return a new random transaction whithout signature.
func NewRandomTransaction(size int) *core.Transaction {
	return core.NewTransaction(RandomBytes(size))
}

func NewRandomTransactionWithSignature(t *testing.T, privKey *crypto_lib.PrivateKey, size int) *core.Transaction {
	tx := NewRandomTransaction(size)
	assert.Nil(t, tx.Sign(privKey))
	return tx
}

func NewRandomBlock(t *testing.T, height uint32, prevBlockHash core_types.Hash) *core.Block {
	txSigner := crypto_lib.GeneratePrivateKey()
	tx := NewRandomTransactionWithSignature(t, txSigner, 100)
	header := &core.Header{
		Version:       1,
		PrevBlockHash: prevBlockHash,
		Height:        height,
		Timestamp:     uint64(time.Now().UnixNano()),
	}
	b := core.NewBlock(header, []*core.Transaction{tx})
	dataHash, err := core.CalculateDataHash(b.Transactions)
	assert.Nil(t, err)
	b.Header.DataHash = dataHash

	return b
}

func NewRandomBlockWithSignature(t *testing.T, pk *crypto_lib.PrivateKey, height uint32, prevHash core_types.Hash) *core.Block {
	b := NewRandomBlock(t, height, prevHash)
	assert.Nil(t, b.Sign(pk))

	return b
}
