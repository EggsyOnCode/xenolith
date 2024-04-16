package core

import (
	"fmt"
	"os"
	"testing"

	"github.com/EggsyOnCode/xenolith/core_types"
	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
)

func TestSendNativeTokenTransferSuccess(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	a := bc.accountState
	// teh validator's priv key
	signer := crypto_lib.GeneratePrivateKey()

	block := randomBlock(t, 1, getPrevBlockHash(t, bc, 1))

	pkAlice := crypto_lib.GeneratePrivateKey()
	addrAlice := pkAlice.PublicKey().Address()
	accAlice := a.CreateAccount(addrAlice)
	pkBob := crypto_lib.GeneratePrivateKey()
	accBob := a.CreateAccount(pkBob.PublicKey().Address())
	fmt.Printf("Alice: %v\n", accAlice.Address)
	fmt.Printf("Bob: %v\n", accBob.Address)

	a.accounts[addrAlice].Balance = uint64(150)

	tx := NewTransaction([]byte{})
	tx.From = pkAlice.PublicKey()
	tx.To = pkBob.PublicKey()
	tx.Value = uint64(100)
	tx.Sign(pkAlice)

	assert.Nil(t, block.AddTx(tx))
	assert.Nil(t, block.Sign(signer))
	newDataHash, _ := CalculateDataHash(block.Transactions)
	assert.Equal(t, block.Header.DataHash, newDataHash)

	// add the block to the blockchain
	assert.Nil(t, bc.AddBlock(block))
	assert.Equal(t, bc.accountState.accounts[addrAlice].Address, addrAlice)
	assert.Equal(t, bc.accountState.accounts[pkBob.PublicKey().Address()].Address, pkBob.PublicKey().Address())
	assert.Equal(t, bc.accountState.accounts[addrAlice].Balance, uint64(50))
}

func TestSendNativeTokenTransferHackingAttempt(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	a := bc.accountState
	// teh validator's priv key
	signer := crypto_lib.GeneratePrivateKey()

	block := randomBlock(t, 1, getPrevBlockHash(t, bc, 1))

	pkAlice := crypto_lib.GeneratePrivateKey()
	addrAlice := pkAlice.PublicKey().Address()
	accAlice := a.CreateAccount(addrAlice)
	pkBob := crypto_lib.GeneratePrivateKey()
	accBob := a.CreateAccount(pkBob.PublicKey().Address())
	fmt.Printf("Alice: %v\n", accAlice.Address)
	fmt.Printf("Bob: %v\n", accBob.Address)

	a.accounts[addrAlice].Balance = uint64(150)

	tx := NewTransaction([]byte{})
	tx.From = pkAlice.PublicKey()
	tx.To = pkBob.PublicKey()
	tx.Value = uint64(100)
	tx.Sign(pkAlice)

	assert.Nil(t, block.Sign(signer))
	newDataHash, _ := CalculateDataHash(block.Transactions)
	assert.Equal(t, block.Header.DataHash, newDataHash)

	// the tx is intercepted by a bad actor
	hackerPk := crypto_lib.GeneratePrivateKey()
	tx.To = hackerPk.PublicKey()

	assert.Equal(t, tx.To.Address(), hackerPk.PublicKey().Address())
	assert.Nil(t, block.AddTx(tx))
	assert.Nil(t, bc.AddBlock(block))
	assert.Equal(t, bc.accountState.accounts[hackerPk.PublicKey().Address()].Balance, uint64(100))
}
func TestBlockchain(t *testing.T) {
	//genesis block
	bc := newBlockchainWithGenesis(t)
	assert.Equal(t, bc.Height(), uint32(0))
}

func TestHasBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	assert.True(t, bc.HasBlock(0))
	assert.False(t, bc.HasBlock(100))
}

func TestAddBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	lenB := 2
	for i := 0; i < lenB; i++ {
		prevHash := getPrevBlockHash(t, bc, uint32(i+1))
		block := randomBlockWithSignature(t, uint32(i+1), (prevHash))
		err := bc.AddBlock(block)
		assert.Nil(t, err)
	}

	assert.Equal(t, bc.Height(), uint32(lenB))
	assert.Equal(t, len(bc.headers), lenB+1)
	//since the current height of the blockchain is 1000, adding a block with height 89 should return an error
	assert.NotNil(t, bc.AddBlock(randomBlockWithSignature(t, 89, core_types.Hash{})))
}

func TestBlockTooHigh(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	block := randomBlockWithSignature(t, 1000, core_types.Hash{})
	assert.NotNil(t, bc.AddBlock(block))
	block1 := randomBlockWithSignature(t, 1, core_types.Hash{})
	assert.NotNil(t, bc.AddBlock(block1))
}
func TestGetHeaders(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	for i := 0; i < 2; i++ {
		prevHash := getPrevBlockHash(t, bc, uint32(i+1))
		block := randomBlockWithSignature(t, uint32(i+1), prevHash)
		err := bc.AddBlock(block)
		assert.Nil(t, err)
		header, err := bc.GetHeaders(block.Header.Height)
		assert.Nil(t, err)
		assert.Equal(t, header, block.Header)
	}

}

func TestGetBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	lenB := 10
	for i := 1; i < lenB; i++ {
		prevHash := getPrevBlockHash(t, bc, uint32(i))
		block := randomBlockWithSignature(t, uint32(i), (prevHash))
		err := bc.AddBlock(block)
		assert.Nil(t, err)
		block1, err1 := bc.GetBlock(block.Header.Height)
		assert.Nil(t, err1)
		assert.Equal(t, block, block1)
	}
}

func newBlockchainWithGenesis(t *testing.T) *Blockchain {
	block := randomBlockWithSignature(t, 1, core_types.Hash{})
	logger := log.NewLogfmtLogger(os.Stderr)
	bc, err := NewBlockchain(block, logger)
	assert.Nil(t, err)
	return bc
}

func getPrevBlockHash(t *testing.T, bc *Blockchain, height uint32) core_types.Hash {
	prevHeader, err := bc.GetHeaders(height - 1)
	assert.Nil(t, err)
	return BlockHasher{}.Hash(prevHeader)
}

// func TestGetGenesisBlock(t *testing.T) {
// 	bc := newBlockchainWithGenesis(t)
// 	block, err := bc.GetBlock(0)
// 	assert.Nil(t, err)
// 	assert.Equal(t, block.Header.Height, uint32(1))
// }
