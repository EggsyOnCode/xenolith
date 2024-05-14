package consensus

import (
	"os"
	"sync"
	"time"
	"github.com/EggsyOnCode/xenolith/core"
	"github.com/EggsyOnCode/xenolith/network"
	"github.com/go-kit/log"
)

// Both the exec client and the consensus client use the same  underlying TCP
// Transporter; so both share the same RPC msgs and peers Ch ;
// we can segregate them by using an identifier so that the msgs FOR Consensus
// are not received by the Exec and vice versa

// committee is a random collection of validators; created at the start of every epoch
type Committee struct {
	Validators []*network.Server
}

type ConsensusClientOpts struct {
	BlockTime    time.Duration
	EpochTime    time.Duration
	Logger       log.Logger
	RPCProcessor network.RPCProcessor
	// TCPTransport *network.TCPTransport
	ExecutionClient *network.Server
}

type ConsensusClient struct {
	ConsensusClientOpts
	mu         *sync.RWMutex
	validators []*network.Server
	rpcCh      chan network.RPC
	Committees []*Committee
	ID         string
	quitCh     chan struct{}
}

func NewConsensusClient(opts ConsensusClientOpts) *ConsensusClient {
	if opts.Logger == nil {
		opts.Logger = log.NewLogfmtLogger(os.Stderr)
		// opts.Logger = log.With(opts.Logger, "address", opts.ID)
	}
	cs := &ConsensusClient{
		mu:                  &sync.RWMutex{},
		rpcCh:               make(chan network.RPC),
		validators:          make([]*network.Server, 0),
		Committees:          make([]*Committee, 0),
		ConsensusClientOpts: opts,
		ID:                  opts.ExecutionClient.ID + " consensus",
	}
	if opts.RPCProcessor == nil {
		cs.RPCProcessor = cs
	}

	return cs
}

// consensus server start
func (cs *ConsensusClient) Start() error {

	cs.ExecutionClient.Logger.Log("msg", "consensus client started! Execution and Consensus share the same transporter")

free:
	for {
		select {
		case rpc := <-cs.ExecutionClient.RpcCh:

			msg, err := network.DefaultRPCDecodeFunc(rpc)
			if err != nil {
				cs.ExecutionClient.Logger.Log("err", err)
				continue
			}

			switch msg.Data.(type) {
			// msg of type ValidatorNotification is to be handled inside the consensus layer
			case network.ValidatorNotification:
				if err := cs.RPCProcessor.ProcessMessage(msg); err != nil {
					if err != core.ErrBlockKnown {
						cs.ExecutionClient.Logger.Log("err", err)
					}
				}
			default:
				continue
			}
		case <-cs.quitCh:
			break free
		}
	}

	cs.ExecutionClient.Logger.Log("msg", "consensus client shutting down!!!!...")
	return nil
}

func (cs *ConsensusClient) ProcessMessage(msg *network.DecodedMsg) error {

	switch t := msg.Data.(type) {
	case *network.ValidatorNotification:
		//where t is essentially the msg.Data
		return cs.processNewValidator(t)
	}

	return nil
}

func (cs *ConsensusClient) processNewValidator(msg *network.ValidatorNotification) error {
	cs.mu.Lock()
	//this is to keep track of the validators in the network
	cs.validators = append(cs.validators, msg.Server)
	cs.mu.Unlock()

	return nil
}
