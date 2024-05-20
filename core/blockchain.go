package core

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/EggsyOnCode/xenolith/core_types"
	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/go-kit/log"
)

const (
	//target readjusted after this number
	HEIGHT_DIVISOR = 5
	// 60 sec
	AVG_TARGET_TIME = 1 * 60
	TARGET_GENESIS  = 0x00ffff0000000000000000000000000000000000000000000000000000
)

type ReturnTypeForkHandler byte

const (
	FORK_IN_LONGEST_CHAIN     ReturnTypeForkHandler = 0x1
	FORK_NOT_IN_LONGEST_CHAIN ReturnTypeForkHandler = 0x2
	NOT_FORKING               ReturnTypeForkHandler = 0x3
)

type Blockchain struct {
	Version uint32
	logger  log.Logger

	lock    sync.Mutex
	headers []*Header
	// blocks  []*Block

	ChainTip  *Block
	forkLock  *sync.RWMutex
	forkCount uint32
	ForkSlice
	//since blochchain is a linked list of blocks, we just need to first genesis block and we can accses the rest of the link using it
	block            *Block
	blockStore       map[core_types.Hash]*Block
	blockStoreHeight map[uint32]*Block
	//TODO: add a diff mutex to make txSTore thread safe; currenlty both teh stores are usint he same mutex; which is bad!
	txStore map[core_types.Hash]*Transaction
	//TODO: impelment a better data structure to store the collection like merkle trees etc
	//TODO: add a diff mutex to make collectionStore thread safe; currenlty both teh stores are usint he same mutex; which is bad!
	collectionStore map[core_types.Hash]*CollectionTx
	mintStore       map[core_types.Hash]*MintTx
	stateLock       sync.RWMutex
	accountState    *AccountState

	store     Storage
	Validator Validator
	//to store the state of all the smart contracts on the blockchain
	//TODO implement an interface for the State
	contractState *State
	target        *big.Int
	// a channel shared between blockchain and server for sharing orphaned Tx into server's mempool
	txCh chan *Transaction
}

// Constructor for Blckchain
func NewBlockchain(genesis *Block, logger log.Logger) (*Blockchain, error) {
	//the responsibility of creating and managing the account state falls on the blockchain
	//read the accountState from Disk (TODO)
	accountState := NewAccountState()

	//the default value of Public Key
	coinbase := crypto_lib.PublicKey{}
	fmt.Println("COINBASE : ", coinbase.Address())
	accountState.CreateAccount(coinbase.Address())

	bc := &Blockchain{
		contractState: NewState(),
		headers:       []*Header{},
		store:         NewMemoryStore(),
		logger:        logger,
		Version:       1,
		// blocks:           make([]*Block, 1),
		ChainTip:         genesis,
		block:            genesis,
		ForkSlice:        make(ForkSlice),
		forkLock:         &sync.RWMutex{},
		forkCount:        0,
		blockStore:       make(map[core_types.Hash]*Block),
		blockStoreHeight: make(map[uint32]*Block),
		txStore:          make(map[core_types.Hash]*Transaction),
		collectionStore:  make(map[core_types.Hash]*CollectionTx),
		mintStore:        make(map[core_types.Hash]*MintTx),
		accountState:     accountState,
		stateLock:        sync.RWMutex{},
	}

	bc.Validator = NewBlockValidator(bc)

	err := bc.addBlockWithoutValidation(genesis)
	//--> what the implementation should be!
	// err := bc.AddBlock(genesis)
	if err != nil {
		return bc, err
	}

	return bc, nil
}

// A dynamic setter for Validator
func (bc *Blockchain) SetValidator(v Validator) {
	bc.Validator = v
}

// A dynamic setter for txChan
func (bc *Blockchain) SetTxChan(t chan *Transaction) {
	bc.txCh = t
}

// adding a new block to the chain
func (bc *Blockchain) AddBlock(b *Block) error {
	//validate block

	err := bc.Validator.ValidateBlock(b)
	if err != nil {
		return err
	}

	return bc.addBlockWithoutValidation(b)

}

// Return height of the Blockchain
func (bc *Blockchain) Height() uint32 {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	return uint32(len(bc.headers) - 1)
}

