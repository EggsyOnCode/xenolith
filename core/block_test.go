package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/EggsyOnCode/xenolith/core_types"
	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/stretchr/testify/assert"
)

func randomBlock(height uint32) *Block {
	header := &Header{
		Version:       1,
		Height:        height,
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

func randomBlockWithSignature(t *testing.T, height uint32, prevHash core_types.Hash) *Block {
	header := &Header{
		Version:       1,
		Height:        height,
		PrevBlockHash: prevHash,
		Timestamp:     uint64(time.Now().UnixNano()),
	}

	//generating a private key
	priv := crypto_lib.GeneratePrivateKey()
	tx := &Transaction{
		Data: []byte("Hello World"),
	}
	tx.Sign(priv)
	block := &Block{
		Header:       header,
		Transactions: []Transaction{*tx},
	}
	assert.Nil(t, block.Sign(priv))
	return block
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

// func TestVerifyFail(t *testing.T){
// 	block := randomBlock(1)
// 	priv := crypto_lib.GeneratePrivateKey()
// 	err := block.Sign(priv)
// 	ver, _ := block.Verify()
// 	assert.True(t, ver)
// 	assert.Nil(t, err)
// 	assert.Equal(t, block.Validator, priv.PublicKey())
// 	fmt.Println(block.Signature)

// 	//header info changing;
// 	block.Header.Head = 20
// 	verification, _ := block.Verify()
// 	fmt.Println(verification)
// 	assert.False(t, verification)
// }
