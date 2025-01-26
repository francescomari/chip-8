package emulator

var Fonts [80]uint8 = [80]uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

type Emulator struct {
	memory [4096]uint8 // Main memory (4KB)
	v      [16]uint8   // Register array (V0 to VF)
	i      uint16      // Index register (12-bit)
	stack  [16]uint16  // Stack frames
	sp     uint8       // Pointer to the next available stack frame
	dt     uint8       // Delaty timer
	st     uint8       // Sound timer
	pc     uint16      // Program counter (12-bit)
}

func (e *Emulator) Memory() []uint8 {
	return e.memory[:]
}

func (e *Emulator) V() []uint8 {
	return e.v[:]
}

func (e *Emulator) I() uint16 {
	return e.i
}

func (e *Emulator) Stack() []uint16 {
	return e.stack[:]
}

func (e *Emulator) SP() uint8 {
	return e.sp
}

func (e *Emulator) DT() uint8 {
	return e.dt
}

func (e *Emulator) ST() uint8 {
	return e.st
}

func (e *Emulator) PC() uint16 {
	return e.pc
}

func (e *Emulator) Init() {
	for i := range len(e.memory) {
		e.memory[i] = 0
	}

	copy(e.memory[:], Fonts[:])

	for i := range len(e.v) {
		e.v[i] = 0
	}

	for i := range len(e.stack) {
		e.stack[i] = 0
	}

	e.i = 0
	e.sp = 0
	e.dt = 0
	e.st = 0
	e.pc = 0x200
}

func (e *Emulator) Load(program []uint8) {
	copy(e.memory[0x200:], program)
}

func (e *Emulator) Step() {
	op := uint16(e.memory[e.pc])<<8 | uint16(e.memory[e.pc+1])

	switch op & 0xF000 {
	case 0x6000:
		e.v[(op&0x0F00)>>8] = uint8(op & 0x00FF)
		e.pc += 2
	case 0x7000:
		e.v[(op&0x0F00)>>8] += uint8(op & 0x00FF)
		e.pc += 2
	}
}
