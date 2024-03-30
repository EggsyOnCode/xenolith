package core

import "io"

type Encoder[T any] interface {
	Encode(io.Writer, T) error
}
type Decoder[T any] interface {
	Decode(io.Reader, T) error
}

// Concrete Implementation/ Strategies for Encoder and Decoder
type BlockEncoder struct{}
type BlockDecoder struct{}
