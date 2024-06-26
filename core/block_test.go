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
		NBits:         0x1d00ffff,
	}

	//generating a private key
	priv := crypto_lib.GeneratePrivateKey()
	block := &Block{
		Header:     header,
		Validator:  priv.PublicKey(),
		PrevBlock:  b,
		NextBlocks: make([]*Block, 0),
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
		NBits:         0x1d00ffff,
	}

	header.Target = compactToTarget(header.NBits)
	//generating a private key
	priv := crypto_lib.GeneratePrivateKey()
	block := &Block{
		Header:     header,
		Validator:  priv.PublicKey(),
		PrevBlock:  &Block{},
		NextBlocks: make([]*Block, 0),
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

func genesisBlockWithSig(t *testing.T, height uint32, prevHash core_types.Hash) *Block {
	header := &Header{
		Version:       1,
		Height:        height,
		PrevBlockHash: prevHash,
		Timestamp:     uint64(time.Now().UnixNano()),
		NBits:         0x1d00ffff,
	}

	header.Target = compactToTarget(header.NBits)
	// header.Target = new(big.Int)
	// _, ok := header.Target.SetString("0x00000000FFFF0000000000000000000000000000000000000000000000000000", 16)
	// if !ok {
	// 	fmt.Errorf("error setting target")
	// }

	//generating a private key
	priv := crypto_lib.GeneratePrivateKey()
	block := &Block{
		Header:     header,
		Validator:  priv.PublicKey(),
		PrevBlock:  &Block{},
		NextBlocks: make([]*Block, 0),
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

func TestHashFunc(t *testing.T) {
	block := randomBlockWithSignature(t, 1, core_types.GenerateRandomHash(32))
	hash := block.HashWithoutCache(BlockHasher{})
	fmt.Println(hash)

	block.Header.Nonce++
	hash2 := block.HashWithoutCache(BlockHasher{})
	fmt.Println(hash2)
	assert.NotEqual(t, hash, hash2)
}
