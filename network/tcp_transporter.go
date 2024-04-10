package network

import (
	"fmt"
	"net"
)

type TCPTransporter struct {
	listenAddr NetAddr
	listener   net.Listener
}

type TCPPeer struct {
	conn net.Conn
}

func NewTCPPeer(conn net.Conn) *TCPPeer {
	return &TCPPeer{conn: conn}
}

func NewTCPTransporter(addr string) *TCPTransporter {
	return &TCPTransporter{
		listenAddr: NetAddr(addr),
	}
}

func (t *TCPTransporter) acceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("error accepting connection: %v\n", err)
			continue
		}
		fmt.Printf("accepted connection from %v\n", conn.RemoteAddr())

		go t.readLoop(NewTCPPeer(conn))

	}
}

func (t *TCPTransporter) readLoop(peer *TCPPeer) {
	buf := make([]byte, 2048)
	for {
		n, err := peer.conn.Read(buf)
		if err != nil {
			fmt.Printf("error reading from connection: %v\n", err)
			return
		}
		msg := buf[:n]
		fmt.Printf("msg: %v\n", string(msg))
	}
}

func (t *TCPTransporter) Start() error {
	ln, err := net.Listen("tcp", string(t.listenAddr))
	if err != nil {
		return err
	}

	t.listener = ln

	go t.acceptLoop()

	fmt.Printf("TCP listening on port %v\n", t.listenAddr)

	return nil
}
