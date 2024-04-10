package main

import (
	"net"
	"time"

	"github.com/EggsyOnCode/xenolith/network"
)

var transports = []network.Transport{
	network.NewLocalTransport("LOCAL"),
	network.NewLocalTransport("Remote_0"),
	network.NewLocalTransport("Remote_1"),
	// network.NewLocalTransport("Remote_2"),
}

func main() {
	tr := network.NewTCPTransporter(":3000")

	go tr.Start()

	time.Sleep(2 * time.Second)
	TCPTester()
	select {}
}

func TCPTester() {
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		panic(err)
	}

	_, err = conn.Write([]byte("Hello, world!"))
	if err != nil {
		panic(err)
	}
}

// func main() {

// 	initRemoteServers(transports)
// 	localNode := transports[0]
// 	remoteNodeA := transports[1]
// 	// remoteNodeC := transports[3]

// 	go func() {
// 		for {
// 			//local transport sending to remote transport
// 			if err := sendTx(remoteNodeA, localNode.Addr()); err != nil {
// 				logrus.Error(err)
// 			}
// 			time.Sleep(2 * time.Second)
// 		}
// 	}()

// 	go func() {
// 		time.Sleep(4 * time.Second)
// 		trLate := network.NewLocalTransport("TR_LATE")
// 		trLate.Connect(remoteNodeA)
// 		remoteNodeA.Connect(trLate)
// 		trLateServer := makeServer(trLate, nil, "TR_LATE")
// 		go trLateServer.Start()

// 	}()

// 	//validator node
// 	pk := crypto_lib.GeneratePrivateKey()
// 	localServer := makeServer(transports[0], pk, "LOCAL")
// 	localServer.Start()
// }

// func initRemoteServers(tr []network.Transport) {
// 	for i := 0; i < len(tr)-1; i++ {
// 		id := fmt.Sprintf("REMOTE_%d", i)
// 		server := makeServer(tr[i], nil, id)
// 		go server.Start()
// 	}
// }

// func makeServer(transport network.Transport, pk *crypto_lib.PrivateKey, id string) *network.Server {
// 	serverOpts := network.ServerOpts{
// 		Transport:    transport,
// 		ID:           id,
// 		Transporters: transports,
// 		BlockTime:    5 * time.Second,
// 		PrivateKey:   pk,
// 	}
// 	server, err := network.NewServer(serverOpts)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return server
// }

// func sendTx(localT network.Transport, to network.NetAddr) error {
// 	pk := crypto_lib.GeneratePrivateKey()
// 	// data := []byte{0x03, 0x0a, 0x04, 0x0a, 0x0b, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}
// 	tx := core.NewTransaction(contract())
// 	tx.Sign(pk)
// 	tx.SetTimeStamp(time.Now().Unix())

// 	buf := &bytes.Buffer{}
// 	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
// 		return err
// 	}
// 	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())
// 	return localT.SendMsg(to, msg.Bytes())
// }

// func contract() []byte {
// 	dataKey := []byte{0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d}
// 	data := []byte{0x03, 0x0a, 0x04, 0x0a, 0x0b, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}

// 	data = append(data, dataKey...)
// 	return data
// }

// const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// // RandString generates a random string of length n.
// func RandString(n int) string {
// 	rand.Seed(time.Now().UnixNano())
// 	b := make([]byte, n)
// 	for i := range b {
// 		b[i] = letterBytes[rand.Intn(len(letterBytes))]
// 	}
// 	return string(b)
// }
