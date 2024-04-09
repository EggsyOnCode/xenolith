package core

import (
	"fmt"
	"sync"

	"github.com/go-kit/log"
)

type Blockchain struct {
	Version   uint32
	logger    log.Logger
	lock      sync.Mutex
	headers   []*Header
	store     Storage
	Validator Validator
	//to store the state of all the smart contracts on the blockchain
	//TODO implement an interface for the State
	contractState *State
}

// Constructor for Blockchain
func NewBlockchain(genesis *Block, logger log.Logger) (*Blockchain, error) {
	bc := &Blockchain{
		contractState: NewState(),
		headers:       []*Header{},
		store:         NewMemoryStore(),
		logger:        logger,
		Version:       1,
	}

	bc.Validator = NewBlockValidator(bc)

	err := bc.addBlockWithoutValidation(genesis)
	if err != nil {
		return bc, err
	}

	return bc, nil
}

// A dynamic setter for Validator
func (bc *Blockchain) SetValidator(v Validator) {
	bc.Validator = v
}

// adding a new block to the chain
func (bc *Blockchain) AddBlock(b *Block) error {
	//validate block

	err := bc.Validator.ValidateBlock(b)
	if err != nil {
		return err
	}

	//run the block data i.e the code on the VM
	for _, tx := range b.Transactions {
		bc.logger.Log("msg", "executing code", "tx", tx.Hash(&TxHasher{}), "len of the data", len(tx.Data))

		vm := NewVM(tx.Data, bc.contractState)

		if err := vm.Run(); err != nil {
			return err
		}
		// fmt.Printf("STATE : %+v\n", vm.contractState)
		result := vm.stack.Pop()
		fmt.Printf("VM : %+v\n", result)

	}

	//adding the block headers to blockchain headers list
	bc.lock.Lock()
	bc.headers = append(bc.headers, b.Header)
	bc.lock.Unlock()
	//add block to the chain via Put method of store
	bc.store.Put(b)

	bc.logger.Log(
		"msg", "added new block to the chain",
		"hash", b.Hash(BlockHasher{}),
		"height", b.Header.Height,
		"transactions", len(b.Transactions),
	)

	return nil
}

// Return height of the Blockchain
func (bc *Blockchain) Height() uint32 {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	return uint32(len(bc.headers) - 1)
}

func (bc *Blockchain) GetHeaders(height uint32) (*Header, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("Block with height %v is too high", height)
	}

	bc.lock.Lock()
	defer bc.lock.Unlock()
	//when adding the first block with height 1 , the height of the blockchain is 0 therefore we can't access bc.headers[1]
	return bc.headers[height], nil

}

// for adding say a genesis  block
func (bc *Blockchain) addBlockWithoutValidation(b *Block) error {
	bc.lock.Lock()
	bc.headers = append(bc.headers, b.Header)
	bc.lock.Unlock()
	//add block to the chain
	bc.store.Put(b)

	return nil
}

func (bc *Blockchain) HasBlock(height uint32) bool {
	return height <= bc.Height()
}

func (bc *Blockchain) SetLogger(l log.Logger) {
	bc.logger = l
}