func (bc *Blockchain) GetBlockByHash(hash core_types.Hash) (*Block, error) {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	block, ok := bc.blockStore[hash]
	if !ok {
		return nil, fmt.Errorf("block with hash (%s) not found", hash)
	}

	return block, nil
}

func (bc *Blockchain) GetBlock(height uint32) (*Block, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("Block with height %v is too high", height)
	}

	bc.lock.Lock()
	defer bc.lock.Unlock()
	//when adding the first block with height 1 , the height of the blockchain is 0 therefore we can't access bc.headers[1]
	return bc.blockStoreHeight[height+1], nil

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

func (bc *Blockchain) GetTxByHash(hash core_types.Hash) (*Transaction, error) {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	tx, ok := bc.txStore[hash]
	if !ok {
		return nil, fmt.Errorf("transaction with hash (%s) not found", hash)
	}

	return tx, nil

}

func (bc *Blockchain) handleNativeNFT(tx *Transaction) error {
	hash := tx.Hash(&TxHasher{})
	switch t := tx.TxInner.(type) {
	case *CollectionTx:
		bc.collectionStore[hash] = t

		bc.logger.Log("msg", "added collection tx to the store", "hash", hash)
	case *MintTx:
		_, ok := bc.collectionStore[t.Collection]
		if !ok {
			return fmt.Errorf("collection (%v) does NOt exist ", t.Collection)
		}

		bc.mintStore[hash] = t

		bc.logger.Log("msg", "created new NFT mint", "nft", t.NFT, "collection", t.Collection)

	default:
		return fmt.Errorf("unsupported tx type %v", t)
	}

	return nil
}

func (bc *Blockchain) handleTx(tx *Transaction) error {

	// execute the tx on the vm only if the data field is populated
	if len(tx.Data) > 0 {
		bc.logger.Log("msg", "executing code", "tx", tx.Hash(&TxHasher{}), "len of the data", len(tx.Data))
		vm := NewVM(tx.Data, bc.contractState)
		if err := vm.Run(); err != nil {
			return err
		}
		// fmt.Printf("STATE : %+v\n", vm.contractState)
		// result := vm.stack.Pop()
		// fmt.Printf("VM : %+v\n", result)

	}

	//handling native NFT tokens
	if tx.TxInner != nil {
		if err := bc.handleNativeNFT(tx); err != nil {
			return err
		}
	}

	//otherwise handle the native token tx
	if tx.Value > 0 {
		if err := bc.handleTransferNativeTokens(tx); err != nil {
			bc.logger.Log("err", "error while handling transfer of native tokens", "err", err)
			return err
		}
	}

	return nil

}

func (bc *Blockchain) revertTx(tx *Transaction) error {
	if err := tx.Revert(); err != nil {
		return err
	}
	return bc.handleTx(tx)
}

