package network

import (
	"io/ioutil"
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

	rpc := <-remote.Consume()
	buf := make([]byte, len(msg))

	n, err := rpc.Payload.Read(buf)
	assert.Nil(t, err)

	assert.Equal(t, n, len(msg))
	assert.Equal(t, buf, msg)
	assert.Equal(t, rpc.From, lt.Addr())
}

func TestBroadcastMsg(t *testing.T) {
	lt := NewLocalTransport(":3000")
	remote := NewLocalTransport(":4000")
	remote2 := NewLocalTransport(":5000")
	lt.Connect(remote)
	lt.Connect(remote2)

	msg := []byte("Hello World")

	lt.Broadcast(msg)

	rpc := <-remote.Consume()
	b, err := ioutil.ReadAll(rpc.Payload)
	assert.Nil(t, err)
	assert.Equal(t, b, msg)


	rpc2 := <-remote2.Consume()
	b2, err2 := ioutil.ReadAll(rpc2.Payload)
	assert.Nil(t, err2)
	assert.Equal(t, b2, msg)
}
