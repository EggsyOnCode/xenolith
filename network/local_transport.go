package network

import (
	"bytes"
	"fmt"
	"sync"
)

type LocalTransport struct {
	addr NetAddr
	// peers is a map of NetAddr to LocalTransport pointers
	peers     map[NetAddr]*LocalTransport
	consumeCh chan RPC
	lock      sync.RWMutex
}

func NewLocalTransport(addr NetAddr) *LocalTransport {
	return &LocalTransport{
		addr:      addr,
		peers:     make(map[NetAddr]*LocalTransport),
		consumeCh: make(chan RPC, 1024),
		lock:      sync.RWMutex{},
	}
}

func (t *LocalTransport) Consume() <-chan RPC {
	return t.consumeCh
}

func (t *LocalTransport) Addr() NetAddr {
	return t.addr
}

// for now we are providing method for connecting local tranport to another local transport
func (t *LocalTransport) Connect(tr Transport) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	//adding a diff local tranport as peer to our current trasnport
	t.peers[tr.Addr()] = tr.(*LocalTransport)
	return nil
}

func (t *LocalTransport) SendMsg(addr NetAddr, payload []byte) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	// sending msg to peer
	fmt.Println(t.peers)
	peer, ok := t.peers[addr]
	if ok {
		peer.consumeCh <- RPC{From: t.addr, Payload: bytes.NewReader(payload)}
		return nil
	}
	return fmt.Errorf("peer not found")

}

func (t *LocalTransport) Broadcast(payload []byte) error{
	for _, peer := range t.peers{
		if err := t.SendMsg(peer.Addr(), payload); err !=nil{
			return err
		}
	}

	return nil
}