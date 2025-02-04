package emulator

import (
	"fmt"
	"math/rand/v2"
)

var fonts = [80]uint8{
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

const (
	DisplayWidth  = 64
	DisplayHeight = 32
)

type (
	Memory    [4096]uint8
	Registers [16]uint8
	Stack     [16]uint16
	Display   [DisplayHeight][DisplayWidth]uint8
	Keys      [16]bool
)

type State struct {
	V       Registers // General-purpose registers
	I       uint16    // Index register
	SP      uint8     // Index to the next available stack entry
	DT      uint8     // Delay timer
	ST      uint8     // Sound timer
	PC      uint16    // Program counter
	Stack   Stack     // The stack
	Memory  Memory    // The memory
	Display Display   // The display
	Keys    Keys      // Currently pressed keys
}

func (s *State) Instruction() uint16 {
	return uint16(s.Memory[s.PC])<<8 | uint16(s.Memory[s.PC+1])
}

type Emulator struct {
	state           State
	waitKey         bool          // Waiting for a key press?
	waitKeyRegister uint8         // Where to store the pressed key, if waiting
	rng             func() uint32 // Random number generator
	sound           func()        // Callback called when the sound timer
}

func New() *Emulator {
	var e Emulator

	// Copy the fonts to the beginning of the memory.
	copy(e.state.Memory[:], fonts[:])

	// Set the program counter to the beginning of the program's memory.
	e.state.PC = 0x200

	return &e
}

func (e *Emulator) State(state *State) {
	*state = e.state
}

func (e *Emulator) Clock() {
	if e.state.DT > 0 {
		e.state.DT--
	}

	if e.state.ST > 0 {
		e.state.ST--

		if e.state.ST == 0 && e.sound != nil {
			e.sound()
		}
	}
}

func (e *Emulator) KeyDown(key uint8) {
	e.state.Keys[key&0xf] = true
}

func (e *Emulator) KeyUp(key uint8) {
	e.state.Keys[key&0xf] = false

	if e.waitKey {
		e.state.V[e.waitKeyRegister] = key
		e.waitKey = false
		e.state.PC += 2
	}
}

func (e *Emulator) SetRNG(rng func() uint32) {
	e.rng = rng
}

func (e *Emulator) SetSound(sound func()) {
	e.sound = sound
}

func (e *Emulator) Load(program []uint8) {
	copy(e.state.Memory[0x200:], program)
}

func (e *Emulator) Step() bool {
	op := e.state.Instruction()

	// The opcode 0NNN jumps to a machine code routine at address NNN, but it is
	// only used on the computers on which CHIP-8 was implemented. This
	// interpreter implements an opcode of this form as a HALT instruction.

	switch op & 0xf000 {
	case 0x0000:
		switch op & 0x00ff {
		case 0x00e0:
			e.clearDisplay()
		case 0x00ee:
			e.functionReturn()
		case 0x0000:
			return false
		default:
			panic(fmt.Sprintf("invalid opcode: %x", op))
		}
	case 0x1000:
		e.jump(op)
	case 0x2000:
		e.functionCall(op)
	case 0x3000:
		e.skipIfConstantEqual(op)
	case 0x4000:
		e.skipIfConstantNotEqual(op)
	case 0x5000:
		e.skipIfRegisterEqual(op)
	case 0x6000:
		e.loadRegisterFromConstant(op)
	case 0x7000:
		e.incrementRegister(op)
	case 0x8000:
		switch op & 0x000f {
		case 0x0000:
			e.loadRegisterFromRegister(op)
		case 0x0001:
			e.bitwiseOr(op)
		case 0x0002:
			e.bitwiseAnd(op)
		case 0x0003:
			e.bitwiseXor(op)
		case 0x0004:
			e.addWithCarry(op)
		case 0x0005:
			e.subtractRightWithBorrow(op)
		case 0x0006:
			e.shiftRight(op)
		case 0x0007:
			e.subtractLeftWithBorrow(op)
		case 0x000e:
			e.shiftLeft(op)
		default:
			panic(fmt.Sprintf("invalid opcode: %x", op))
		}
	case 0x9000:
		e.skipIfRegisterNotEqual(op)
	case 0xa000:
		e.loadIndex(op)
	case 0xb000:
		e.jumpRelative(op)
	case 0xc000:
		e.generateRandomNumber(op)
	case 0xd000:
		e.draw(op)
	case 0xe000:
		switch op & 0x00ff {
		case 0x009e:
			e.skipIfKeyPressed(op)
		case 0x00a1:
			e.skipIfKeyNotPressed(op)
		default:
			panic(fmt.Sprintf("invalid opcode: %x", op))
		}
	case 0xf000:
		switch op & 0x00ff {
		case 0x0007:
			e.loadRegisterFromDelayTimer(op)
		case 0x0000a:
			e.waitForKeyPress(op)
		case 0x0015:
			e.loadDelayTimer(op)
		case 0x0018:
			e.loadSoundTimer(op)
		case 0x001e:
			e.incrementIndex(op)
		case 0x0029:
			e.loadIndexFromSprite(op)
		case 0x0033:
			e.loadMemoryFromBCD(op)
		case 0x0055:
			e.loadMemoryFromRegisters(op)
		case 0x0065:
			e.loadRegistersFromMemory(op)
		default:
			panic(fmt.Sprintf("invalid opcode: %x", op))
		}
	}

	return true
}

func (e *Emulator) clearDisplay() {
	e.state.Display = Display{}
	e.state.PC += 2
}

func (e *Emulator) functionReturn() {
	e.state.SP--
	e.state.PC = e.state.Stack[e.state.SP]
	e.state.PC += 2
}

func (e *Emulator) jump(op uint16) {
	e.state.PC = op & 0x0fff
}

func (e *Emulator) functionCall(op uint16) {
	e.state.Stack[e.state.SP] = e.state.PC
	e.state.SP++
	e.state.PC = op & 0xfff
}

func (e *Emulator) skipIfConstantEqual(op uint16) {
	x := (op & 0x0f00) >> 8
	n := uint8(op & 0x00ff)

	if e.state.V[x] == n {
		e.state.PC += 4
	} else {
		e.state.PC += 2
	}
}

func (e *Emulator) skipIfConstantNotEqual(op uint16) {
	x := (op & 0x0f00) >> 8
	n := uint8(op & 0x00ff)

	if e.state.V[x] != n {
		e.state.PC += 4
	} else {
		e.state.PC += 2
	}
}

func (e *Emulator) skipIfRegisterEqual(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4

	if e.state.V[x] == e.state.V[y] {
		e.state.PC += 4
	} else {
		e.state.PC += 2
	}
}

func (e *Emulator) loadRegisterFromConstant(op uint16) {
	x := (op & 0x0f00) >> 8
	v := uint8(op & 0x00ff)
	e.state.V[x] = v
	e.state.PC += 2
}

func (e *Emulator) incrementRegister(op uint16) {
	x := (op & 0x0f00) >> 8
	v := uint8(op & 0x00ff)
	e.state.V[x] += v
	e.state.PC += 2
}

func (e *Emulator) loadRegisterFromRegister(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4
	e.state.V[x] = e.state.V[y]
	e.state.PC += 2
}

func (e *Emulator) bitwiseOr(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4
	e.state.V[x] |= e.state.V[y]
	e.state.V[0xf] = 0
	e.state.PC += 2
}

func (e *Emulator) bitwiseAnd(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4
	e.state.V[x] &= e.state.V[y]
	e.state.V[0xf] = 0
	e.state.PC += 2
}

func (e *Emulator) bitwiseXor(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4
	e.state.V[x] ^= e.state.V[y]
	e.state.V[0xf] = 0
	e.state.PC += 2
}

func (e *Emulator) addWithCarry(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4

	var carry bool

	if e.state.V[x] > 0xff-e.state.V[y] {
		carry = true
	}

	e.state.V[x] += e.state.V[y]

	if carry {
		e.state.V[0xf] = 1
	} else {
		e.state.V[0xf] = 0
	}

	e.state.PC += 2
}

func (e *Emulator) subtractRightWithBorrow(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4

	var noBorrow bool

	if e.state.V[x] >= e.state.V[y] {
		noBorrow = true
	}

	e.state.V[x] -= e.state.V[y]

	if noBorrow {
		e.state.V[0xf] = 1
	} else {
		e.state.V[0xf] = 0
	}

	e.state.PC += 2
}

func (e *Emulator) subtractLeftWithBorrow(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4

	var noBorrow bool

	if e.state.V[y] >= e.state.V[x] {
		noBorrow = true
	}

	e.state.V[x] = e.state.V[y] - e.state.V[x]

	if noBorrow {
		e.state.V[0xf] = 1
	} else {
		e.state.V[0xf] = 0
	}

	e.state.PC += 2
}

func (e *Emulator) shiftRight(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4

	e.state.V[x] = e.state.V[y]

	var carry bool

	if e.state.V[x]&0x01 != 0 {
		carry = true
	}

	e.state.V[x] >>= 1

	if carry {
		e.state.V[0xf] = 1
	} else {
		e.state.V[0xf] = 0
	}

	e.state.PC += 2
}

func (e *Emulator) shiftLeft(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4

	e.state.V[x] = e.state.V[y]

	var carry bool

	if e.state.V[x]&0x80 != 0 {
		carry = true
	}

	e.state.V[x] <<= 1

	if carry {
		e.state.V[0xf] = 1
	} else {
		e.state.V[0xf] = 0
	}

	e.state.PC += 2
}

func (e *Emulator) skipIfRegisterNotEqual(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4

	if e.state.V[x] != e.state.V[y] {
		e.state.PC += 4
	} else {
		e.state.PC += 2
	}
}

func (e *Emulator) loadIndex(op uint16) {
	n := op & 0x0fff
	e.state.I = n
	e.state.PC += 2
}

func (e *Emulator) jumpRelative(op uint16) {
	n := op & 0x0fff
	e.state.PC = uint16(e.state.V[0]) + n
}

func (e *Emulator) generateRandomNumber(op uint16) {
	x := (op & 0x0f00) >> 8
	n := op & 0x00ff

	var r uint32

	if e.rng != nil {
		r = e.rng()
	} else {
		r = rand.Uint32()
	}

	e.state.V[x] = uint8(r) & uint8(n)
	e.state.PC += 2
}

func (e *Emulator) draw(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4
	n := op & 0x000f

	e.state.V[0xf] = 0

	bx := e.state.V[x] % DisplayWidth
	by := e.state.V[y] % DisplayHeight

	for dy := range n {
		sprite := e.state.Memory[e.state.I+dy]

		py := int(by) + int(dy)

		if py >= DisplayHeight {
			break
		}

		for dx := range 8 {
			px := int(bx) + dx

			if px >= DisplayWidth {
				break
			}

			if bit := sprite & (0x80 >> dx); bit != 0 {
				if e.state.Display[py][px] != 0 {
					e.state.V[0xf] = 1
				}

				e.state.Display[py][px] ^= 1
			}
		}
	}

	e.state.PC += 2
}

func (e *Emulator) skipIfKeyPressed(op uint16) {
	x := (op & 0x0f00) >> 8
	k := e.state.V[x] & 0xf

	if e.state.Keys[k] {
		e.state.PC += 4
	} else {
		e.state.PC += 2
	}
}

func (e *Emulator) skipIfKeyNotPressed(op uint16) {
	x := (op & 0x0f00) >> 8
	k := e.state.V[x] & 0xf

	if e.state.Keys[k] {
		e.state.PC += 2
	} else {
		e.state.PC += 4
	}
}

func (e *Emulator) loadRegisterFromDelayTimer(op uint16) {
	x := (op & 0x0f00) >> 8
	e.state.V[x] = e.state.DT
	e.state.PC += 2
}

func (e *Emulator) waitForKeyPress(op uint16) {
	x := (op & 0x0f00) >> 8
	e.waitKey = true
	e.waitKeyRegister = uint8(x)
}

func (e *Emulator) loadDelayTimer(op uint16) {
	x := (op & 0x0f00) >> 8
	e.state.DT = e.state.V[x]
	e.state.PC += 2
}

func (e *Emulator) loadSoundTimer(op uint16) {
	x := (op & 0x0f00) >> 8
	e.state.ST = e.state.V[x]
	e.state.PC += 2
}

func (e *Emulator) incrementIndex(op uint16) {
	x := (op & 0x0f00) >> 8
	e.state.I += uint16(e.state.V[x])
	e.state.PC += 2
}

func (e *Emulator) loadIndexFromSprite(op uint16) {
	x := (op & 0x0f00) >> 8
	e.state.I = uint16(5 * e.state.V[x])
	e.state.PC += 2
}

func (e *Emulator) loadMemoryFromBCD(op uint16) {
	x := (op & 0x0f00) >> 8
	e.state.Memory[e.state.I] = e.state.V[x] / 100
	e.state.Memory[e.state.I+1] = (e.state.V[x] % 100) / 10
	e.state.Memory[e.state.I+2] = e.state.V[x] % 10
	e.state.PC += 2
}

func (e *Emulator) loadMemoryFromRegisters(op uint16) {
	x := (op & 0x0f00) >> 8

	for n := range x + 1 {
		e.state.Memory[e.state.I] = e.state.V[n]
		e.state.I++
	}

	e.state.PC += 2
}

func (e *Emulator) loadRegistersFromMemory(op uint16) {
	x := (op & 0x0f00) >> 8

	for n := range x + 1 {
		e.state.V[n] = e.state.Memory[e.state.I]
		e.state.I++
	}

	e.state.PC += 2
}
