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

	contractState := NewState()
	vm := NewVM([]byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}, contractState)

	log.Printf("VM processing data : %v", vm.data)
	assert.Nil(t, vm.Run())
	fmt.Printf("Stack: %v\n", vm.stack)
	assert.Equal(t, 3, vm.stack.Pop())
}

func TestVMMul(t *testing.T) {
	contractState := NewState()
	vm := NewVM([]byte{0x02, 0x0a, 0x04, 0x0a, 0x1b}, contractState)

	log.Printf("VM processing data : %v", vm.data)
	assert.Nil(t, vm.Run())
	fmt.Printf("Stack: %v\n", vm.stack)
	assert.Equal(t, 8, vm.stack.Pop().(int))
}
func TestVMDiv(t *testing.T) {
	contractState := NewState()
	vm := NewVM([]byte{0x08, 0x0a, 0x02, 0x0a, 0x1c}, contractState)

	log.Printf("VM processing data : %v", vm.data)
	assert.Nil(t, vm.Run())
	fmt.Printf("Stack: %v\n", vm.stack)
	assert.Equal(t, 4, vm.stack.Pop())
}
func TestVMByte(t *testing.T) {
	data := []byte{0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d}
	contractState := NewState()
	vm := NewVM(data, contractState)

	assert.Nil(t, vm.Run())
	result := vm.stack.Pop().([]byte)
	assert.Equal(t, "FOO", string(result))
}

func TestVMStrings(t *testing.T) {
	str := []byte("hello world is ok!")
	lenStr := len(str)
	data := []byte{}
	for i := 0; i < lenStr; i++ {
		data = append(data, str[i])
		data = append(data, 0x0c)
	}
	data = append(data, byte(lenStr), 0x0a)
	data = append(data, 0x0d)
	contractState := NewState()
	vm := NewVM(data, contractState)

	assert.Nil(t, vm.Run())
	fmt.Printf("result: %v\n", vm.stack.data)
	result := vm.stack.Pop().([]byte)
	assert.Equal(t, "hello world is ok!", string(result))
}

func TestVMState(t *testing.T) {
	// FOO : key
	// serializedInt as value
	data := []byte{0x03, 0x0a, 0x04, 0x0a, 0x0b, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}
	contractState := NewState()
	vm := NewVM(data, contractState)

	assert.Nil(t, vm.Run())
	conState, err := contractState.Get([]byte("FOO"))
	assert.Nil(t, err)
	conVal := DeserializeInt64(conState)
	fmt.Printf("vm stack data %+v\n", vm.stack.data)
	fmt.Printf("contract state value %+v\n", contractState)

	assert.Equal(t, int64(7), conVal)
}

func TestGetVM(t *testing.T) {
	dataKey := []byte{0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d}
	// dataVal := []byte{0x03, 0x0a, 0x04, 0x0a, 0x0b}
	data := []byte{0x03, 0x0a, 0x04, 0x0a, 0x0b, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}
	contractState := NewState()
	vm := NewVM(data, contractState)
	vm.data = append(vm.data, dataKey...)
	vm.data = append(vm.data, 0x1a)

	assert.Nil(t, vm.Run())
	// conState, err := contractState.Get([]byte("FOO"))
	// assert.Nil(t, err)
	// conVal := DeserializeInt64(conState)
	// fmt.Printf("vm stack data %+v\n", vm.stack.data)
	// fmt.Printf("contract state value %+v\n", contractState)
	// assert.Equal(t, int64(7), conVal)

	fmt.Printf("vm stack data again %+v\n", vm.stack.data)
	conVal := DeserializeInt64(vm.stack.Pop().([]byte))
	assert.Equal(t, int64(7), conVal)

}
