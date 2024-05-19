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
	belongsToFork := true
	// check if the block is meant for a fork
	_, err := v.bc.ForkSlice.FindBlock(b.PrevBlock.hash)
	if err != nil {
		belongsToFork = false
	}
	// if the height of the proposed block is not less than the current height, return an error
	var block *Block
	if v.bc.HasBlock(b.Header.Height) {
		block, _ = v.bc.GetBlock(b.Header.Height)
		// if it returns false then that means a fork has been detected
		if block.Header.PrevBlockHash != b.Header.PrevBlockHash && !belongsToFork {
			// return fmt.Errorf("Block with height %v and hash %v already exists", b.Header.Height, b.Hash(BlockHasher{}))
			return ErrBlockKnown
		}
	}

	///the height of the proposed block should be one greater than the current height
	if b.Header.Height != v.bc.Height()+1 && (block.Header.PrevBlockHash != b.Header.PrevBlockHash) && !belongsToFork {
		return fmt.Errorf("block (%s) with height (%d) is too high => current height (%d)", b.Hash(BlockHasher{}), b.Header.Height, v.bc.Height())
	}

	//getting headers for current block
	prevBlockHeader, err := v.bc.GetHeaders(b.Header.Height - 1)
	if err != nil {
		return err
	}

	//the previous block hash of the proposed block should be equal to the hash of the last block in the blockchain
	hash := BlockHasher{}.Hash(prevBlockHeader)
	hashOfHeaderBlock := v.bc.block.Hash(BlockHasher{})
	// reason for dual assertion is that; in a forking scenario
	// the block headers and the blockchain linked list could go out of sync
	if hash != b.Header.PrevBlockHash && hashOfHeaderBlock != b.Header.PrevBlockHash && !belongsToFork {
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
