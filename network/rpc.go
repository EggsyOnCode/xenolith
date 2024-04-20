package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"

	"github.com/EggsyOnCode/xenolith/core"
	"github.com/sirupsen/logrus"
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
	MessageTypeTx              MessageType = 0x1
	MessageTypeBlock           MessageType = 0x2
	MessageTypeGetBlocks       MessageType = 0x3
	MessageStatusType          MessageType = 0x4
	MessageGetStatusType       MessageType = 0x5
	MessageTypeBlocks          MessageType = 0x6
	MessageTypeValidatorInform MessageType = 0x7
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
	ProcessMessage(*DecodedMsg) error
}

// concrete RPCHandler implementation
type DecodedMsg struct {
	From NetAddr
	Data any
}

type RPCDecodeFunc func(RPC) (*DecodedMsg, error)

func DefaultRPCDecodeFunc(rpc RPC) (*DecodedMsg, error) {
	//decoding the payload of hte rpc via gob decoder
	msg := &Message{}
	if err := gob.NewDecoder(rpc.Payload).Decode(msg); err != nil {
		return nil, fmt.Errorf("failed to decode message: %v ; from : %v", err, rpc.From)
	}

	// to print out the incoming msg type and the origin
	logrus.WithFields(logrus.Fields{
		"from": rpc.From,
		"type": msg.Headers,
	}).Debug("incoming message")

	switch msg.Headers {
	case MessageTypeTx:
		tx := new(core.Transaction)
		//msg.Data --> io reader --> feeding into the gob decoder --> decode into tx
		if err := tx.Decode(core.NewGobTxDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, err
		}

		return &DecodedMsg{
			From: rpc.From,
			Data: tx,
		}, nil
	case MessageTypeGetBlocks:
		getBlockMsg := new(GetBlockMessage)
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(getBlockMsg); err != nil {
			return nil, err
		}
		return &DecodedMsg{
			From: rpc.From,
			Data: getBlockMsg,
		}, nil
	case MessageTypeBlock:
		block := new(core.Block)
		if err := block.Decode(core.NewGobBlockDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, err
		}

		return &DecodedMsg{
			From: rpc.From,
			Data: block,
		}, nil
	case MessageTypeBlocks:
		blocMsg := &BlocksMessage{}
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(blocMsg); err != nil {
			return nil, err
		}

		return &DecodedMsg{
			From: rpc.From,
			Data: blocMsg,
		}, nil
	case MessageTypeValidatorInform:
		validatorMsg := &ValidatorNotification{}
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(validatorMsg); err != nil {
			return nil, err
		}

		return &DecodedMsg{
			From: rpc.From,
			Data: validatorMsg,
		}, nil
	case MessageGetStatusType:
		return &DecodedMsg{
			From: rpc.From,
			Data: &GetStatusMessage{},
		}, nil
	case MessageStatusType:
		statusMsg := new(StatusMessage)
		// new decoder takes in a reader
		// teh reader is the byte stream of the obj that has the encoded data
		// decode takes in some structure to store the decoded data
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(statusMsg); err != nil {
			return nil, err
		}

		return &DecodedMsg{
			From: rpc.From,
			Data: statusMsg,
		}, nil

	default:
		fmt.Println(msg)
		return nil, fmt.Errorf("unknown message type: %v", msg.Headers)
	}
}

func init() {
	gob.Register(Message{})
}
