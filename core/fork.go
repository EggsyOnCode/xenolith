package core

import (
	"fmt"

	"github.com/EggsyOnCode/xenolith/core_types"
)

type Fork struct {
	//the block at the tip of the chian fork
	ChainTip       core_types.Hash
	ForkingBlock   core_types.Hash
	blocks         []*Block
	Confirmations  uint32
	IsLongestChain bool
}

type ForkPair struct {
	Forks           []*Fork
	ProcessingQueue []*Block
	blockStore      map[core_types.Hash]*Block
}

func NewForkPair(forks []*Fork) *ForkPair {
	return &ForkPair{
		Forks:           forks,
		ProcessingQueue: make([]*Block, 0),
		blockStore:      make(map[core_types.Hash]*Block),
	}
}

// Adds Block to Processing Queue
func (f *ForkPair) AddBlockToProcessingQ(block *Block) {
	f.ProcessingQueue = append(f.ProcessingQueue, block)
}

func (f *ForkPair) AddBlock(fo *Fork, b *Block) {
	for _, fork := range f.Forks {
		if fork == fo {
			fork.blocks = append(fork.blocks, b)
			f.blockStore[b.Hash(BlockHasher{})] = b
		}
	}
}

func (f *ForkPair) GetBlock(hash core_types.Hash) (*Block, error) {
	if block, ok := f.blockStore[hash]; ok {
		return block, nil
	}
	return nil, fmt.Errorf("block not found in the fork pair")
}

// fork ID --> slice of competing forks
type ForkSlice map[uint32]*ForkPair

// {
// 1: [
// fork : {chiantip Block + confirmations + starting block (forking block)},
// fork : {chaintip Block + confirmations + forking block()}
// ],
// ...
// }
// total Chaintips: 2x of the number of forks Id

// returns which fork a block belongs to
func (f *ForkSlice) FindBlockFork(hash core_types.Hash) (*Fork, error) {
	for _, forks := range *f {
		for _, fork := range forks.Forks {
			if fork.ChainTip == hash {
				return fork, nil
			}
		}
	}
	return nil, fmt.Errorf("block not found in the fork slice")
}

// checks if a block exists in any of the forkPairs
func (f *ForkSlice) FindBlock(hash core_types.Hash) (*Block, error) {
	for _, forkPair := range *f {
		if block, ok := forkPair.blockStore[hash]; ok {
			return block, nil
		}
	}
	return nil, fmt.Errorf("block not found in the blocStore")
}

func (f *ForkSlice) FindForkPairId(fo *Fork) uint32 {
	for id, forkPair := range *f {
		for _, fork := range forkPair.Forks {
			if fork == fo {
				return id
			}
		}
	}
	return 0
}

func (f *ForkSlice) GetForkPair(fo *Fork) *ForkPair {
	for _, forkPair := range *f {
		for _, fork := range forkPair.Forks {
			if fork == fo {
				return forkPair
			}
		}
	}
	return nil
}

func (f *ForkSlice) RemoveForkPair(fp *ForkPair) {
	for id, forkPair := range *f {
		if forkPair == fp {
			delete(*f, id)
		}
	}
}

func (f *ForkSlice) GetBlockByHash(hash core_types.Hash) (*Block, error) {
	for _, forks := range *f {
		for _, block := range forks.ProcessingQueue {
			if block.Hash(BlockHasher{}) == hash {
				return block, nil
			}
		}
	}
	return nil, fmt.Errorf("block not found in the fork slice")
}
