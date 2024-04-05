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
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	Transporters  []Transport
	PrivateKey    *crypto_lib.PrivateKey
	//time interval after  which the server will fetch Tx from teh Mempool and create a block
	BlockTime time.Duration
}

type Server struct {
	ServerOpts
	//is the PvK is not nil then the server is a validator
	isValidator bool
	rpcCh       chan RPC
	memPool     *TxPool
	quitCh      chan struct{}
}

func NewServer(opts ServerOpts) *Server {
	if opts.BlockTime == time.Duration(0) {
		opts.BlockTime = defaultBlockTime
	}
	if opts.RPCDecodeFunc == nil {
		opts.RPCDecodeFunc = DefaultRPCDecodeFunc
	}
	s := &Server{
		ServerOpts:  opts,
		rpcCh:       make(chan RPC),
		isValidator: opts.PrivateKey != nil,
		quitCh:      make(chan struct{}),
		memPool:     NewTxPool(),
	}

	// if the rpc processor has not been defined; then we can assume that the server itself is the processor
	if s.RPCProcessor == nil {
		s.RPCProcessor = s
	}

	return s
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
			msg, err := DefaultRPCDecodeFunc(rpc)
			if err != nil {
				logrus.Error(err)
			}

			if err := s.RPCProcessor.ProcessMessage(msg); err != nil {
				logrus.Error(err)
			}
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

//process Msg acts as the router routing the deocded msg to their appropriate handlers
func (s *Server) ProcessMessage(msg *DecodedMsg) error {
	switch t := msg.Data.(type) {
	case *core.Transaction:
		//where t is essentially the msg.Data
		return s.processTx(t)
	}

	return nil
}

// either the server fetches tx from the mempool or receives Tx from the transporters; this func would handle the tx from both
func (s *Server) processTx(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})

	if s.memPool.Has(hash) {
		logrus.WithFields(logrus.Fields{
			"hash": hash,
			"from": tx.From,
		}).Info("tx already exists in mempool")
		return nil
	}

	//verify the transaction
	if ans, _ := tx.Verify(); !ans {
		return fmt.Errorf("tx not signed")
	}

	//setting the timestamp for the incoming tx
	tx.SetTimeStamp(time.Now().Unix())

	logrus.WithFields(logrus.Fields{
		"hash": hash,
		"from": tx.From,
	}).Info("adding new tx to the mempool")

	go s.broadcastTx(tx)

	return s.memPool.Add(tx)
}

func (s *Server) broadcast(payload []byte) error {
	for _, tr := range s.Transporters {
		if err := tr.Broadcast(payload); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) broadcastTx(tx *core.Transaction) error {
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeTx, buf.Bytes())
	return s.broadcast(msg.Bytes())
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
