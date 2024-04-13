package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/EggsyOnCode/xenolith/core"
	"github.com/go-kit/log"
	"github.com/labstack/echo/v4"
)

// ///helper types for user friendly transforamtions
type Block struct {
	Version       uint32
	DataHash      string
	PrevBlockHash string
	Height        uint32
	Timestamp     string
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
	if err == nil {
		block, err := s.bc.GetBlock(uint32(height))
		if err != nil {
			return c.JSON(http.StatusNotFound, err)
		}

		//covnerting block to user friendly format jsonBlock
		jsonBlock := &Block{
			Version:       block.Header.Version,
			DataHash:      block.Header.DataHash.String(),
			PrevBlockHash: block.Header.PrevBlockHash.String(),
			Height:        block.Header.Height,
			Timestamp:     time.Unix(int64(block.Header.Timestamp), 0).String(),
		}
		return c.JSON(http.StatusOK, jsonBlock)
	}
	panic("not implemented getBlockByHash!!!")
}
