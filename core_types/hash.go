package core_types

type Hash [32]uint8

// HashFromBytes converts a byte slice to a Hash
func HashFromBytes(b []byte) Hash {
	if len(b) != 32 {
		panic("Binary Hash must be 32 bytes long")
	}
	var v [32]uint8
	for i := 0; i < 32; i++ {
		v[i] = b[i]
	}

	return Hash(v)
}
