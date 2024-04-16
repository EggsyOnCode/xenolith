package core

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/EggsyOnCode/xenolith/core_types"
	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/stretchr/testify/assert"
)

func randomBlock(t *testing.T, height uint32, prevHash core_types.Hash) *Block {
	header := &Header{
		Version:       1,
		Height:        height,
		PrevBlockHash: prevHash,
		Timestamp:     uint64(time.Now().UnixNano()),
	}
	tx := randomTxWithSignature(t)
	block := &Block{
		Header: header,
		Transactions: []*Transaction{tx},
	}
	datahash, err := CalculateDataHash(block.Transactions)
	assert.Nil(t, err)
	block.Header.DataHash = datahash

	return block
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
		Transactions: []*Transaction{tx},
	}
	datahash, err := CalculateDataHash(block.Transactions)
	assert.Nil(t, err)
	block.Header.DataHash = datahash

	assert.Nil(t, block.Sign(priv))
	return block
}

// func TestBlock(t *testing.T) {
// 	block := randomBlock(t, 1, getPrevBlockHash(t, ))
// 	fmt.Println(block.Hash(BlockHasher{}))
// }

func TestSignAndVerifyBlock(t *testing.T) {
	block := randomBlockWithSignature(t, 3, core_types.GenerateRandomHash(32))
	priv := crypto_lib.GeneratePrivateKey()
	err := block.Sign(priv)
	assert.Nil(t, err)
	assert.Equal(t, block.Validator, crypto_lib.PublicKey(priv.PublicKey()))
	fmt.Println(block.Signature)

	err = block.Verify()
	assert.Nil(t, err)
}

func TestCodecBlock(t *testing.T) {
	block := randomBlockWithSignature(t, 1, core_types.GenerateRandomHash(32))
	buf := &bytes.Buffer{}
	assert.Nil(t, block.Encode(NewGobBlockEncoder(buf)))

	blockDecoded := new(Block)
	assert.Nil(t, blockDecoded.Decode(NewGobBlockDecoder(buf)))

	assert.Equal(t, block.Header, blockDecoded.Header)

}

// func TestVerifyFail(t *testing.T){
// 	block := randomBlock(1)
// 	priv := crypto_lib.GeneratePrivateKey()
// 	err := block.Sign(priv)
// 	ver, _ := block.Verify()
// 	assert.True(t, ver)
// 	assert.Nil(t, err)
// 	assert.Equal(t, block.Validator, crypto_lib.PublicKey(priv.PublicKey()))
// 	fmt.Println(block.Signature)

// 	//header info changing;
// 	block.Header.Head = 20
// 	verification, _ := block.Verify()
// 	fmt.Println(verification)
// 	assert.False(t, verification)
// }
