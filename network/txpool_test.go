package network

import (
	"fmt"
	"testing"

	"github.com/EggsyOnCode/xenolith/core"
	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/stretchr/testify/assert"
)

func randomTx() *core.Transaction {
	return core.NewTransaction([]byte("random data"))
}


func TestTxPoolAddTx(t *testing.T){
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

// func TestDuplicateTxAddingToPool(t *testing.T){
// 	p := NewTxPool()
// 	tx := randomTx()
// 	pvK := crypto_lib.GeneratePrivateKey()
// 	tx.Sign(pvK)

// 	//adding verified tx to the mempool
// 	assert.Nil(t, p.Add(tx))

// 	//checking if the tx is present in the mempool
// 	assert.True(t, p.Has(tx.Hash(core.TxHasher{})))

// 	//checking if the length of the mempool is 1
// 	assert.Equal(t, 1, p.Len())

// 	//adding a duplicate tx to the mempool
// 	tx1 := core.NewTransaction([]byte("random data"))
// 	tx1.Sign(pvK)
// 	assert.Nil(t, p.Add(tx1))

// 	// ensuring that the duplicate tx was not added to the pool
// 	assert.Equal(t, 1, p.Len())
// }
// func TestUnSignedTxAddingToPool(t *testing.T){
// 	p := NewTxPool()
// 	tx := randomTx()

// 	//adding verified tx to the mempool
// 	assert.NotNil(t, p.Add(tx))

// }