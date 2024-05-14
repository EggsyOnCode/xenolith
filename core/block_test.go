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
		Header:       header,
		Transactions: []*Transaction{tx},
	}
	datahash, err := CalculateDataHash(block.Transactions)
	assert.Nil(t, err)
	block.Header.DataHash = datahash

	return block
}

func randomBlockWithSignatureAndPrevBlock(t *testing.T, height uint32, prevHash core_types.Hash, b *Block) *Block {
	header := &Header{
		Version:       1,
		Height:        height,
		PrevBlockHash: prevHash,
		Timestamp:     uint64(time.Now().UnixNano()),
	}

	//generating a private key
	priv := crypto_lib.GeneratePrivateKey()
	block := &Block{
		Header:    header,
		Validator: priv.PublicKey(),
		PrevBlock: b,
	}
	tx := randomTxWithSignature(t)
	fmt.Printf("validator is %v\n", block.Validator)
	block.Transactions = append(block.Transactions, tx)
	datahash, err := CalculateDataHash(block.Transactions)
	assert.Nil(t, err)
	block.Header.DataHash = datahash

	assert.Nil(t, block.Sign(priv))
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
	block := &Block{
		Header:    header,
		Validator: priv.PublicKey(),
		PrevBlock: &Block{},
	}
	tx := randomTxWithSignature(t)
	fmt.Printf("validator is %v\n", block.Validator)
	block.Transactions = append(block.Transactions, tx)
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

func TestVerifyFail(t *testing.T) {
	block := randomBlockWithSignature(t, 1, core_types.GenerateRandomHash(32))
	err := block.Verify()
	assert.Nil(t, err)
	fmt.Printf("validator is before chaning %+v\n", block.Validator)

	otherPrivKey := crypto_lib.GeneratePrivateKey()
	block.Validator = otherPrivKey.PublicKey()
	assert.NotNil(t, block.Verify())

	//header info changing;
	block.Header.Height = 100
	assert.NotNil(t, block.Verify())
}
