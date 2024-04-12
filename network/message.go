package network

import (
	"github.com/EggsyOnCode/xenolith/core"
)

type GetBlockMessage struct {
	From uint32
	//if 0 return max no of blocks
	To uint32
}

type GetStatusMessage struct{}

type StatusMessage struct {
	Version       uint32
	CurrentHeight uint32
}

func NewStatusMessage(v, c uint32) *StatusMessage {
	return &StatusMessage{
		Version:       v,
		CurrentHeight: c,
	}
}

type BlocksMessage struct {
	Blocks []*core.Block
}
