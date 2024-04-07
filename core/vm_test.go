package core

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVMAdd(t *testing.T) {
	// 1 + 2 = 3
	//1 push
	//2 push
	//add 

	vm := NewVM([]byte{0x01, 0x0a, 0x02, 0x0a, 0x0b})

	log.Printf("VM processing data : %v", vm.data)
	assert.Nil(t, vm.Run())
	fmt.Printf("Stack: %v\n", vm.stack)
	assert.Equal(t, 3, vm.stack.Pop())
}

func TestVMByte(t *testing.T) {
	data := []byte{0x03, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x0d}
	vm := NewVM(data)

	assert.Nil(t, vm.Run())
	result := vm.stack.Pop().([]byte)
	assert.Equal(t, "FOO", string(result))
}

func TestVMSub(t *testing.T) {
	//2 -1 = 3
	// 2 push
	// 1 push
	// sub

	vm := NewVM([]byte{0x02, 0x0a, 0x01, 0x0a, 0x0e})

	log.Printf("VM processing data : %v", vm.data)
	assert.Nil(t, vm.Run())
	fmt.Printf("Stack: %v\n", vm.stack)
	assert.Equal(t, 1, vm.stack.Pop())
}