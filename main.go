package main

import (
	"bytes"
	"log"
	"math/rand"
	"time"

	"github.com/EggsyOnCode/xenolith/core"
	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/EggsyOnCode/xenolith/network"
	"github.com/sirupsen/logrus"
)

func main() {
	tr := network.NewLocalTransport(":3000")
	rt := network.NewLocalTransport(":4000")
	tr.Connect(rt)
	rt.Connect(tr)

	go func() {
		for {
			//local transport sending to remote transport
			if err := sendTx(rt, tr); err != nil {
				logrus.Error(err)
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	//validator node
	pk := crypto_lib.GeneratePrivateKey()
	serverOpts := network.ServerOpts{
		ID:           "local",
		Transporters: []network.Transport{tr},
		BlockTime:    5 * time.Second,
		PrivateKey:   pk,
	}
	server, err := network.NewServer(serverOpts)
	if err != nil {
		log.Fatal(err)
	}
	server.Start()
	select {}
}

func sendTx(tr network.Transport, to network.Transport) error {
	pk := crypto_lib.GeneratePrivateKey()
	tx := core.NewTransaction([]byte(RandString(10)))
	tx.Sign(pk)
	tx.SetTimeStamp(time.Now().Unix())

	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}
	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())
	return tr.SendMsg(to.Addr(), msg.Bytes())
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
