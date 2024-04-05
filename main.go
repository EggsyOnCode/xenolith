package main

import (
	"time"

	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/EggsyOnCode/xenolith/network"
)

func main() {
	tr := network.NewLocalTransport(":3000")
	rt := network.NewLocalTransport(":4000")
	tr.Connect(rt)
	rt.Connect(tr)

	go func() {
		for {
			rt.SendMsg(":3000", []byte("hello local"))
			time.Sleep(500 * time.Millisecond)
		}
	}()

	//validator node
	pk := crypto_lib.GeneratePrivateKey()
	serverOpts := network.ServerOpts{
		Transporters: []network.Transport{tr},
		BlockTime:    5 * time.Second,
		PrivateKey:   pk,
	}
	server := network.NewServer(serverOpts)
	server.Start()
	select {}
}
