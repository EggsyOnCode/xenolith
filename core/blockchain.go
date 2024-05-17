package core

import (
	"fmt"
	"math"
	"math/big"
	"sync"

	"github.com/EggsyOnCode/xenolith/core_types"
	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/go-kit/log"
)

const (
	HEIGHT_DIVISOR = 5
	// 60 sec
	AVG_TARGET_TIME = 1 * 60
	TARGET_GENESIS  = 0x00ffff0000000000000000000000000000000000000000000000000000
)

type Blockchain struct {
	Version uint32
	logger  log.Logger

	lock    sync.Mutex
	headers []*Header
	// blocks  []*Block

	forkLock  *sync.RWMutex
	forkCount uint32
	core_types.ForkSlice
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
		block:            genesis,
		ForkSlice:        make(core_types.ForkSlice),
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

// checks to see if the new incoming block is causing any forks in the chain
// or is attaching itself to a particular forked block etc
func (bc *Blockchain) handleAndTrackForks(b *Block) (bool, error) {
	// 1st check: if the block is causing a fork --> insertion in forkSlice
	// false indicates that the block is not causing a fork or becoming part of a fork
	if len(b.PrevBlock.NextBlocks) >= 1 {
		forkingFork := &core_types.Fork{
			ChainTip:      b.Hash(BlockHasher{}),
			ForkingBlock:  b.PrevBlock.Hash(BlockHasher{}),
			Confirmations: 0,
		}

		competitorFork := &core_types.Fork{
			ChainTip:      b.PrevBlock.NextBlocks[0].Hash(BlockHasher{}),
			ForkingBlock:  b.PrevBlock.Hash(BlockHasher{}),
			Confirmations: 0,
		}

		forks := []*core_types.Fork{forkingFork, competitorFork}

		bc.forkLock.Lock()
		bc.forkCount++
		bc.ForkSlice[bc.forkCount] = forks
		bc.forkLock.Unlock()

		return true, nil
	}
	// 2nd check: if the block is attaching itself to a forked block (a block in a fork chain temp or otherwise)--> update the forkSlice

	fork, err := bc.ForkSlice.FindBlock(b.Header.PrevBlockHash)
	if err != nil {
		return false, err
	}

	fork.Confirmations++
	blockAtForkTip, err := bc.GetBlockByHash(fork.ChainTip)
	if err != nil {
		return true, err
	}

	blockAtForkTip.NextBlocks = append(blockAtForkTip.NextBlocks, b)
	fork.ChainTip = b.Hash(BlockHasher{})

	return true, nil
}

func (bc *Blockchain) addBlockWithoutValidation(b *Block) error {
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
	// bc.block itself is the head ptr
	// adding the new block to the next block of the current block

	isFork := true
	//handle forks if present
	if b.Header.Height > 1 {
		isFork, _ = bc.handleAndTrackForks(b)
	}
	// if the block is not a fork then we can add the block normally to the ll
	if !isFork {
		bc.block.NextBlocks = append(bc.block.NextBlocks, b)
		bc.block = b
	}

	//adding block to the blockStore
	bc.blockStore[b.Hash(BlockHasher{})] = b
	bc.blockStoreHeight[b.Header.Height+1] = b
	bc.lock.Unlock()

	bc.store.Put(b)

	bc.logger.Log(
		"msg", "added new block to the chain",
		"hash", b.Hash(BlockHasher{}),
		"height", b.Header.Height,
		"transactions", len(b.Transactions),
	)

	return bc.store.Put(b)
}

func (bc *Blockchain) mineBlock(b *Block) error {
	b.Hash(BlockHasher{})
	return nil
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

func (bc *Blockchain) calcTargetValue(b *Block) (string, error) {
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
			return "", err
		}
		tsCompBlock := comparisonBlock.Header.Timestamp
		tsCurrentBlock, err1 := bc.GetBlock(currentHeight)
		if err1 != nil {
			return "", err1
		}
		// gives us time Diff in sec
		timeDiff := (tsCurrentBlock.Header.Timestamp - tsCompBlock)
		// diff of the next block (the to-be-processed block)
		next_diff := tsCurrentBlock.Header.Difficulty * AVG_TARGET_TIME / float64(timeDiff)
		rounded_next_diff := math.Round(next_diff)
		diff_hash := fmt.Sprintf("%x", math.Float64bits(rounded_next_diff))
		tar, tar_compact := getTargetFromDiff(diff_hash, fmt.Sprintf("%f", tsCurrentBlock.Header.Difficulty))
		fmt.Printf("Target : %s\nTarget Compact : %s\n", tar, tar_compact)
		b.Header.Target, _ = new(big.Int).SetString(tar, 16)
		b.Header.Difficulty = next_diff
		return tar, nil
	}

	return "", nil
}
func getTargetFromDiff(diff string, currentDiff string) (string, string) {
	// Extract exponent and mantissa from the compact format
	exponentHex, mantissaHex := extractExpoAndMantissaFromHex(diff)

	// Convert exponent to a big integer and adjust it
	exponent := new(big.Int)
	exponent.SetString(exponentHex, 16)
	exponent.Sub(exponent, big.NewInt(3))

	// Convert mantissa to a big integer
	mantissa := new(big.Int)
	mantissa.SetString(mantissaHex, 16)

	// Calculate the target: mantissa * (256 ^ exponent)
	base := big.NewInt(256)
	exp := new(big.Int).Exp(base, exponent, nil)
	hTarget := new(big.Int).Mul(mantissa, exp)

	// Convert current difficulty to big.Float
	difficulty := new(big.Float)
	difficulty.SetString(currentDiff)

	// Calculate current target: hTarget / difficulty
	hTargetFloat := new(big.Float).SetInt(hTarget)
	cTargetFloat := new(big.Float).Quo(hTargetFloat, difficulty)

	// Convert cTargetFloat to big.Int and round up
	cTarget := new(big.Int)
	cTargetFloat.Add(cTargetFloat, big.NewFloat(0.999999999)).Int(cTarget)

	// Convert current target to hex
	targetHex := cTarget.Text(16)

	// Convert current target to compact format
	mantissaCompact := targetHex
	if len(targetHex) > 6 {
		mantissaCompact = targetHex[:6]
	}
	exponentCompact := fmt.Sprintf("%02x", len(targetHex)/2+3)
	targetCompact := exponentCompact + mantissaCompact

	return targetHex, targetCompact
}

func extractExpoAndMantissaFromHex(h string) (string, string) {
	if len(h) < 2 {
		fmt.Errorf("Invalid hex string")
		return "", ""
	}

	// Extract the first byte (exponent) and remaining bytes (mantissa)
	firstByteHex := h[:2]
	remainingBytesHex := h[2:]

	return firstByteHex, remainingBytesHex
}
