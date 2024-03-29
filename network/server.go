package network

import (
	"fmt"
	"time"
)

type ServerOpts struct {
	Transporters []Transport
}

type Server struct {
	ServerOpts

	rpcCh  chan RPC
	quitCh chan struct{}
}

func NewServer(opts ServerOpts) *Server {
	return &Server{
		ServerOpts: opts,
		rpcCh:      make(chan RPC),
		quitCh:     make(chan struct{}),
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
			fmt.Println("Server received msg from ", rpc.From, " with payload ", string(rpc.Payload))
		case <-s.quitCh:
			break free
		case <-ticker.C:
			fmt.Println("server is doing yyy in x seconds")
		}
	}
	fmt.Println("Server Stopped!")
	return nil
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
