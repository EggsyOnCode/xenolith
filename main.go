package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/EggsyOnCode/xenolith/core"
	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/EggsyOnCode/xenolith/network"
)

var BootStrapNodes = []string{":4000", ":5000"}

func main() {
	// tr := network.NewTCPTransporter(":3000")

	// go tr.Start()
	remoteNode0 := makeServer(nil, "Remote_0", ":4000", nil, "")
	go remoteNode0.Start()

	remoteNode1 := makeServer(nil, "Remote_1", ":5000", nil, "")
	go remoteNode1.Start()

	pk := crypto_lib.GeneratePrivateKey()

	localNode := makeServer(pk, "LOCAL", ":3000", BootStrapNodes, ":9999")
	go localNode.Start()

	go func() {
		time.Sleep(11 * time.Second)

		lateNode := makeServer(nil, "LATE", ":6000", []string{":4000"}, "")
		go lateNode.Start()
	}()

	time.Sleep(2 * time.Second)

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			go TCPTester()
			<-ticker.C
		}
	}()
	select {}
}

func makeServer(pk *crypto_lib.PrivateKey, id string, listenAddr string, seedNodes []string, apiListenAddr string) *network.Server {
	serverOpts := network.ServerOpts{
		ListenAddr:     listenAddr,
		ID:             id,
		PrivateKey:     pk,
		BootStrapNodes: seedNodes,
		APIListenAddr:  apiListenAddr,
	}
	localNode, err := network.NewServer(serverOpts)
	if err != nil {
		log.Fatal(err)
	}
	return localNode
}
func TCPTester() {

	pk := crypto_lib.GeneratePrivateKey()
	// data := []byte{0x03, 0x0a, 0x04, 0x0a, 0x0b, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}
	tx := core.NewTransaction(contract())
	fmt.Printf("====> tx hash %x\n", tx.Hash(core.TxHasher{}))
	tx.Sign(pk)
	tx.SetTimeStamp(time.Now().Unix())

	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		fmt.Println(err)
	}
	// msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())

	// _, err = conn.Write(msg.Bytes())
	// if err != nil {
	// 	panic(err)
	// }

	//making http req to our json rpc server ; sending tx over the wire

	err := sendViaHTTP(buf.Bytes())
	if err != nil {
		fmt.Println(err)
	}
}

func sendViaHTTP(b []byte) error {
	req, err := http.NewRequest("POST", "http://localhost:9999/tx", bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	clent := http.Client{}
	_, err1 := clent.Do(req)
	if err1 != nil {
		return err1
	}

	return nil
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
// 		localNode := makeServer(tr[i], nil, id)
// 		go localNode.Start()
// 	}
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

func contract() []byte {
	dataKey := []byte{0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d}
	data := []byte{0x03, 0x0a, 0x04, 0x0a, 0x0b, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}

	data = append(data, dataKey...)
	return data
}

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
