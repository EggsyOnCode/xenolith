package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newBlockWithGenesis(t *testing.T) *Blockchain {
	block := randomBlockWithSignature(t, 1)
	bc, err := NewBlockchain(block)
	assert.Nil(t, err)
	return bc
}
func TestBlockchain(t *testing.T) {
	//genesis block
	bc := newBlockWithGenesis(t)
	assert.Equal(t, bc.Height(), uint32(0))
}

func TestHasBlock(t *testing.T) {
	bc := newBlockWithGenesis(t)
	assert.True(t, bc.HasBlock(0))
}

func TestAddBlock(t *testing.T) {
	bc := newBlockWithGenesis(t)
	for i := 1; i < 1000; i++ {
		block := randomBlockWithSignature(t, uint32(i))
		err := bc.AddBlock(block)
		assert.Nil(t, err)
		assert.Equal(t, bc.Height(), uint32(i))
	}

	assert.Equal(t, len(bc.headers), 1000)
	//since tthe current height of the blockchain is 1000, adding a block with height 89 should return an error
	assert.NotNil(t, bc.AddBlock(randomBlockWithSignature(t, 89)))
}
