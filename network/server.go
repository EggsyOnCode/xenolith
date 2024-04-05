package network

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/EggsyOnCode/xenolith/core"
	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/sirupsen/logrus"
)

var defaultBlockTime = 5 * time.Second

type ServerOpts struct {
	RPCHandler   RPCHandler
	Transporters []Transport
	PrivateKey   *crypto_lib.PrivateKey
	//time interval after  which the server will fetch Tx from teh Mempool and create a block
	BlockTime time.Duration
}

type Server struct {
	ServerOpts
	//is the PvK is not nil then the server is a validator
	isValidator bool
	rpcCh       chan RPC
	blocktime   time.Duration
	memPool     *TxPool
	quitCh      chan struct{}
}

func NewServer(opts ServerOpts) *Server {
	if opts.BlockTime == time.Duration(0) {
		opts.BlockTime = defaultBlockTime
	}
	return &Server{
		ServerOpts:  opts,
		blocktime:   opts.BlockTime,
		rpcCh:       make(chan RPC),
		isValidator: opts.PrivateKey != nil,
		quitCh:      make(chan struct{}),
		memPool:     NewTxPool(),
	}
}

func (s *Server) Start() error {
	s.initTransporters()
	ticker := time.NewTicker(5 * time.Second)
	//infinite loop reading the rpc msgs from the transporters
	// free label used
free:
	for {
		select {
		case rpc := <-s.rpcCh:
			fmt.Println("Server received msg from ", rpc.From, " with payload ", readerToString(rpc.Payload))
		case <-s.quitCh:
			break free
		case <-ticker.C:
			if s.isValidator {
				//TODO: consensus logic will be written here
				s.createNewBlock()
			}
		}
	}
	fmt.Println("Server Stopped!")
	return nil
}

func (s *Server) createNewBlock() {
	fmt.Println("Creating a new block")
}

// either the server fetches tx from the mempool or receives Tx from the transporters; this func would handle the tx from both
func (s *Server) handleTx(tx *core.Transaction) error {
	//verify the transaction
	if ans, _ := tx.Verify(); !ans {
		return fmt.Errorf("tx not signed")
	}

	hash := tx.Hash(core.TxHasher{})

	if s.memPool.Has(hash) {
		logrus.WithFields(logrus.Fields{
			"hash": hash,
		}).Info("tx already exists in mempool")
		return nil
	}

	logrus.WithFields(logrus.Fields{
		"hash": hash,
	}).Info("adding new tx to the mempool")

	return s.memPool.Add(tx)
}

func (s *Server) initTransporters() error {
	for _, tr := range s.Transporters {
		// reading the msg channels of each of the connected transportes and piping htem into the server's rpc for faster and safer processing
		go func(tr Transport) {
			for rpc := range tr.Consume() {
				s.rpcCh <- rpc
			}
		}(tr)
	}

	return nil
}

func readerToString(r io.Reader) string {
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(r)
	return buffer.String()
}
