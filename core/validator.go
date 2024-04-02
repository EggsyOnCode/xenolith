package core

import "fmt"

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
	//for now: if the height of the proposed block is not one greater than the current height, return an error
	if v.bc.HasBlock(b.Header.Height) {
		return fmt.Errorf("Block with height %v and hash %v already exists", b.Header.Height, b.Hash(BlockHasher{}))
	}

	//verifies if the block has been signed
	if ans, err := b.Verify(); err!=nil&&!ans{
		return err
	}
	return nil
}
