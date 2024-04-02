package core

type Blockchain struct {
	headers   []*Header
	store     Storage
	Validator Validator
}

// Constructor for Blockchain
func NewBlockchain(genesis *Block) (*Blockchain, error) {
	bc := &Blockchain{
		headers: []*Header{},
		store:   NewMemoryStore(),
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
	if err!=nil{
		return err
	}
	//adding the block headers to blockchain headers list
	bc.headers = append(bc.headers, b.Header)
	//add block to the chain via Put method of store
	bc.store.Put(b)

	return nil
}

// Return height of the Blockchain
func (bc *Blockchain) Height() uint32 {
	return uint32(len(bc.headers) - 1)
}

// for adding say a genesis  block
func (bc *Blockchain) addBlockWithoutValidation(b *Block) error {
	bc.headers = append(bc.headers, b.Header)
	//add block to the chain
	bc.store.Put(b)

	return nil
}

func (bc *Blockchain) HasBlock(height uint32) bool{
	return height <= bc.Height()
}