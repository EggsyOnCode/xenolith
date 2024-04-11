package network

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/EggsyOnCode/xenolith/core"
	"github.com/EggsyOnCode/xenolith/core_types"
	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/go-kit/log"
)

var defaultBlockTime = 5 * time.Second

type ServerOpts struct {
	BootStrapNodes []string
	ListenAddr     string
	ID             string
	Logger         log.Logger
	RPCDecodeFunc  RPCDecodeFunc
	RPCProcessor   RPCProcessor
	PrivateKey     *crypto_lib.PrivateKey
	//time interval after  which the server will fetch Tx from teh Mempool and create a block
	BlockTime time.Duration
}

type Server struct {
	ServerOpts
	//is the PvK is not nil then the server is a validator
	peerCh       chan *TCPPeer
	peerMap      map[NetAddr]*TCPPeer
	isValidator  bool
	TCPTransport *TCPTransport
	chain        *core.Blockchain
	rpcCh        chan RPC
	memPool      *TxPool
	quitCh       chan struct{}
}

func NewServer(opts ServerOpts) (*Server, error) {
	if opts.BlockTime == time.Duration(0) {
		opts.BlockTime = defaultBlockTime
	}
	if opts.RPCDecodeFunc == nil {
		opts.RPCDecodeFunc = DefaultRPCDecodeFunc
	}
	if opts.Logger == nil {
		opts.Logger = log.NewLogfmtLogger(os.Stderr)
		opts.Logger = log.With(opts.Logger, "address", opts.ID)
	}

	newChain, err := core.NewBlockchain(genesisBlock(), opts.Logger)
	if err != nil {
		return nil, err
	}

	peerCh := make(chan *TCPPeer)
	tr := NewTCPTransporter(opts.ListenAddr, peerCh)

	s := &Server{
		ServerOpts:   opts,
		TCPTransport: tr,
		peerCh:       peerCh,
		rpcCh:        make(chan RPC),
		peerMap:      make(map[NetAddr]*TCPPeer),
		chain:        newChain,
		isValidator:  opts.PrivateKey != nil,
		quitCh:       make(chan struct{}),
		memPool:      NewTxPool(1000),
	}

	// if the rpc processor has not been defined; then we can assume that the server itself is the processor
	if s.RPCProcessor == nil {
		s.RPCProcessor = s
	}

	if s.isValidator {
		go s.validatorLoop()
	}

	return s, nil
}

func (s *Server) Start() error {
	s.TCPTransport.Start()
	//infinite loop reading the rpc msgs from the transporters
	// free label used

	s.Logger.Log("msg", "accepting TCP connections on", "address", s.ListenAddr)
	if len(s.BootStrapNodes) > 0 {
		go s.bootstrapNodes()

	}
free:
	for {
		select {
		case peer := <-s.peerCh:
			//TODO: add a mutex here !!!!
			s.peerMap[NetAddr(peer.conn.RemoteAddr().String())] = peer
			go peer.readLoop(s.rpcCh)
			fmt.Printf("new peer connected %v\n", peer)
		case rpc := <-s.rpcCh:
			msg, err := DefaultRPCDecodeFunc(rpc)
			if err != nil {
				s.Logger.Log("err", err)
				continue
			}

			switch msg.Data.(type) {
			case *GetBlockMessage:
				fmt.Println("getBlockMessage received by %v from %v", s.ID, msg.From)
			}

			if err := s.RPCProcessor.ProcessMessage(msg); err != nil {
				if err != core.ErrBlockKnown {
					s.Logger.Log("err", err)
				}
			}
		case <-s.quitCh:
			break free
		}
	}
	s.Logger.Log("msg", "server stopping...")
	return nil
}

func (s *Server) validatorLoop() {
	ticker := time.NewTicker(s.BlockTime)

	s.Logger.Log("msg", "server validator loop staring...", "blockTime", s.BlockTime)
	for {
		//whenver the ticker value is decremented
		<-ticker.C
		s.createNewBlock()
	}
}

