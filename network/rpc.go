package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"

	"github.com/EggsyOnCode/xenolith/core"
)

type NetAddr string

type RPC struct {
	From    NetAddr
	Payload io.Reader
}

type MessageType byte

const (
	MessageTypeTx    MessageType = 0x1
	MessageTypeBlock MessageType = 0x2
)

type Message struct {
	Headers MessageType
	Data    []byte
}

type RPCHandler interface {
	HandleRPC(RPC) error
}

type RPCProcessor interface {
	HandleTx(NetAddr, *core.Transaction) error
}

// concrete RPCHandler implementation
type DefaultRPCHandler struct {
	p RPCProcessor
}

func NewRPCHandler(processor RPCProcessor) *DefaultRPCHandler {
	return &DefaultRPCHandler{
		p: processor,
	}
}

func (rpc *DefaultRPCHandler) HandleRPC(r RPC) error {
	//decoding the payload of hte rpc via gob decoder
	msg := &Message{}
	if err := gob.NewDecoder(r.Payload).Decode(msg); err != nil {
		return err
	}

	switch msg.Headers {
	case MessageTypeTx:
		tx := new(core.Transaction)
		//msg.Data --> io reader --> feeding into the gob decoder --> decode into tx
		if err := tx.Decode(core.NewGobTxDecoder(bytes.NewReader(msg.Data))); err != nil {
			return err
		}

		rpc.p.HandleTx(r.From, tx)
	default:
		return fmt.Errorf("unknown message type: %v", msg.Headers)
	}
	return nil
}
