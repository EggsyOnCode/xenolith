package core

import "io"

type Transaction struct {}

func (t *Transaction) EncodeBinary(w io.Writer) error{ return nil}
func (t *Transaction) DecodeBinary(r io.Reader) error{ return nil}