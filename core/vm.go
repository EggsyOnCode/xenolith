package core

import "fmt"

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

func (s *Stack) Push(v any) {
	fmt.Printf("stack before push %v\n", s.data)
	s.data[s.sp] = v
	s.sp++
	fmt.Printf("stack after push %v\n", s.data)
}

func (s *Stack) Pop() any {
	fmt.Printf("stack before pop %v\n", s.data)
	value := s.data[0]
	s.data = append(s.data[:0], s.data[1:]...)
	s.sp--

	fmt.Printf("stack after pop %v\n", s.data)
	return value
}

type VM struct {
	data  []byte
	ip    int // instruction pointer
	stack *Stack
}

func NewVM(data []byte) *VM {
	return &VM{
		data:  data,
		ip:    0,
		stack: NewStack(128),
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

		vm.stack.Push(b)

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
	}

	return nil
}
