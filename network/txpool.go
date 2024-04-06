package network

import (
	"sort"

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

func (p *TxPool) Transactions() []*core.Transaction {
	s := NewTxMapSorter(p.transactions)
	return s.transactions
}

func (p *TxPool) Has(hash core_types.Hash) bool {
	_, ok := p.transactions[hash]
	return ok
}

// TxMapSorter implements the sort.Interface for []Transaction
// we need overrides for the Len, Less and Swap methods
type TxMapSorter struct {
	transactions []*core.Transaction
}

func NewTxMapSorter(txx map[core_types.Hash]*core.Transaction) *TxMapSorter {
	tMap := make([]*core.Transaction, len(txx))

	i := 0
	for _, tx := range txx {
		tMap[i] = tx
		i++
	}

	s := &TxMapSorter{
		transactions: tMap,
	}

	sort.Sort(s)
	return s
}

func (s *TxMapSorter) Len() int {
	return len(s.transactions)
}

func (s *TxMapSorter) Less(i, j int) bool {
	return s.transactions[i].TimeStamp() < s.transactions[j].TimeStamp()
}

//syntax of swapping two numbers in Go
func (s *TxMapSorter) Swap(i, j int) {
	s.transactions[i], s.transactions[j] = s.transactions[j], s.transactions[i]
}