// / @dev : this internal func gets executed during a chain reorg; this happens when fork confirmations reach >= 3
func (bc *Blockchain) handleChainReorg(fork *Fork) error {
	if fork.IsLongestChain {
		// remove the forkPair from the fork slice ;
		bc.ForkSlice.RemoveForkPair(bc.ForkSlice[bc.ForkSlice.FindForkPairId(fork)])
	}
	// the winning chian is one in the processing queue
	// put the blocks which are to be removed in an array
	// iterate thru each block and revert its tx and remove it from blockstore and headers
	// remove the forkPair

	forkPair := bc.ForkSlice[bc.ForkSlice.FindForkPairId(fork)]
	var toBeRemovedFork *Fork
	var toBeRemovedBlocks []*Block
	if forkPair.Forks[0] == fork {
		toBeRemovedFork = forkPair.Forks[1]
	} else {
		toBeRemovedFork = forkPair.Forks[0]
	}
	//bock from which the fork started not the one that caused the fork
	forkingBlock, _ := bc.GetBlockByHash(toBeRemovedFork.ForkingBlock)
	// forkingBlock.NextBlocks[0] will be the longestChain so [1] will be the forked block
	startingBlock := forkingBlock.NextBlocks[0]

	forkChainTipBlock := bc.block

	// toBeRemoved Block exists on Chain so normalGetters used
	toBeRemovedBlocks = append(toBeRemovedBlocks, startingBlock)

	for {
		// Ensure startingBlock has NextBlocks before accessing
		if len(startingBlock.NextBlocks) == 0 {
			fmt.Printf("No next blocks found for block %v\n", startingBlock.Hash(BlockHasher{}))
			break
		}

		nextBlock := startingBlock.NextBlocks[0]
		toBeRemovedBlocks = append(toBeRemovedBlocks, nextBlock)

		// Update startingBlock
		startingBlock = nextBlock

		// Check if we reached the chain tip block
		if startingBlock == forkChainTipBlock {
			break
		}
	}

	// forkingBlock, _ := bc.GetBlockByHash(forkingBlockHash)
	// forkingBlock.NextBlocks[0] = toBeRemovedBlocks[0]
	// process each block; revert tx; remove from blockstore and headers
	for _, block := range toBeRemovedBlocks {
		for _, tx := range block.Transactions {
			orgTx := tx
			go bc.revertTx(tx)
			delete(bc.txStore, tx.hash)
			// sending this orphaned tx to server's mempool
			bc.txCh <- orgTx
		}

		// removing the block from the Headers, blockStore, blockSToreHeihgt

		//removing the block from the headers
		for i, header := range bc.headers {
			if header == block.Header {
				bc.headers = append(bc.headers[:i], bc.headers[i+1:]...)
				break
			}
		}
		// removing the block from the blockStore
		delete(bc.blockStore, block.Hash(BlockHasher{}))

		//removing the block from the blockStoreHeight
		delete(bc.blockStoreHeight, block.Header.Height+1)

	}

	// attaching the winning fork to the forkingBlock ; the block from which the fork started
	//re init the nextBlocks of the forkingBlock
	forkingBlock.NextBlocks = []*Block{}
	//putting the pointer rolled back to the forkingBlock from where the chain history diverged
	bc.block = forkingBlock
	for _, block := range fork.blocks {
		block.NextBlocks = []*Block{}
		bc.addBlockWithoutValidation(block)
	}

	bc.ForkSlice.RemoveForkPair(forkPair)

	return nil
}

// checks to see if the new incoming block is causing any forks in the chain
// or is attaching itself to a particular forked block etc
// / @dev returns true if the block is causing a fork or attaching itself to a fork
func (bc *Blockchain) handleAndTrackForks(b *Block) ReturnTypeForkHandler {
	// 1st check: if the block is causing a fork --> insertion in forkSlice
	// false indicates that the block is not causing a fork or becoming part of a fork
	var prevBlock *Block
	if b.PrevBlock != nil {
		prevBlock = b.PrevBlock
	} else {
		prevBlock, _ = bc.GetBlockByHash(b.Header.PrevBlockHash)
	}
	if len(prevBlock.NextBlocks) >= 1 {
		forkingFork := &Fork{
			ChainTip:       b.Hash(BlockHasher{}),
			ForkingBlock:   prevBlock.Hash(BlockHasher{}),
			Confirmations:  0,
			IsLongestChain: false,
		}

		competitorFork := &Fork{
			ChainTip:       prevBlock.NextBlocks[0].Hash(BlockHasher{}),
			ForkingBlock:   prevBlock.Hash(BlockHasher{}),
			Confirmations:  0,
			IsLongestChain: true,
		}

		forks := []*Fork{forkingFork, competitorFork}

		bc.forkLock.Lock()
		bc.forkCount++
		bc.ForkSlice[bc.forkCount] = NewForkPair(forks)
		// because this incoming block is causing the fork; its not part of the longest chain
		// hence needs to be added to the processingQueue of hte fork
		bc.ForkSlice[bc.forkCount].AddBlockToProcessingQ(b)
		bc.ForkSlice[bc.forkCount].AddBlock(forkingFork, b)
		bc.ForkSlice[bc.forkCount].AddBlock(competitorFork, b.PrevBlock.NextBlocks[0])

		bc.forkLock.Unlock()

		return FORK_NOT_IN_LONGEST_CHAIN
	}
	// 2nd check: if the block is attaching itself to a forked block (a block in a fork chain temp or otherwise)--> update the forkSlice

	//checking if the incoming block means to attach itself to a fork
	// returns errr if its not a doing that
	_, err := bc.ForkSlice.FindBlock(b.Header.PrevBlockHash)
	if err != nil {
		return NOT_FORKING
	}

	fork, err := bc.ForkSlice.FindBlockFork(b.Header.PrevBlockHash)
	if err != nil {
		return NOT_FORKING
	}

	fork.Confirmations++

	// CHAIN REORG
	if fork.Confirmations >= 3 {
		bc.handleChainReorg(fork)
		return FORK_IN_LONGEST_CHAIN
	}

	/// updating the chaintip of the fork
	forkPair := bc.ForkSlice[bc.ForkSlice.FindForkPairId(fork)]
	blockAtForkTip, _ := forkPair.GetBlock(fork.ChainTip)

	blockAtForkTip.NextBlocks = append(blockAtForkTip.NextBlocks, b)
	fork.ChainTip = b.Hash(BlockHasher{})

	//check if the incoming block is attaching itself to fork with the longest chain or not
	// if YES; that means it will be added to the blockStore and hence the chain ; otehrwise; it will be added to the processing queue
	if fork.IsLongestChain {
		return FORK_IN_LONGEST_CHAIN
	}

	forkId := bc.ForkSlice.FindForkPairId(fork)
	bc.ForkSlice[forkId].AddBlockToProcessingQ(b)
	bc.ForkSlice[forkId].AddBlock(fork, b)

	return FORK_NOT_IN_LONGEST_CHAIN
}

