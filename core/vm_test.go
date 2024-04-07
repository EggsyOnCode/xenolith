package core

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVMAdd(t *testing.T) {
	// 1 + 2 = 3
	//push 1
	//push 2
	//add
	//push 3

	vm := NewVM([]byte{0x01, 0x0a, 0x02, 0x0a, 0x0b})

	log.Printf("VM processing data : %v", vm.data)
	assert.Nil(t, vm.Run())
	fmt.Printf("Stack: %v\n", vm.stack)
	assert.Equal(t, 3, vm.stack.Pop())
}

func TestVMByte(t *testing.T) {

}
