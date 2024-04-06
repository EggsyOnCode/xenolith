package core

import (
	"crypto/elliptic"
	"encoding/gob"
	"io"
)

type Encoder[T any] interface {
	Encode(T) error
}
type Decoder[T any] interface {
	Decode(T) error
}

// Concrete Implementation/ Strategies for Encoder and Decoder
type BlockEncoder struct{}
type BlockDecoder struct{}





type GobTxEncoder struct {
	writer io.Writer
}
type GobTxDecoder struct {
	reader io.Reader
}

func NewGobTxEncoder(w io.Writer) *GobTxEncoder {
	gob.Register(elliptic.P256())
	return &GobTxEncoder{
		writer: w,
	}
}

func (g GobTxEncoder) Encode(tx *Transaction) error {
	return gob.NewEncoder(g.writer).Encode(tx)
}

func NewGobTxDecoder(r io.Reader) *GobTxDecoder {
	gob.Register(elliptic.P256())
	return &GobTxDecoder{
		reader: r,
	}
}

// decode encoded tx
func (g GobTxDecoder) Decode(tx *Transaction) error {
	return gob.NewDecoder(g.reader).Decode(tx)
}
