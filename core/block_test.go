package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/EggsyOnCode/xenolith/core_types"
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
