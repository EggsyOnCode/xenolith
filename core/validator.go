package core

import (
	"errors"
	"fmt"
)

var ErrBlockKnown = errors.New("Block already known")

type Validator interface {
	ValidateBlock(*Block) error
}

type BlockValidator struct {
	bc *Blockchain
}

func NewBlockValidator(bc *Blockchain) *BlockValidator {
	return &BlockValidator{bc: bc}
}

func (v *BlockValidator) ValidateBlock(b *Block) error {
	// Validate block
	// if the height of the proposed block is not less than the current height, return an error
	if v.bc.HasBlock(b.Header.Height) {
		// return fmt.Errorf("Block with height %v and hash %v already exists", b.Header.Height, b.Hash(BlockHasher{}))
		return ErrBlockKnown
	}

	///the height of the proposed block should be one greater than the current height
	if b.Header.Height != v.bc.Height()+1 {
		return fmt.Errorf("block (%s) with height (%d) is too high => current height (%d)", b.Hash(BlockHasher{}), b.Header.Height, v.bc.Height())
	}

	//getting headers for current block
	prevHash, err := v.bc.GetHeaders(b.Header.Height - 1)
	if err != nil {
		return err
	}

	//the previous block hash of the proposed block should be equal to the hash of the last block in the blockchain
	hash := BlockHasher{}.Hash(prevHash)
	if hash != b.Header.PrevBlockHash {
		return fmt.Errorf("Block with height %v and hash %v has a different previous block hash %v", b.Header.Height, b.Hash(BlockHasher{}), b.Header.PrevBlockHash)
	}

	//verifies the transactions in the block
	for _, tx := range b.Transactions {
		if ans, err := tx.Verify(); err != nil || !ans {
			fmt.Printf("Error: %v\n", err)
			return err
		}
	}

	//verifies if the block has been signed
	if err := b.Verify(); err != nil {
		return err
	}

	return nil
}