func (bc *Blockchain) addBlockWithoutValidation(b *Block) error {
	// bc.block itself is the head ptr
	// adding the new block to the next block of the current block

	var forkHandlerRes ReturnTypeForkHandler
	//handle forks if present
	// won;t be present for the genesis block
	if b.Header.Height != 0 {
		forkHandlerRes = bc.handleAndTrackForks(b)

		// if the block is not a fork then we can add the block normally to the ll
		// if the block attaching itself to non-dominant fork then we can add the block to the processing queue
		// and hence do nothing here
		// if the block is attaching itself to the longestChainFork then we can add the block to the chain
		switch forkHandlerRes {
		case FORK_NOT_IN_LONGEST_CHAIN:
			// do nothing
			return nil
		default:
			//bc.block rep the ll of blocks
			bc.block.NextBlocks = append(bc.block.NextBlocks, b)
			bc.block = b
		}
	}

	// state changes will  only be of those blocks which are part of the longest chain
	bc.stateLock.Lock()

	//run the block data i.e the code on the VM
	for _, tx := range b.Transactions {
		if err := bc.handleTx(tx); err != nil {
			fmt.Printf("error while handling tx %v\n", err)
			continue
		}
		bc.txStore[tx.Hash(&TxHasher{})] = tx
	}

	bc.stateLock.Unlock()

	fmt.Println("==========>>>ACCOUNT STATE<<<<<===========")
	fmt.Printf("Account state : %+v\n", bc.accountState.accounts)
	fmt.Println("==========>>>ACCOUNT STATE<<<<<===========")

	bc.lock.Lock()

	bc.headers = append(bc.headers, b.Header)
	//adding block to the blockStore
	bc.blockStore[b.Hash(BlockHasher{})] = b
	bc.blockStoreHeight[b.Header.Height+1] = b
	bc.lock.Unlock()
	//updating the chainTip
	bc.ChainTip = b

	bc.logger.Log(
		"msg", "added new block to the chain",
		"hash", b.Hash(BlockHasher{}),
		"height", b.Header.Height,
		"transactions", len(b.Transactions),
	)

	return bc.store.Put(b)
}

func (bc *Blockchain) handleTransferNativeTokens(tx *Transaction) error {
	bc.logger.Log("msg", "trasnfering native tokens between addresses", "from", tx.From, "to", tx.To, "value", tx.Value)

	return bc.accountState.Transfer(tx.From.Address(), tx.To.Address(), tx.Value)
}

func (bc *Blockchain) HasBlock(height uint32) bool {
	fmt.Printf("Height : %d\n", bc.Height())
	return height <= bc.Height()
}

func (bc *Blockchain) SetLogger(l log.Logger) {
	bc.logger = l
}

