package emulator

var Fonts [80]uint8 = [80]uint8{
	0xf0, 0x90, 0x90, 0x90, 0xf0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xf0, 0x10, 0xf0, 0x80, 0xf0, // 2
	0xf0, 0x10, 0xf0, 0x10, 0xf0, // 3
	0x90, 0x90, 0xf0, 0x10, 0x10, // 4
	0xf0, 0x80, 0xf0, 0x10, 0xf0, // 5
	0xf0, 0x80, 0xf0, 0x90, 0xf0, // 6
	0xf0, 0x10, 0x20, 0x40, 0x40, // 7
	0xf0, 0x90, 0xf0, 0x90, 0xf0, // 8
	0xf0, 0x90, 0xf0, 0x10, 0xf0, // 9
	0xf0, 0x90, 0xf0, 0x90, 0x90, // A
	0xe0, 0x90, 0xe0, 0x90, 0xe0, // B
	0xf0, 0x80, 0x80, 0x80, 0xf0, // C
	0xe0, 0x90, 0x90, 0x90, 0xe0, // D
	0xf0, 0x80, 0xf0, 0x80, 0xf0, // E
	0xf0, 0x80, 0xf0, 0x80, 0x80, // F
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

func (e *Emulator) Run() {
	for e.Step() {
		// Run the next instruction.
	}
}

func (e *Emulator) Step() bool {
	hi := uint16(e.memory[e.pc])
	lo := uint16(e.memory[e.pc+1])
	op := (hi << 8) | lo

	switch op & 0xf000 {
	case 0x0000:
		kind := (op & 0x00ff)

		switch kind {
		case 0x00e0:
			panic("not implemented")
		case 0x00ee:
			panic("not implemented")
		default:
			// The opcode 0NNN jumps to a machine code routine at address NNN, but it
			// is only used on the computers on which CHIP-8 was implemented. This
			// interpreter implements an opcode of this form as a HALT instruction.
			return false
		}
	case 0x3000:
		x := (op & 0x0f00) >> 8
		n := uint8(op & 0x00ff)

		if e.v[x] == n {
			e.pc += 4
		} else {
			e.pc += 2
		}
	case 0x4000:
		x := (op & 0x0f00) >> 8
		n := uint8(op & 0x00ff)

		if e.v[x] != n {
			e.pc += 4
		} else {
			e.pc += 2
		}
	case 0x5000:
		x := (op & 0x0f00) >> 8
		y := (op & 0x00f0) >> 4

		if e.v[x] == e.v[y] {
			e.pc += 4
		} else {
			e.pc += 2
		}
	case 0x6000:
		x := (op & 0x0f00) >> 8
		v := uint8(op & 0x00ff)
		e.v[x] = v
		e.pc += 2
	case 0x7000:
		x := (op & 0x0f00) >> 8
		v := uint8(op & 0x00ff)
		e.v[x] += v
		e.pc += 2
	case 0x8000:
		x := (op & 0x0f00) >> 8
		y := (op & 0x00f0) >> 4

		kind := op & 0x000f

		switch kind {
		case 0x0000:
			e.v[x] = e.v[y]
		case 0x0001:
			e.v[x] |= e.v[y]
		case 0x0002:
			e.v[x] &= e.v[y]
		case 0x0003:
			e.v[x] ^= e.v[y]
		case 0x0004:
			if e.v[x] > 0xff-e.v[y] {
				e.v[0xf] = 1
			} else {
				e.v[0xf] = 0
			}
			e.v[x] += e.v[y]
		case 0x0005:
			if e.v[x] >= e.v[y] {
				e.v[0xf] = 1
			} else {
				e.v[0xf] = 0
			}
			e.v[x] -= e.v[y]
		case 0x0006:
			e.v[0xf] = e.v[x] & 0x01
			e.v[x] >>= 1
		case 0x0007:
			if e.v[y] >= e.v[x] {
				e.v[0xf] = 1
			} else {
				e.v[0xf] = 0
			}
			e.v[x] = e.v[y] - e.v[x]
		case 0x000e:
			e.v[0xf] = (e.v[x] & 0x80) >> 7
			e.v[x] <<= 1
		}

		e.pc += 2
	case 0x9000:
		x := (op & 0x0f00) >> 8
		y := (op & 0x00f0) >> 4

		if e.v[x] != e.v[y] {
			e.pc += 4
		} else {
			e.pc += 2
		}
	}

	return true
}
