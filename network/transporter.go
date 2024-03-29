package network

type NetAddr string

type RPC struct {
	From    NetAddr
	Payload []byte
}

type Transport interface {
	Connect(Transport) error
	SendMsg(NetAddr, []byte) error
	Consume() <-chan RPC
	Addr() NetAddr
}