func (s *Server) bootstrapNodes() {
	//these are outgoing connection to remote nodes/servers
	// we are dialing connections to them and adding them as outgoing peers
	for _, addr := range s.BootStrapNodes {
		conn, err := net.Dial("tcp", (addr))
		if err != nil {
			fmt.Printf("could not connect to %v: %v\n", addr, err)
			continue
		}

		s.peerCh <- NewTCPPeer(conn, true)
	}

}

func (s *Server) createNewBlock() error {
	//fetch current block;s headers
	currentHedaer, err := s.chain.GetHeaders(s.chain.Height())
	if err != nil {
		return err
	}
	// For now are including all hte tx in the mempoool in the block
	// later we can introduce a complexity func to  detemrine how many tx to be
	// include in one block
	txx := s.memPool.Pending()

	block, err := core.NewBlockFromPrevHeader(currentHedaer, txx)
	if err != nil {
		return err
	}

	//signing the block
	block.Sign(s.PrivateKey)

	// TODO: pending pool of tx should only reflect on validator nodes.
	// Right now "normal nodes" do not have their pending pool cleared.
	s.memPool.ClearPending()

	if err := s.chain.AddBlock(block); err != nil {
		return err
	}

	return s.broadcastBlock(block)
}

// func (s *Server) sendGetStatusMsg() error {
// 	msg := new(GetStatusMessage)
// 	buf := &bytes.Buffer{}
// 	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
// 		return err
// 	}
// 	rpcMsg := NewMessage(MessageGetStatusType, buf.Bytes())
// 	for _, peer := range s.Transport.Peers() {
// 		if err := s.Transport.SendMsg(peer.Addr(), rpcMsg.Bytes()); err != nil {
// 			return err
// 		}

// 	}
// 	// s.broadcast(rpcMsg.Bytes(), s.Transporters[0].Addr())
// 	return nil
// }

// process Msg acts as the router routing the deocded msg to their appropriate handlers
func (s *Server) ProcessMessage(msg *DecodedMsg) error {
	switch t := msg.Data.(type) {
	case *core.Transaction:
		//where t is essentially the msg.Data
		return s.processTx(t)
	case *core.Block:
		return s.processBlock(t, msg.From)
	case *StatusMessage:
		// return s.processStatusMsg(msg.From, t)
	case *GetStatusMessage:
		// return s.processGetStatusMsg(msg.From)
	case *GetBlockMessage:
		// return s.processBlockRequestedMsg(msg.From, t)
	}

	return nil
}

// // when the server receives a req from another node to send its status msg
// func (s *Server) processGetStatusMsg(from NetAddr) error {
// 	if from == "TR_LATE" && (s.Transport.Addr() == "Remote_0" || s.Transport.Addr() == "Remote_1") {
// 		fmt.Printf("received GetStatusMsg from TR_LATE by %v\n", s.Transport.Addr())
// 	}
// 	if s.Transport.Addr() != from {
// 		statusMsg := NewStatusMessage(s.chain.Height(), s.chain.Version)
// 		buf := &bytes.Buffer{}
// 		if err := gob.NewEncoder(buf).Encode(statusMsg); err != nil {
// 			return err
// 		}
// 		msg := NewMessage(MessageStatusType, buf.Bytes())
// 		buffer := &bytes.Buffer{}
// 		if err := gob.NewEncoder(buffer).Encode(msg); err != nil {
// 			return err
// 		}
// 		if err := s.Transport.SendMsg(from, buffer.Bytes()); err != nil {
// 			return err
// 		}
// 		if from == "TR_LATE" {
// 			fmt.Printf("sent status msg to TR_LATE from %v\n", s.Transport.Addr())
// 		}
// 	}
// 	// for _, tr := range s.Transporters {
// 	// 	//TODO: a way to fetch the verison of blockchain
// 	// 	continue
// 	// }
// 	return nil
// }

