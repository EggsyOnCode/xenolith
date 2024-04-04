package network

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/EggsyOnCode/xenolith/core"
	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/stretchr/testify/assert"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// RandString generates a random string of length n.
func RandString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
func randomTx() *core.Transaction {
	tx := core.NewTransaction([]byte(RandString(10)))
	tx.SetTimeStamp(time.Now().Unix())
	return tx
}

func TestTxPoolAddTx(t *testing.T) {
	p := NewTxPool()
	tx := randomTx()
	pvK := crypto_lib.GeneratePrivateKey()
	fmt.Println(pvK.PublicKey())
	tx.Sign(pvK)

	//testing if the tx has been signed
	assert.True(t, tx.Signature != nil)
	fmt.Println(tx.From)

	//adding verified tx to the mempool
	assert.Nil(t, p.Add(tx))

	//checking if the tx is present in the mempool
	assert.True(t, p.Has(tx.Hash(core.TxHasher{})))

	//checking if the length of the mempool is 1
	assert.Equal(t, 1, p.Len())

	//flushing the mempool
	p.Flush()
	assert.Equal(t, 0, p.Len())
}

func TestTxPoolSorter(t *testing.T) {
	p := NewTxPool()
	count := 100
	for i := 0; i < count; i++ {
		tx := randomTx()
		pvK := crypto_lib.GeneratePrivateKey()
		tx.Sign(pvK)
		assert.Nil(t, p.Add(tx))
	}

	tx2 := randomTx()
	pvK := crypto_lib.GeneratePrivateKey()
	tx2.Sign(pvK)
	assert.Nil(t, p.Add(tx2))
	//sorting the txs
	s := NewTxMapSorter(p.transactions)
	assert.Equal(t, count+1, len(s.transactions))

	//sorting the txs
	sort.Sort(s)
	assert.Equal(t, tx2.TimeStamp(), s.transactions[count].TimeStamp())
}
