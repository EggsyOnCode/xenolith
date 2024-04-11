package network

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

type TCPTransport struct {
	listenAddr NetAddr
	listener   net.Listener
	peerCh     chan *TCPPeer
}

// /TCP PEER
type TCPPeer struct {
	conn     net.Conn
	Outgoing bool
}

func NewTCPPeer(conn net.Conn, outgoing bool) *TCPPeer {
	return &TCPPeer{conn: conn, Outgoing: outgoing}
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.conn.Write(b)
	return err
}

func (p *TCPPeer) readLoop(rpcCh chan RPC) {
	buf := make([]byte, 2048)
	for {
		n, err := p.conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Connection closed with %v\n", p.conn.RemoteAddr())
			} else {
				fmt.Println("Error reading:", err)
			}
			return
		}
		msg := buf[:n]
		rpcCh <- RPC{
			From:    NetAddr(p.conn.RemoteAddr().String()),
			Payload: bytes.NewReader(msg),
		}
	}
}

// //
func NewTCPTransporter(addr string, peerCh chan *TCPPeer) *TCPTransport {
	return &TCPTransport{
		listenAddr: NetAddr(addr),
		peerCh:     peerCh,
	}
}

func (t *TCPTransport) Start() error {
	ln, err := net.Listen("tcp", string(t.listenAddr))
	if err != nil {
		return err
	}

	t.listener = ln

	go t.acceptLoop()

	return nil
}

func (t *TCPTransport) acceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("error accepting connection: %v\n", err)
			continue
		}
		fmt.Printf("accepted connection from %v\n", conn.RemoteAddr())
		peer := NewTCPPeer(conn, false)
		t.peerCh <- peer

	}
}
