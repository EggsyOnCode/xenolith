package core

type Instruction byte

const (
	InstrPushInt Instruction = 0x0a //10 (0-9 are reserved for ints)
	InstrAdd     Instruction = 0x0b //11
)

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
	s.data[s.sp] = v
	s.sp++
}

func (s *Stack) Pop() any {
	value := s.data[0]
	s.data = append(s.data[:0], s.data[1:]...)
	s.sp--
	return value
}

type VM struct {
	data  []byte
	ip    int //instruction pointer
	stack *Stack
}

func NewVM(data []byte) *VM {
	return &VM{
		data:  data,
		ip:    0,
		stack: NewStack(100),
	}
}

// Run func that reads instructions from the stack register and processes them
func (vm *VM) Run() error {
	for {
		instr := vm.data[vm.ip]
		if err := vm.Exec(Instruction(instr)); err != nil {
			return err
		}
		vm.ip++
		if vm.ip >= len(vm.data) {
			break
		}
	}
	return nil
}

func (vm *VM) Exec(instr Instruction) error {
	switch instr {
	case (InstrPushInt):
		//we have to explicity convert byte to int
		vm.stack.Push(int(vm.data[vm.ip+1]))
	case (InstrAdd):
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := a + b
		vm.stack.Push(c)
	}

	return nil
}
