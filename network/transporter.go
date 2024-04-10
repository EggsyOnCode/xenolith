package network

type Transport interface {
	Connect(Transport) error
	SendMsg(NetAddr, []byte) error
	//the extra parameter is to exclude a peer from broadcasting
	Broadcast([]byte, NetAddr) error
	Consume() <-chan RPC
	Addr() NetAddr
	Peers() map[NetAddr]Transport
	AddPeer(tr Transport) 
}
