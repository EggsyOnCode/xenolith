package core

import (
	"encoding/binary"
	"fmt"
)

type Instruction byte

const (
	InstrPushInt  Instruction = 0x0a // 10
	InstrAdd      Instruction = 0x0b // 11
	InstrPushByte Instruction = 0x0c // 12
	InstrPack     Instruction = 0x0d // 13
	InstrSub      Instruction = 0x0e // 14
	InstrStore    Instruction = 0x0f // 15
)

// the structure is FIFO not LIFO
type Stack struct {
	data []any
	sp   int
}

func NewStack(size int) *Stack {
	return &Stack{
		data: make([]any, size),
		sp:   0,
	}
}

// strings will be stored in reverse order
func (s *Stack) Push(v any) {
	// inserting item to the top of the stack
	// stack will grow from r->l
	s.data = append([]any{v}, s.data...)
	s.sp++
}

func (s *Stack) Pop() any {
	value := s.data[0]
	s.data = append(s.data[:0], s.data[1:]...)
	s.sp--
	fmt.Printf("popped value is %v\n", value)

	return value
}

type VM struct {
	data          []byte
	ip            int // instruction pointer
	stack         *Stack
	contractState *State
}

func NewVM(data []byte, contractState *State) *VM {
	return &VM{
		data:          data,
		contractState: contractState,
		ip:            0,
		stack:         NewStack(128),
	}
}

func (vm *VM) Run() error {
	for {
		instr := Instruction(vm.data[vm.ip])

		if err := vm.Exec(instr); err != nil {
			return err
		}

		vm.ip++

		if vm.ip > len(vm.data)-1 {
			break
		}
	}

	return nil
}

func (vm *VM) Exec(instr Instruction) error {
	switch instr {
	case InstrPushInt:
		vm.stack.Push(int(vm.data[vm.ip-1]))

	case InstrPushByte:
		vm.stack.Push(byte(vm.data[vm.ip-1]))

	case InstrPack:
		n := vm.stack.Pop().(int)
		b := make([]byte, n)

		for i := 0; i < n; i++ {
			b[i] = vm.stack.Pop().(byte)
		}

		reverseByteOrder := ReverseByteOrder(b)

		vm.stack.Push(reverseByteOrder)

	case InstrAdd:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := a + b
		vm.stack.Push(c)

	case InstrSub:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := a - b
		vm.stack.Push(c)
	case InstrStore:
		key := vm.stack.Pop().([]byte)
		value := vm.stack.Pop()
		var serializedVal []byte

		switch t := value.(type) {
		case int:
			serializedVal = serializeInt64(int64(t))
		default:
			panic("TODO: implement serialization for other types")
		}

		// fmt.Printf("%v\n", key)
		// fmt.Printf("%v\n", value)

		vm.contractState.Put(key, serializedVal)
	}

	return nil
}

func serializeInt64(val int64) []byte {
	// 8 byte long byte slice
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(val))

	return buf
}

func DeserializeInt64(b []byte) int64 {
	val := binary.LittleEndian.Uint64(b)
	return int64(val)
}

func ReverseByteOrder(b []byte) []byte {
	for i := 0; i < len(b)/2; i++ {
		b[i], b[len(b)-i-1] = b[len(b)-i-1], b[i]
	}
	return b
}
