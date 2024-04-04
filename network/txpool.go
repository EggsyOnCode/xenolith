package network

import (
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

// Add adds a transaction to the pool; the responsibilty of checking if the tx already exists in the mempool is server's
func (p *TxPool) Add(tx *core.Transaction) error {
	p.transactions[tx.Hash(core.TxHasher{})] = tx

	return nil
}

func (p *TxPool) Has(hash core_types.Hash) bool {
	_, ok := p.transactions[hash]
	return ok
}