func isLowerThanTarget(x *big.Int, y *big.Int) int {
	return x.Cmp(y)
}

// calculates the target and NBits value for the next block
func (bc *Blockchain) calcTargetValue(b *Block) (*big.Int, error) {
	currentHeight := bc.Height()
	if currentHeight%HEIGHT_DIVISOR == 0 {
		var comparisonBlock *Block
		var err error
		if (currentHeight - HEIGHT_DIVISOR) == 0 {
			comparisonBlock, err = bc.GetBlock(1)
		} else {
			comparisonBlock, err = bc.GetBlock(currentHeight - HEIGHT_DIVISOR)
		}
		if err != nil {
			return nil, err
		}
		tsCompBlock := comparisonBlock.Header.Timestamp
		tsCurrentBlock, err1 := bc.GetBlock(currentHeight)
		if err1 != nil {
			return nil, err1
		}
		// gives us time Diff in sec
		timeDiff := (tsCurrentBlock.Header.Timestamp - tsCompBlock)
		new_target := compactToTarget(tsCurrentBlock.Header.NBits)
		timeBigInt := new(big.Int).SetUint64(timeDiff)
		actualTimeDiffBigInt := new(big.Int).SetUint64(AVG_TARGET_TIME)
		new_target.Mul(new_target, timeBigInt)
		new_target.Div(new_target, actualTimeDiffBigInt)

		b.Header.NBits = targetToCompact(new_target)
		fmt.Printf("target %064x \n nBit are %v\n", new_target, b.Header.NBits)
		bc.target = new_target
		return new_target, nil
	}

	return nil, nil
}

func compactToTarget(compact uint32) *big.Int {
	// Extract mantissa and exponent
	mantissa := compact & 0x007fffff
	exponent := uint(compact >> 24)

	// Calculate the coefficient (first 3 bytes of mantissa)
	coefficient := mantissa & 0xffffff

	// Calculate the target value
	target := new(big.Int).SetUint64(uint64(coefficient))
	target.Lsh(target, uint(8*(exponent-3)))

	return target
}

func targetToCompact(target *big.Int) uint32 {
	// Convert target to bytes
	targetBytes := target.Bytes()

	// Check if the first byte of targetBytes is greater than 0x7f
	prependZero := false
	if len(targetBytes) > 0 && targetBytes[0] > 0x7f {
		prependZero = true
	}

	// Prepend a zero byte if necessary
	if prependZero {
		targetBytes = append([]byte{0}, targetBytes...)
	}

	// Prepend the length of the byte slice
	targetBytes = append([]byte{byte(len(targetBytes))}, targetBytes...)

	// Right-pad with zeros if there are less than 4 bytes
	for len(targetBytes) < 4 {
		targetBytes = append(targetBytes, 0)
	}

	// Only keep 2 bytes of precision
	targetBytes = targetBytes[:4]

	// Convert the byte slice to uint32
	bits := binary.BigEndian.Uint32(targetBytes)

	return bits
}

func (bc *Blockchain) MineBlock(b *Block) error {
	var targetForBlock *big.Int
	var err error
	if (bc.Height() % HEIGHT_DIVISOR) == 0 {
		targetForBlock, err = bc.calcTargetValue(b)
		if err != nil {
			return err
		}
	}
	targetForBlock = bc.target
	bHash := b.HashWithoutCache(BlockHasher{})
	hashBigInt, _ := new(big.Int).SetString(bHash.String(), 16)
	for isLowerThanTarget(hashBigInt, targetForBlock) != -1 {
		nonce := b.Header.Nonce
		b.Header.Nonce++
		bHash = b.HashWithoutCache(BlockHasher{})
		hashBigInt.SetString(bHash.String(), 16)
		fmt.Printf("trying new combo with nonce %v block hash %s and target %x \n", nonce, bHash.String(), targetForBlock)
	}

	// updating timestamp
	b.Header.Timestamp = uint64(time.Now().UnixNano())
	b.Header.Target = targetForBlock
	b.Header.NBits = targetToCompact(targetForBlock)

	fmt.Printf("block mined with hash %s and target %x \n", bHash.String(), targetForBlock)

	return nil
}
