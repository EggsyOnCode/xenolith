package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/EggsyOnCode/xenolith/core"
	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/EggsyOnCode/xenolith/network"
	"github.com/sirupsen/logrus"
)

func main() {
	localT := network.NewLocalTransport("LOCAL")
	remoteA := network.NewLocalTransport("Remote_A")
	remoteB := network.NewLocalTransport("Remote_B")
	remoteC := network.NewLocalTransport("Remote_C")
	localT.Connect(remoteA)
	remoteA.Connect(remoteB)
	remoteA.Connect(remoteC)
	remoteA.Connect(localT)

	initRemoteServers([]network.Transport{remoteA, remoteB, remoteC})

	go func() {
		for {
			//local transport sending to remote transport
			if err := sendTx(remoteA, localT.Addr()); err != nil {
				logrus.Error(err)
			}
			time.Sleep(2 * time.Second)
		}
	}()

	// go func() {
	// 	time.Sleep(7 * time.Second)
	// 	trLate := network.NewLocalTransport("TR_LATE")
	// 	remoteC.Connect(trLate)
	// 	trLateServer := makeServer(trLate, nil, "TR_LATE_SERVER")

	// 	go trLateServer.Start()
	// }()

	//validator node
	pk := crypto_lib.GeneratePrivateKey()
	localServer := makeServer(localT, pk, "LOCAL_SERVER")
	localServer.Start()
}

func initRemoteServers(tr []network.Transport) {
	for i := 0; i < len(tr); i++ {
		id := fmt.Sprintf("REMOTE_%d", i)
		server := makeServer(tr[i], nil, id)
		go server.Start()
	}
}

func makeServer(transport network.Transport, pk *crypto_lib.PrivateKey, id string) *network.Server {
	serverOpts := network.ServerOpts{
		ID:           id,
		Transporters: []network.Transport{transport},
		BlockTime:    5 * time.Second,
		PrivateKey:   pk,
	}
	server, err := network.NewServer(serverOpts)
	if err != nil {
		log.Fatal(err)
	}
	return server
}

func sendTx(localT network.Transport, to network.NetAddr) error {
	pk := crypto_lib.GeneratePrivateKey()
	data := []byte{0x03, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x0d, 0x05, 0x0a, 0x0f}
	tx := core.NewTransaction(data)
	tx.Sign(pk)
	tx.SetTimeStamp(time.Now().Unix())

	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}
	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())
	return localT.SendMsg(to, msg.Bytes())
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// RandString generates a random string of length n.
func RandString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
