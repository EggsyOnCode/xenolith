package core

import (
	"fmt"

	"github.com/EggsyOnCode/xenolith/core_types"
)

type Fork struct {
	//the block at the tip of the chian fork
	ChainTip       core_types.Hash
	ForkingBlock   core_types.Hash
	Confirmations  uint32
	IsLongestChain bool
}

type ForkPair struct {
	Forks           []*Fork
	ProcessingQueue []*Block
}

func NewForkPair(forks []*Fork) *ForkPair {
	return &ForkPair{
		Forks:           forks,
		ProcessingQueue: make([]*Block, 0),
	}
}

//Adds Block to Processing Queue
func (f *ForkPair) AddBlock(block *Block) {
	f.ProcessingQueue = append(f.ProcessingQueue, block)
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
func (f *ForkSlice) FindBlock(hash core_types.Hash) (*Fork, error) {
	for _, forks := range *f {
		for _, fork := range forks.Forks {
			if fork.ChainTip == hash {
				return fork, nil
			}
		}
	}
	return nil, fmt.Errorf("block not found in the fork slice")
}

func (f *ForkSlice) FindForkPairId(fork *Fork) uint32 {
	for id, forkPair := range *f {
		for _, fork := range forkPair.Forks {
			if fork == fork {
				return id
			}
		}
	}
	return 0
}
