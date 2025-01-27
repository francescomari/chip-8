package emulator

import (
	"math/rand/v2"
	"sync"
)

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

type (
	Memory    [4096]uint8
	Registers [16]uint8
	Stack     [16]uint16
	Display   [64][128]uint8
)

type Emulator struct {
	mu          sync.RWMutex
	memory      Memory        // Main memory (4KB)
	v           Registers     // Register array (V0 to VF)
	i           uint16        // Index register (12-bit)
	stack       Stack         // Stack frames
	sp          uint8         // Pointer to the next available stack frame
	dt          uint8         // Delay timer
	st          uint8         // Sound timer
	pc          uint16        // Program counter (12-bit)
	display     Display       // Display
	key         uint8         // Currently pressed key, if any
	keyDown     bool          // Is a key pressed?
	lastKey     uint8         // Last key pressed, if waiting and set
	lastKeyWait bool          // Waiting for a key?
	lastKeySet  bool          // Key set while waiting?
	rng         func() uint32 // Random number generator
}

func (e *Emulator) Memory(buffer []uint8) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	copy(buffer, e.memory[:])
}

func (e *Emulator) V(buffer []uint8) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	copy(buffer, e.v[:])
}

func (e *Emulator) I() uint16 {
	return e.i
}

func (e *Emulator) Stack(buffer []uint16) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	copy(buffer, e.stack[:])
}

func (e *Emulator) SP() uint8 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.sp
}

func (e *Emulator) DT() uint8 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.dt
}

func (e *Emulator) DTClock() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.dt != 0 {
		e.dt--
	}
}

func (e *Emulator) ST() uint8 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.st
}

func (e *Emulator) STClock() bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.st == 0 {
		return false
	}

	e.st--

	return e.st == 0
}

func (e *Emulator) PC() uint16 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.pc
}

func (e *Emulator) Display(buffer *Display) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for y := range buffer {
		for x := range buffer[y] {
			buffer[y][x] = e.display[y][x]
		}
	}
}

func (e *Emulator) KeyDown(key uint8) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.keyDown = true
	e.key = key

	if e.lastKeyWait && !e.lastKeySet {
		e.lastKey = key
		e.lastKeySet = true
	}
}

func (e *Emulator) KeyUp() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.keyDown = false
}

func (e *Emulator) SetRNG(rng func() uint32) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.rng = rng
}

func (e *Emulator) Init() {
	e.mu.Lock()
	defer e.mu.Unlock()

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

	for y := range e.display {
		for x := range e.display[y] {
			e.display[y][x] = 0
		}
	}

	e.i = 0
	e.sp = 0
	e.dt = 0
	e.st = 0
	e.pc = 0x200
}

func (e *Emulator) Load(program []uint8) {
	e.mu.Lock()
	defer e.mu.Unlock()

	copy(e.memory[0x200:], program)
}

func (e *Emulator) Run() {
	for e.Step() {
		// Run the next instruction.
	}
}

func (e *Emulator) Step() bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	hi := uint16(e.memory[e.pc])
	lo := uint16(e.memory[e.pc+1])
	op := (hi << 8) | lo

	switch op & 0xf000 {
	case 0x0000:
		kind := (op & 0x00ff)

		switch kind {
		case 0x00e0:
			for y := range e.display {
				for x := range e.display[y] {
					e.display[y][x] = 0
				}
			}

			e.pc += 2
		case 0x00ee:
			e.sp--
			e.pc = e.stack[e.sp]
			e.pc += 2
		default:
			// The opcode 0NNN jumps to a machine code routine at address NNN, but it
			// is only used on the computers on which CHIP-8 was implemented. This
			// interpreter implements an opcode of this form as a HALT instruction.
			return false
		}
	case 0x1000:
		e.pc = op & 0x0fff
	case 0x2000:
		e.stack[e.sp] = e.pc
		e.sp++
		e.pc = op & 0xfff
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
	case 0xa000:
		n := op & 0x0fff
		e.i = n
		e.pc += 2
	case 0xb000:
		n := op & 0x0fff
		e.pc = uint16(e.v[0]) + n
	case 0xc000:
		x := (op & 0x0f00) >> 8
		n := op & 0x00ff

		var r uint32

		if e.rng != nil {
			r = e.rng()
		} else {
			r = rand.Uint32()
		}

		e.v[x] = uint8(r) & uint8(n)
		e.pc += 2
	case 0xd000:
		x := (op & 0x0f00) >> 8
		y := (op & 0x00f0) >> 4
		n := op & 0x000f

		e.v[0xf] = 0

		bx := e.v[x]
		by := e.v[y]

		for dy := range n {
			sprite := e.memory[e.i+dy]

			for dx := range 8 {
				px := int(bx) + int(dx)
				py := int(by) + int(dy)

				if bit := sprite & (0x80 >> dx); bit != 0 {
					if e.display[py][px] != 0 {
						e.v[0xf] = 1
					}

					e.display[py][px] ^= 1
				}
			}
		}

		e.pc += 2
	case 0xe000:
		x := (op & 0x0f00) >> 8

		kind := op & 0x00ff

		switch kind {
		case 0x009e:
			if e.keyDown && e.key == e.v[x] {
				e.pc += 4
			} else {
				e.pc += 2
			}
		case 0x00a1:
			if !e.keyDown || e.key != e.v[x] {
				e.pc += 4
			} else {
				e.pc += 2
			}
		}
	case 0xf000:
		x := (op & 0x0f00) >> 8

		kind := op & 0x00ff

		switch kind {
		case 0x0007:
			e.v[x] = e.dt
			e.pc += 2
		case 0x0000a:
			if e.lastKeyWait {
				if e.lastKeySet {
					e.v[x] = e.lastKey
					e.lastKeyWait = false
					e.lastKeySet = false
					e.pc += 2
				}
			} else {
				e.lastKeyWait = true
			}
		case 0x0015:
			e.dt = e.v[x]
			e.pc += 2
		case 0x0018:
			e.st = e.v[x]
			e.pc += 2
		case 0x001e:
			e.i += uint16(e.v[x])
			e.pc += 2
		case 0x0029:
			e.i = uint16(5 * e.v[x])
			e.pc += 2
		case 0x0033:
			e.memory[e.i] = e.v[x] / 100
			e.memory[e.i+1] = (e.v[x] % 100) / 10
			e.memory[e.i+2] = e.v[x] % 10
			e.pc += 2
		case 0x0055:
			for n := range x + 1 {
				e.memory[e.i+n] = e.v[n]
			}
			e.pc += 2
		case 0x0065:
			for n := range x + 1 {
				e.v[n] = e.memory[e.i+n]
			}
			e.pc += 2
		}
	}

	return true
}
