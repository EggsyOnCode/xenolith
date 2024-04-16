package api

import (
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/EggsyOnCode/xenolith/core"
	"github.com/EggsyOnCode/xenolith/core_types"
	"github.com/go-kit/log"
	"github.com/labstack/echo/v4"
)

// ///helper types for user friendly transformations
type TxResponse struct {
	TxCount  uint32
	TxHashes []string
}
type Block struct {
	Hash          string
	Version       uint32
	DataHash      string
	PrevBlockHash string
	Height        uint32
	Timestamp     string
	Validator     string
	Signature     string
	*TxResponse
}

type Transaction struct {
	Data      []byte
	From      string
	Signature string
	TimeStamp string
	Hash      string
	Nonce     int64
}

///////////////////

type APIError struct {
	Error string
}
type ServerConfig struct {
	ListenAddr string
	Logger     log.Logger
}

type Server struct {
	ServerConfig
	bc     *core.Blockchain
	txChan chan *core.Transaction
}

func NewAPIServer(cfg ServerConfig, bc *core.Blockchain, ch chan *core.Transaction) *Server {
	return &Server{
		ServerConfig: cfg,
		bc:           bc,
		txChan:       ch,
	}
}

func (s *Server) Start() error {
	echo := echo.New()
	echo.GET("/blocks/:hashID", s.handleGetBlock)
	echo.GET("/tx/:txHash", s.handleGetTx)
	echo.POST("/tx", s.handlePostTx)

	return echo.Start(s.ListenAddr)
}

func (s *Server) handleGetBlock(c echo.Context) error {
	hashID := c.Param("hashID")

	height, err := strconv.Atoi(hashID)
	//we are assuming that the height would be given as a default
	if err == nil {
		block, err := s.bc.GetBlock(uint32(height))
		if err != nil {
			return c.JSON(http.StatusNotFound, APIError{Error: err.Error()})
		}

		return c.JSON(http.StatusOK, intoJsonBlock(block))
	}

	// if not we will move with the assumption that its a hash
	h, err := hex.DecodeString(hashID)
	if err != nil {
		return c.JSON(http.StatusNotFound, APIError{Error: err.Error()})
	}

	block, err := s.bc.GetBlockByHash(core_types.HashFromBytes(h))
	if err != nil {
		return c.JSON(http.StatusNotFound, APIError{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, intoJsonBlock(block))
}

func (s *Server) handleGetTx(c echo.Context) error {
	txHash := c.Param("txHash")

	h, err := hex.DecodeString(txHash)
	if err != nil {
		return c.JSON(http.StatusNotFound, APIError{Error: err.Error()})
	}

	tx, err := s.bc.GetTxByHash(core_types.HashFromBytes(h))
	if err != nil {
		return c.JSON(http.StatusNotFound, APIError{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, intoJsonTx(tx))
}

func (s *Server) handlePostTx(c echo.Context) error {
	tx := new(core.Transaction)
	if err := tx.Decode(core.NewGobTxDecoder(c.Request().Body)); err != nil {
		return c.JSON(http.StatusNotFound, APIError{Error: err.Error()})
	}

	s.txChan <- tx

	return nil
}

// //utils
func intoJsonBlock(block *core.Block) *Block {
	txResponse := &TxResponse{
		TxCount:  uint32(len(block.Transactions)),
		TxHashes: make([]string, 0, len(block.Transactions)),
	}

	for i := 0; i < len(block.Transactions); i++ {
		txResponse.TxHashes = append(txResponse.TxHashes, block.Transactions[i].Hash(core.TxHasher{}).String())
	}

	return &Block{
		Hash:          block.Hash(core.BlockHasher{}).String(),
		Version:       block.Header.Version,
		DataHash:      block.Header.DataHash.String(),
		PrevBlockHash: block.Header.PrevBlockHash.String(),
		Height:        block.Header.Height,
		Timestamp:     time.Unix(int64(block.Header.Timestamp), 0).String(),
		Validator:     block.Validator.Address().String(),
		Signature:     block.Signature.String(),
		TxResponse:    txResponse,
	}

}

func intoJsonTx(tx *core.Transaction) *Transaction {
	return &Transaction{
		//TODO: find a more graceful way to send out data byte slice
		Data:      (tx.Data),
		From:      tx.From.Address().String(),
		Signature: tx.Signature.String(),
		Hash:      tx.Hash(core.TxHasher{}).String(),
		TimeStamp: time.Unix((tx.TimeStamp()), 0).String(),
		Nonce:     tx.Nonce,
	}
}
