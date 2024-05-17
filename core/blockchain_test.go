package core

import (
	"fmt"
	"math/big"
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
func TestSendNativeTransferInsufficientBalance(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	signer := crypto_lib.GeneratePrivateKey()

	block := randomBlock(t, 1, getPrevBlockHash(t, bc, 1))
	assert.Nil(t, block.Sign(signer))

	privKeyBob := crypto_lib.GeneratePrivateKey()
	privKeyAlice := crypto_lib.GeneratePrivateKey()
	amount := uint64(100)

	accountBob := bc.accountState.CreateAccount(privKeyBob.PublicKey().Address())
	accountBob.Balance = uint64(99)

	tx := NewTransaction([]byte{})
	tx.From = privKeyBob.PublicKey()
	tx.To = privKeyAlice.PublicKey()
	tx.Value = amount
	tx.Sign(privKeyBob)

	fmt.Printf("alice => %s\n", privKeyAlice.PublicKey().Address())
	fmt.Printf("bob => %s\n", privKeyBob.PublicKey().Address())

	block.AddTx(tx)
	assert.Nil(t, bc.AddBlock(block))

	_, err := bc.accountState.GetAccount(privKeyAlice.PublicKey().Address())
	assert.NotNil(t, err)

	// this erroneous tx should not be added to the blockchain
	// therefore the query should return with an error
	hash := tx.Hash(TxHasher{})
	_, err = bc.GetTxByHash(hash)
	assert.NotNil(t, err)
}

func TestSendNativeTokenTransferWithTampering(t *testing.T) {
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
	fmt.Printf("Original To: %v\n", pkBob.PublicKey())

	assert.Nil(t, block.Sign(signer))
	newDataHash, _ := CalculateDataHash(block.Transactions)
	assert.Equal(t, block.Header.DataHash, newDataHash)

	// the tx is intercepted by a bad actor
	hackerPk := crypto_lib.GeneratePrivateKey()
	tx.To = hackerPk.PublicKey()

	fmt.Printf("Hacker: %v\n", hackerPk.PublicKey())

	assert.Equal(t, tx.To.Address(), hackerPk.PublicKey().Address())
	assert.NotNil(t, block.AddTx(tx))
	fmt.Printf("block is signed by: %v\n", block.Validator)
	assert.Nil(t, bc.AddBlock(block))
	// the hacker account won't exist hence hte below code would throw null ptr exception
	// assert.NotNil(t, bc.accountState.accounts[hackerPk.PublicKey().Address()].Balance)
	assert.Equal(t, bc.accountState.accounts[addrAlice].Balance, uint64(150))
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

func TestNBitsToTargetFunc(t *testing.T) {
	compact := uint32(0x1b0404cb)
	ans := compactToTarget((compact))
	expectedAns, _ := new(big.Int).SetString("00000000000404CB000000000000000000000000000000000000000000000000", 16)
	fmt.Printf("expectedTarget is %x\n", expectedAns)
	assert.Equal(t, 0, ans.Cmp(expectedAns), "Expected target does not match")
	comp := targetToCompact(ans)
	assert.Equal(t, comp, compact)
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

func TestForkBlockAddition(t *testing.T) {
	gB, bc := newBlockchainWithGenesisAndReturnsGenesis(t)
	fmt.Print(bc)
	prevHash := getPrevBlockHash(t, bc, uint32(1))
	block := randomBlockWithSignature(t, uint32(1), (prevHash))
	err := bc.AddBlock(block)
	assert.Nil(t, err)
	forkingBlock := randomBlockWithSignatureAndPrevBlock(t, uint32(2), (prevHash), gB)
	err1 := bc.AddBlock(forkingBlock)
	assert.Nil(t, err1)
}

func TestTargetValueForBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	lenB := HEIGHT_DIVISOR*2 + 1
	for i := 1; i < lenB; i++ {
		prevHash := getPrevBlockHash(t, bc, uint32(i))
		block := randomBlockWithSignature(t, uint32(i), (prevHash))
		err := bc.AddBlock(block)
		assert.Nil(t, err)
		block1, err1 := bc.GetBlock(block.Header.Height)
		assert.Nil(t, err1)
		assert.Equal(t, block, block1)
	}

	prevBlock, err := bc.GetBlock(HEIGHT_DIVISOR * 2)
	assert.Nil(t, err)
	block := randomBlockWithSignature(t, uint32(6), prevBlock.Hash(BlockHasher{}))
	expectedTarget, err := bc.calcTargetValue(block)
	assert.Nil(t, err)
	//TODO: write better tests for Testing Target
	assert.NotEmpty(t, expectedTarget)
}

func TestMineBlockFunc(t *testing.T){
	bc := newBlockchainWithGenesis(t)
	lenB := HEIGHT_DIVISOR*2 + 1
	for i := 1; i < lenB; i++ {
		prevHash := getPrevBlockHash(t, bc, uint32(i))
		block := randomBlockWithSignature(t, uint32(i), (prevHash))
		err := bc.AddBlock(block)
		assert.Nil(t, err)
		block1, err1 := bc.GetBlock(block.Header.Height)
		assert.Nil(t, err1)
		assert.Equal(t, block, block1)
	}

	prevBlock, err := bc.GetBlock(HEIGHT_DIVISOR * 2)
	assert.Nil(t, err)
	block := randomBlockWithSignature(t, uint32(6), prevBlock.Hash(BlockHasher{}))
	err1 := bc.MineBlock(block)
	assert.Nil(t, err1)
}

func newBlockchainWithGenesis(t *testing.T) *Blockchain {
	block := genesisBlockWithSig(t, 1, core_types.Hash{})
	logger := log.NewLogfmtLogger(os.Stderr)
	bc, err := NewBlockchain(block, logger)
	assert.Nil(t, err)
	return bc
}

func newBlockchainWithGenesisAndReturnsGenesis(t *testing.T) (*Block, *Blockchain) {
	block := genesisBlockWithSig(t, 1, core_types.Hash{})
	logger := log.NewLogfmtLogger(os.Stderr)
	bc, err := NewBlockchain(block, logger)
	assert.Nil(t, err)
	return block, bc
}

func getPrevBlockHash(t *testing.T, bc *Blockchain, height uint32) core_types.Hash {
	prevHeader, err := bc.GetHeaders(height - 1)
	assert.Nil(t, err)
	return BlockHasher{}.Hash(prevHeader)
}

// func TestGetGenesisBlock(t *testing.T) {
// 	bc := newBlockchainWithGenesis(t)
// 	block, err := bc.GetBlock(1)
// 	assert.Nil(t, err)
// 	assert.Equal(t, block.Header.Height, uint32(1))
// }
