package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
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
	peerCh chan *TCPPeer

	mu      *sync.RWMutex
	peerMap map[NetAddr]*TCPPeer

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
		mu:           &sync.RWMutex{},
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
			s.mu.RLock()
			s.peerMap[NetAddr(peer.conn.RemoteAddr().String())] = peer
			s.mu.RUnlock()

			go peer.readLoop(s.rpcCh)

			if err := s.sendGetStatusMsg(peer); err != nil {
				s.Logger.Log("err", err)
				continue
			}

			s.Logger.Log("msg", "peer added to the server", "peer", peer.conn.RemoteAddr(), "outgoing", peer.Outgoing)

		case rpc := <-s.rpcCh:
			msg, err := DefaultRPCDecodeFunc(rpc)
			if err != nil {
				s.Logger.Log("err", err)
				continue
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

// process Msg acts as the router routing the deocded msg to their appropriate handlers
func (s *Server) ProcessMessage(msg *DecodedMsg) error {
	switch t := msg.Data.(type) {
	case *core.Transaction:
		//where t is essentially the msg.Data
		return s.processTx(t)
	case *core.Block:
		return s.processBlock(t, msg.From)
	case *StatusMessage:
		return s.processStatusMsg(msg.From, t)
	case *GetStatusMessage:
		return s.processGetStatusMsg(msg.From)
	case *GetBlockMessage:
		return s.processBlockRequestedMsg(msg.From, t)
	case *BlocksMessage:
		return s.processBlockReceipt(msg.From, t)
	}

	return nil
}

// send req msgs to the peer that;s just been added to teh peermap
func (s *Server) sendGetStatusMsg(peer *TCPPeer) error {
	s.Logger.Log("msg", "sending get status msg request to ", "to", peer.conn.RemoteAddr(), "us", s.ListenAddr)
	msg := new(GetStatusMessage)
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	rpcMsg := NewMessage(MessageGetStatusType, buf.Bytes())
	return peer.Send(rpcMsg.Bytes())
}

// // when the server receives a req from another node to send its status msg
func (s *Server) processGetStatusMsg(from NetAddr) error {
	s.Logger.Log("server", s.ID, "msg", "received get status msg request from ", "from", from)
	if s.TCPTransport.listenAddr != from {
		statusMsg := NewStatusMessage(s.chain.Height(), s.chain.Version)
		buf := &bytes.Buffer{}
		if err := gob.NewEncoder(buf).Encode(statusMsg); err != nil {
			return err
		}
		msg := NewMessage(MessageStatusType, buf.Bytes())
		buffer := &bytes.Buffer{}
		if err := gob.NewEncoder(buffer).Encode(msg); err != nil {
			return err
		}

		s.mu.RLock()
		defer s.mu.RUnlock()

		peer, ok := s.peerMap[from]
		if !ok {
			return fmt.Errorf("peer %s not found", peer.conn.RemoteAddr())
		}

		return peer.Send(msg.Bytes())
	}
	return nil
}

// when the server receives a status msg response back from the nodes
func (s *Server) processStatusMsg(from NetAddr, msg *StatusMessage) error {
	fmt.Printf("%s received status msg response %v from => server at address %v\n", s.ID, msg, from)

	// compare the msg data with the node's own bc; its height and version say
	// cal the differences adn request the missing bits from the peer
	if msg.CurrentHeight <= s.chain.Height() {
		s.Logger.Log("msg", "peer is behind", "peer height", msg.CurrentHeight, "our height", s.chain.Height(), "peer address", from)
	}

	//now we are certain that we are behind peer; so we need to fetch the missing blocks
	go s.reqBlockLoop(from)
	return nil
}

// loop to keep requesting a peer for blocks until we are in sync
// TODO: find a conditoin to detect that we indeed are synced and terminate the loop
// cuz this way we will be congesting the network!!!
func (s *Server) reqBlockLoop(peer NetAddr) error {
	ticker := time.NewTicker(3 * time.Second)
	for {
		ourHeight := s.chain.Height()
		s.Logger.Log("msg", "requesting blocks from peer", "peer", peer, "requesting height", ourHeight+1)

		getBlockMsg := &GetBlockMessage{
			From: s.chain.Height() + 1,
			//0 would signal the remote node to send max no of blocks that they have
			To: 0,
		}
		buf := &bytes.Buffer{}

		if err := gob.NewEncoder(buf).Encode(getBlockMsg); err != nil {
			return err
		}

		rpcMsg := NewMessage(MessageTypeGetBlocks, buf.Bytes())

		s.mu.RLock()
		defer s.mu.RUnlock()

		//to whom we have to send our data
		peer, ok := s.peerMap[peer]
		if !ok {
			return fmt.Errorf("peer %s not found", peer.conn.RemoteAddr())
		}
		if err := peer.Send(rpcMsg.Bytes()); err != nil {
			s.Logger.Log("msg", "failed to send get block msg", "err", err)
		}

		<-ticker.C

	}
}

// when some other nodes requests us for our blocks
func (s *Server) processBlockRequestedMsg(from NetAddr, msg *GetBlockMessage) error {
	fmt.Printf("server %v received get block request msg from %v\n", s.ID, from)
	to := msg.To
	if msg.To == 0 {
		to = s.chain.Height()
	}
	fmt.Printf("blocks requested are from %v to %v\n", msg.From, to)

	//gonna return block headers
	blocks := []*core.Block{}

	//if the remote node is asking for all the blocks
	if msg.To == 0 {
		for i := int(msg.From); i <= int(s.chain.Height()); i++ {
			block, err := s.chain.GetBlock(uint32(i))
			if err != nil {
				return err
			}

			blocks = append(blocks, block)
		}
	}

	//if the remote node is asking for a specific no of blocks then
	for i := msg.From; i <= s.chain.Height(); i++ {
		block, err := s.chain.GetBlock(uint32(i))
		if err != nil {
			return err
		}

		blocks = append(blocks, block)
	}

	fmt.Printf("sending %v blocks to %v\n", (blocks), from)

	s.mu.RLock()
	defer s.mu.RUnlock()

	gob.Register(&BlocksMessage{})

	blocksMsg := &BlocksMessage{
		Blocks: blocks,
	}
	buf := &bytes.Buffer{}

	if err := gob.NewEncoder(buf).Encode(blocksMsg); err != nil {
		return err
	}

	rpcMsg := NewMessage(MessageTypeBlocks, buf.Bytes())

	peer, ok := s.peerMap[from]
	if !ok {
		return fmt.Errorf("peer %s not found", peer.conn.RemoteAddr())
	}

	return peer.Send(rpcMsg.Bytes())
}

// func to process the blocks received from the remote nodes
func (s *Server) processBlockReceipt(from NetAddr, msg *BlocksMessage) error {
	if s.ID == "LATE" {
		fmt.Printf("server %v received blocks from %v\n", s.ID, from)
	}
	for _, block := range msg.Blocks {
		fmt.Println("processing block")
		if err := s.processBlock(block, from); err != nil {
			return err
		}
	}
	return nil
}

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
	s.mu.RLock()
	defer s.mu.RUnlock()
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

	block := core.NewBlock(headers, nil)
	privK := crypto_lib.GeneratePrivateKey()
	block.Sign(privK)

	return block
}
