package network

type Transport interface {
	Connect(Transport) error
	SendMsg(NetAddr, []byte) error
	Broadcast([]byte) error
	Consume() <-chan RPC
	Addr() NetAddr
}
