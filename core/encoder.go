package core

import (
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
type GobBlockEncoder struct {
	w io.Writer
}

type GobBlockDecoder struct {
	r io.Reader
}

func NewGobBlockEncoder(w io.Writer) *GobBlockEncoder {
	return &GobBlockEncoder{
		w: w,
	}
}

func NewGobBlockDecoder(r io.Reader) *GobBlockDecoder {
	return &GobBlockDecoder{
		r: r,
	}
}

func (g GobBlockEncoder) Encode(b *Block) error {
	return gob.NewEncoder(g.w).Encode(b)
}

func (g GobBlockDecoder) Decode(b *Block) error {
	return gob.NewDecoder(g.r).Decode(b)
}

type GobTxEncoder struct {
	writer io.Writer
}
type GobTxDecoder struct {
	reader io.Reader
}

func NewGobTxEncoder(w io.Writer) *GobTxEncoder {
	return &GobTxEncoder{
		writer: w,
	}
}

func (g GobTxEncoder) Encode(tx *Transaction) error {
	return gob.NewEncoder(g.writer).Encode(tx)
}

func NewGobTxDecoder(r io.Reader) *GobTxDecoder {
	return &GobTxDecoder{
		reader: r,
	}
}

// decode encoded tx
func (g GobTxDecoder) Decode(tx *Transaction) error {
	return gob.NewDecoder(g.reader).Decode(tx)
}