// // when the server receives a status msg response back from the nodes
// func (s *Server) processStatusMsg(from NetAddr, msg *StatusMessage) error {
// 	//log it then
// 	if s.Transport.Addr() == "TR_LATE" {
// 		fmt.Println("TR_LATE received status msg from ", from, "data", msg)
// 	}
// 	fmt.Printf("received status msg %v from => server at address %v\n", msg, from)

// 	// compare the msg data with the node's own bc; its height and version say
// 	// cal the differences adn request the missing bits from the peer
// 	if msg.CurrentHeight <= s.chain.Height() {
// 		s.Logger.Log("msg", "peer is behind", "peer height", msg.CurrentHeight, "our height", s.chain.Height(), "peer address", from)
// 	}

// 	//now we are certain that we are behind peer; so we need to fetch the missing blocks
// 	getBlockMsg := &GetBlockMessage{
// 		From: s.chain.Height() + 1,
// 		//0 would signal the remote node to send max no of blocks that they have
// 		To: 0,
// 	}
// 	buf := &bytes.Buffer{}

// 	if err := gob.NewEncoder(buf).Encode(getBlockMsg); err != nil {
// 		return err
// 	}

// 	rpcMsg := NewMessage(MessageTypeGetBlocks, buf.Bytes())
// 	return s.Transport.SendMsg(from, rpcMsg.Bytes())
// }

// func (s *Server) processBlockRequestedMsg(from NetAddr, msg *GetBlockMessage) error {
// 	// sending all the blocks starting from 'msg.from'
// 	if from == "TR_LATE" {
// 		fmt.Println("TR_LATE sent a get block request")
// 	}
// 	fmt.Printf("server %v received get block request msg from %v\n", s.ID, from)

// 	return nil
// }

func (s *Server) processBlock(b *core.Block, origin NetAddr) error {
	//when the block is received from the peers, we need to add it to the local chain
	//this way the incoming block gets validated as well
	if err := s.chain.AddBlock(b); err != nil {
		return err
	}

	s.Logger.Log("msg", "received block from peers", "block hash", core.BlockHasher{}.Hash(b.Header), "chain height", s.chain.Height())
	go s.broadcastBlock(b)

	return nil
}

// either the server fetches tx from the mempool or receives Tx from the transporters; this func would handle the tx from both
func (s *Server) processTx(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})

	if s.memPool.Contains(hash) {
		return nil
	}

	//verify the transaction
	if ans, _ := tx.Verify(); !ans {
		return fmt.Errorf("tx not signed")
	}

	//setting the timestamp for the incoming tx
	tx.SetTimeStamp(time.Now().Unix())

	// s.Logger.Log("msg", "adding new tx to the mempool", "hash", hash, "memPool pending", s.memPool.PendingCount())

	go s.broadcastTx(tx)

	s.memPool.Add(tx)
	return nil
}

func (s *Server) broadcast(payload []byte) error {
	fmt.Println("broadcasting to peers...")
	for _, peer := range s.peerMap {
		if err := peer.Send(payload); err != nil {
			fmt.Printf("error sending to %v: %v\n", peer, err)
			continue
		}
	}
	return nil
}

// broadcast block to peers to share the updated state of the chain
func (s *Server) broadcastBlock(b *core.Block) error {
	buf := &bytes.Buffer{}
	if err := b.Encode(core.NewGobBlockEncoder(buf)); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeBlock, buf.Bytes())
	return s.broadcast(msg.Bytes())
}

func (s *Server) broadcastTx(tx *core.Transaction) error {
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeTx, buf.Bytes())
	return s.broadcast(msg.Bytes())
}

func readerToString(r io.Reader) string {
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(r)
	return buffer.String()
}

func genesisBlock() *core.Block {
	headers := &core.Header{
		Version:   1,
		Height:    0,
		DataHash:  core_types.Hash{},
		Timestamp: 000000,
	}

	return core.NewBlock(headers, nil)
}
