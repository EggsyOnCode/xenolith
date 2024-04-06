package core

import (
	"os"
	"testing"

	"github.com/EggsyOnCode/xenolith/core_types"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
)

func TestBlockchain(t *testing.T) {
	//genesis block
	bc := newBlockchainWithGenesis(t)
	assert.Equal(t, bc.Height(), uint32(0))
}

func TestHasBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	assert.True(t, bc.HasBlock(0))
	assert.False(t, bc.HasBlock(100))
}

func TestAddBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	lenB := 2
	for i := 0; i < lenB; i++ {
		prevHash := getPrevBlockHash(t, bc, uint32(i+1))
		block := randomBlockWithSignature(t, uint32(i+1), (prevHash))
		err := bc.AddBlock(block)
		assert.Nil(t, err)
	}

	assert.Equal(t, bc.Height(), uint32(lenB))
	assert.Equal(t, len(bc.headers), lenB+1)
	//since the current height of the blockchain is 1000, adding a block with height 89 should return an error
	assert.NotNil(t, bc.AddBlock(randomBlockWithSignature(t, 89, core_types.Hash{})))
}

func TestBlockTooHigh(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	block := randomBlockWithSignature(t, 1000, core_types.Hash{})
	assert.NotNil(t, bc.AddBlock(block))
	block1 := randomBlockWithSignature(t, 1, core_types.Hash{})
	assert.NotNil(t, bc.AddBlock(block1))
}
func TestGetHeaders(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	for i := 0; i < 2; i++ {
		prevHash := getPrevBlockHash(t, bc, uint32(i+1))
		block := randomBlockWithSignature(t, uint32(i+1), prevHash)
		err := bc.AddBlock(block)
		assert.Nil(t, err)
		header, err := bc.GetHeaders(block.Header.Height)
		assert.Nil(t, err)
		assert.Equal(t, header, block.Header)
	}

}

func newBlockchainWithGenesis(t *testing.T) *Blockchain {
	block := randomBlockWithSignature(t, 1, core_types.Hash{})
	logger := log.NewLogfmtLogger(os.Stderr)
	bc, err := NewBlockchain(block, logger)
	assert.Nil(t, err)
	return bc
}

func getPrevBlockHash(t *testing.T, bc *Blockchain, height uint32) core_types.Hash {
	prevHeader, err := bc.GetHeaders(height - 1)
	assert.Nil(t, err)
	return BlockHasher{}.Hash(prevHeader)
}
