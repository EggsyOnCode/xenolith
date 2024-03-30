package crypto

import (
	"log"
	"testing"
)

func TestKeyPairGen(t *testing.T) {
	priv := GeneratePrivateKey()
	pb := priv.PublicKey()
	address := pb.Address()

	log.Println(address)
}
