package network

import (
	"fmt"

	"github.com/EggsyOnCode/xenolith/core"
	"github.com/EggsyOnCode/xenolith/core_types"
)

type TxPool struct {
	//mapping of transaction hash to transaction
	transactions map[core_types.Hash]*core.Transaction
}

func NewTxPool() *TxPool {
	return &TxPool{
		transactions: make(map[core_types.Hash]*core.Transaction),
	}
}

func (p *TxPool) Len() int {
	return (len(p.transactions))
}

func (p *TxPool) Flush() {
	p.transactions = make(map[core_types.Hash]*core.Transaction)
}

// we are assuming that we are receiving a verified tx?
func (p *TxPool) Add(tx *core.Transaction) error {
	//verify the transaction
	if ans, _:=  tx.Verify();!ans {
		return fmt.Errorf("tx not signed")
	}

	if p.Has(tx.Hash(core.TxHasher{})) {
		return nil
	}
	p.transactions[tx.Hash(core.TxHasher{})] = tx

	return nil
}

func (p *TxPool) Has(hash core_types.Hash) bool {
	_, ok := p.transactions[hash]
	return ok
}

