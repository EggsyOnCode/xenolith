package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalTransport(t *testing.T) {
	lt := NewLocalTransport(":3000")
	remote := NewLocalTransport(":4000")

	lt.Connect(remote)
	assert.Equal(t, lt.peers[":4000"], remote)
	lt.SendMsg(":4000", []byte("hello"))
	assert.Equal(t, <-remote.consumeCh, RPC{From: ":3000", Payload: []byte("hello")})
}
