package core

import (
	"bytes"
	"math/rand"
	"testing"
	"time"

	"github.com/EggsyOnCode/xenolith/core_types"
	"github.com/stretchr/testify/assert"
)

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func TestCodec(t *testing.T) {
	str := generateRandomString(32)
	header := &Header{
		Version:   1,
		PrevBlock: core_types.HashFromBytes([]byte(str)),
		Head:      1,
		Nonce:     999,
		Timestamp: uint64(time.Now().UnixNano()),
	}
	buf := new(bytes.Buffer)
	assert.Nil(t, header.EncodeBinary(buf))
	//dummy header where the buf will be decoded into
	hDecode := &Header{}
	assert.Nil(t, hDecode.DecodeBinary(buf))
	assert.Equal(t, header, hDecode)
}
