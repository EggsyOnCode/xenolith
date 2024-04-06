package core

type Instruction byte

const (
	InstrPush Instruction = 0x0a //10 (0-9 are reserved for ints)
	InstrAdd  Instruction = 0x0b //11
)

type VM struct {
	data []byte
	ip   int //instruction pointer
	//we could develop our own pkg for struct
	stack []byte
	sp    int //stack pointer
}

func NewVM(data []byte) *VM {
	return &VM{
		data:  data,
		ip:    0,
		stack: make([]byte, 1024),
		sp:    -1,
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
	case (InstrPush):
		vm.pushStack(vm.data[vm.ip+1])
	case (InstrAdd):
		vm.AddOp()
	}

	return nil
}
func (vm *VM) pushStack(v byte) {
	vm.sp++
	vm.stack[vm.sp] = v
}

func (vm *VM) AddOp() {
	a := vm.stack[vm.sp-1]
	b := vm.stack[vm.sp]
	c := a + b
	vm.sp++
	vm.stack[vm.sp] = c
}
