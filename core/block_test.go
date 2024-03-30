package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/EggsyOnCode/xenolith/core_types"
	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/stretchr/testify/assert"
)

func randomBlock(head uint32) *Block {
	header := &Header{
		Version:       1,
		Head:          head,
		PrevBlockHash: core_types.GenerateRandomHash(32),
		Timestamp:     uint64(time.Now().UnixNano()),
	}
	tx := &Transaction{
		Data: []byte("Hello World"),
	}
	return &Block{
		Header:       header,
		Transactions: []Transaction{*tx},
	}
}

func TestBlock(t *testing.T) {
	block := randomBlock(1)
	fmt.Println(block.Hash(BlockHasher{}))
}

func TestSignAndVerifyBlock(t *testing.T) {
	block := randomBlock(1)
	priv := crypto_lib.GeneratePrivateKey()
	err := block.Sign(priv)
	assert.Nil(t, err)
	assert.Equal(t, block.Validator, priv.PublicKey())
	fmt.Println(block.Signature)

	verification, _ := block.Verify()
	assert.True(t, verification)
}

func TestVerifyFail(t *testing.T){
	block := randomBlock(1)
	priv := crypto_lib.GeneratePrivateKey()
	err := block.Sign(priv)
	ver, _ := block.Verify()
	assert.True(t, ver)
	assert.Nil(t, err)
	assert.Equal(t, block.Validator, priv.PublicKey())
	fmt.Println(block.Signature)

	//header info changing; 
	block.Header.Head = 20
	verification, _ := block.Verify()
	fmt.Println(verification)
	assert.False(t, verification)
}