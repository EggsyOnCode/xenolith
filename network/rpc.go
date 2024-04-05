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

func NewRPCMsg(from NetAddr, payload []byte) *RPC {
	return &RPC{From: from, Payload: bytes.NewReader(payload)}
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

func NewMessage(t MessageType, data []byte) *Message {
	return &Message{
		Headers: t,
		Data:    data,
	}
}

func (m *Message) Bytes() []byte {
	buf := new(bytes.Buffer)
	gob.NewEncoder(buf).Encode(m)
	return buf.Bytes()
}

type RPCHandler interface {
	HandleRPC(RPC) error
}

type RPCProcessor interface {
	ProcessTx(NetAddr, *core.Transaction) error
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
		return fmt.Errorf("failed to decode message: %v ; from : %v", err, r.From)
	}

	switch msg.Headers {
	case MessageTypeTx:
		tx := new(core.Transaction)
		//msg.Data --> io reader --> feeding into the gob decoder --> decode into tx
		if err := tx.Decode(core.NewGobTxDecoder(bytes.NewReader(msg.Data))); err != nil {
			return err
		}

		return rpc.p.ProcessTx(r.From, tx)
	default:
		fmt.Println(msg)
		return fmt.Errorf("unknown message type: %v", msg.Headers)
	}
}

func init() {
	gob.Register(Message{})
}
