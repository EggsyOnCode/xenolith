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

// ///helper types for user friendly transforamtions
type Block struct {
	Hash          string
	Version       uint32
	DataHash      string
	PrevBlockHash string
	Height        uint32
	Timestamp     string
	Validator     string
	Signature     string
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
	bc *core.Blockchain
}

func NewAPIServer(cfg ServerConfig, bc *core.Blockchain) *Server {
	return &Server{
		ServerConfig: cfg,
		bc:           bc,
	}
}

func (s *Server) Start() error {
	echo := echo.New()
	echo.GET("/blocks/:hashID", s.handleGetBlock)

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

func intoJsonBlock(block *core.Block) *Block {
	return &Block{
		Hash:          block.Hash(core.BlockHasher{}).String(),
		Version:       block.Header.Version,
		DataHash:      block.Header.DataHash.String(),
		PrevBlockHash: block.Header.PrevBlockHash.String(),
		Height:        block.Header.Height,
		Timestamp:     time.Unix(int64(block.Header.Timestamp), 0).String(),
		Validator:     block.Validator.Address().String(),
		Signature:     block.Signature.String(),
	}

}
