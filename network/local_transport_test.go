package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalTransport(t *testing.T) {
	lt := NewLocalTransport(":3000")
	remote := NewLocalTransport(":4000")

	lt.Connect(remote)
	remote.Connect(lt)
	assert.Equal(t, lt.peers[":4000"], remote)
	assert.Equal(t, remote.peers[":3000"], lt)
}

func TestMsgSend(t *testing.T) {
	lt := NewLocalTransport(":3000")
	remote := NewLocalTransport(":4000")

	lt.Connect(remote)
	remote.Connect(lt)

	msg := []byte("Hello World")
	lt.SendMsg(remote.Addr(), msg)

	rpc := <- remote.Consume()
	buf := make([]byte, len(msg))

	n, err := rpc.Payload.Read(buf)
	assert.Nil(t, err)

	assert.Equal(t, n, len(msg))
	assert.Equal(t, buf, msg)
	assert.Equal(t, rpc.From, lt.Addr())
}
