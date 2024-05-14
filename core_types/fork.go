package core_types

import "fmt"

type Fork struct {
	//the block at the tip of the chian fork
	ChainTip      Hash
	ForkingBlock  Hash
	Confirmations uint32
}

// fork ID --> slice of competing forks
type ForkSlice map[uint32][]*Fork
// {
// 1: [
// fork : {chiantip Block + confirmations + starting block (forking block)},
// fork : {chaintip Block + confirmations + forking block()}
// ],
// ...
// }
// total Chaintips: 2x of the number of forks Id


func (f *ForkSlice) FindBlock(hash Hash) (*Fork, error) {
	for _, forks := range *f {
		for _, fork := range forks {
			if fork.ChainTip == hash {
				return fork, nil
			}
		}
	}
	return nil, fmt.Errorf("block not found in the fork slice")
}
