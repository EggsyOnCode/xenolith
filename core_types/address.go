package core_types

import "encoding/hex"

// address is a 20 bytes hexademical string
// obtained from doing a keccak round on the public key and extracting the last 20 bytes
type Address [20]uint8

func (a *Address) ToSlice() []byte {
	buf := make([]byte, 20)
	for i := 0; i < 20; i++ {
		buf[i] = a[i]
	}
	return buf
}

// AddressFromBytes
func AddressFromBytes(b []byte) Address {
	if len(b) != 20 {
		panic("Binary Address must be 20 bytes long")
	}
	var v [20]uint8
	for i := 0; i < 20; i++ {
		v[i] = b[i]
	}

	return Address(v)
}

// hash is implementing the String interface meaning all its outputs will now be typecasted to  hex string
func (a Address) String() string {
	return hex.EncodeToString(a.ToSlice())
}
