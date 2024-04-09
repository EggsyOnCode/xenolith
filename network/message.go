package network

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
