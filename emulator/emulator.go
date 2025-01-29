package emulator

import (
	"fmt"
	"math/rand/v2"
	"time"
)

var fonts [80]uint8 = [80]uint8{
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
)

type Emulator struct {
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
	drawx       time.Time     // The time after which the next draw is allowed
}

func New() *Emulator {
	var e Emulator
	e.Reset()
	return &e
}

func (e *Emulator) Memory(buffer []uint8) {
	copy(buffer, e.memory[:])
}

func (e *Emulator) V(buffer []uint8) {
	copy(buffer, e.v[:])
}

func (e *Emulator) I() uint16 {
	return e.i
}

func (e *Emulator) Stack(buffer []uint16) {
	copy(buffer, e.stack[:])
}

func (e *Emulator) SP() uint8 {
	return e.sp
}

func (e *Emulator) DT() uint8 {
	return e.dt
}

func (e *Emulator) DTClock() {
	if e.dt != 0 {
		e.dt--
	}
}

func (e *Emulator) ST() uint8 {
	return e.st
}

func (e *Emulator) STClock() bool {
	if e.st == 0 {
		return false
	}

	e.st--

	return e.st == 0
}

func (e *Emulator) PC() uint16 {
	return e.pc
}

func (e *Emulator) Display(buffer *Display) {
	*buffer = e.display
}

func (e *Emulator) KeyDown(key uint8) {
	e.keyDown = true
	e.key = key

	if e.lastKeyWait && !e.lastKeySet {
		e.lastKey = key
		e.lastKeySet = true
	}
}

func (e *Emulator) KeyUp() {
	e.keyDown = false
}

func (e *Emulator) SetRNG(rng func() uint32) {
	e.rng = rng
}

func (e *Emulator) Reset() {
	for i := range len(e.memory) {
		e.memory[i] = 0
	}

	copy(e.memory[:], fonts[:])

	for i := range len(e.v) {
		e.v[i] = 0
	}

	for i := range len(e.stack) {
		e.stack[i] = 0
	}

	e.display = Display{}

	e.i = 0
	e.sp = 0
	e.dt = 0
	e.st = 0
	e.pc = 0x200
	e.keyDown = false
	e.lastKeyWait = false
	e.lastKeySet = false
}

func (e *Emulator) Load(program []uint8) {
	copy(e.memory[0x200:], program)
}

func (e *Emulator) Step() bool {
	hi := uint16(e.memory[e.pc])
	lo := uint16(e.memory[e.pc+1])
	op := (hi << 8) | lo

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
	e.display = Display{}
	e.pc += 2
}

func (e *Emulator) functionReturn() {
	e.sp--
	e.pc = e.stack[e.sp]
	e.pc += 2
}

func (e *Emulator) jump(op uint16) {
	e.pc = op & 0x0fff
}

func (e *Emulator) functionCall(op uint16) {
	e.stack[e.sp] = e.pc
	e.sp++
	e.pc = op & 0xfff
}

func (e *Emulator) skipIfConstantEqual(op uint16) {
	x := (op & 0x0f00) >> 8
	n := uint8(op & 0x00ff)

	if e.v[x] == n {
		e.pc += 4
	} else {
		e.pc += 2
	}
}

func (e *Emulator) skipIfConstantNotEqual(op uint16) {
	x := (op & 0x0f00) >> 8
	n := uint8(op & 0x00ff)

	if e.v[x] != n {
		e.pc += 4
	} else {
		e.pc += 2
	}
}

func (e *Emulator) skipIfRegisterEqual(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4

	if e.v[x] == e.v[y] {
		e.pc += 4
	} else {
		e.pc += 2
	}
}

func (e *Emulator) loadRegisterFromConstant(op uint16) {
	x := (op & 0x0f00) >> 8
	v := uint8(op & 0x00ff)
	e.v[x] = v
	e.pc += 2
}

func (e *Emulator) incrementRegister(op uint16) {
	x := (op & 0x0f00) >> 8
	v := uint8(op & 0x00ff)
	e.v[x] += v
	e.pc += 2
}

func (e *Emulator) loadRegisterFromRegister(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4
	e.v[x] = e.v[y]
	e.pc += 2
}

func (e *Emulator) bitwiseOr(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4
	e.v[x] |= e.v[y]
	e.v[0xf] = 0
	e.pc += 2
}

func (e *Emulator) bitwiseAnd(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4
	e.v[x] &= e.v[y]
	e.v[0xf] = 0
	e.pc += 2
}

func (e *Emulator) bitwiseXor(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4
	e.v[x] ^= e.v[y]
	e.v[0xf] = 0
	e.pc += 2
}

func (e *Emulator) addWithCarry(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4

	var carry bool

	if e.v[x] > 0xff-e.v[y] {
		carry = true
	}

	e.v[x] += e.v[y]

	if carry {
		e.v[0xf] = 1
	} else {
		e.v[0xf] = 0
	}

	e.pc += 2
}

func (e *Emulator) subtractRightWithBorrow(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4

	var noBorrow bool

	if e.v[x] >= e.v[y] {
		noBorrow = true
	}

	e.v[x] -= e.v[y]

	if noBorrow {
		e.v[0xf] = 1
	} else {
		e.v[0xf] = 0
	}

	e.pc += 2
}

func (e *Emulator) subtractLeftWithBorrow(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4

	var noBorrow bool

	if e.v[y] >= e.v[x] {
		noBorrow = true
	}

	e.v[x] = e.v[y] - e.v[x]

	if noBorrow {
		e.v[0xf] = 1
	} else {
		e.v[0xf] = 0
	}

	e.pc += 2
}

func (e *Emulator) shiftRight(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4

	e.v[x] = e.v[y]

	var carry bool

	if e.v[x]&0x01 != 0 {
		carry = true
	}

	e.v[x] >>= 1

	if carry {
		e.v[0xf] = 1
	} else {
		e.v[0xf] = 0
	}

	e.pc += 2
}

func (e *Emulator) shiftLeft(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4

	e.v[x] = e.v[y]

	var carry bool

	if e.v[x]&0x80 != 0 {
		carry = true
	}

	e.v[x] <<= 1

	if carry {
		e.v[0xf] = 1
	} else {
		e.v[0xf] = 0
	}

	e.pc += 2
}

func (e *Emulator) skipIfRegisterNotEqual(op uint16) {
	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4

	if e.v[x] != e.v[y] {
		e.pc += 4
	} else {
		e.pc += 2
	}
}

func (e *Emulator) loadIndex(op uint16) {
	n := op & 0x0fff
	e.i = n
	e.pc += 2
}

func (e *Emulator) jumpRelative(op uint16) {
	n := op & 0x0fff
	e.pc = uint16(e.v[0]) + n
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

	e.v[x] = uint8(r) & uint8(n)
	e.pc += 2
}

func (e *Emulator) throttleDraw() {
	if e.drawx.IsZero() {
		time.Sleep(time.Second / 60)
	} else if time.Now().Before(e.drawx) {
		time.Sleep(e.drawx.Sub(time.Now()))
	}

	e.drawx = time.Now().Add(time.Second / 60)
}

func (e *Emulator) draw(op uint16) {
	e.throttleDraw()

	x := (op & 0x0f00) >> 8
	y := (op & 0x00f0) >> 4
	n := op & 0x000f

	e.v[0xf] = 0

	bx := e.v[x] % DisplayWidth
	by := e.v[y] % DisplayHeight

	for dy := range n {
		sprite := e.memory[e.i+dy]

		py := int(by) + int(dy)

		if py >= DisplayHeight {
			break
		}

		for dx := range 8 {
			px := int(bx) + int(dx)

			if px >= DisplayWidth {
				break
			}

			if bit := sprite & (0x80 >> dx); bit != 0 {
				if e.display[py][px] != 0 {
					e.v[0xf] = 1
				}

				e.display[py][px] ^= 1
			}
		}
	}

	e.pc += 2
}

func (e *Emulator) skipIfKeyPressed(op uint16) {
	x := (op & 0x0f00) >> 8

	if e.keyDown && e.key == e.v[x] {
		e.pc += 4
	} else {
		e.pc += 2
	}
}

func (e *Emulator) skipIfKeyNotPressed(op uint16) {
	x := (op & 0x0f00) >> 8

	if !e.keyDown || e.key != e.v[x] {
		e.pc += 4
	} else {
		e.pc += 2
	}
}

func (e *Emulator) loadRegisterFromDelayTimer(op uint16) {
	x := (op & 0x0f00) >> 8
	e.v[x] = e.dt
	e.pc += 2
}

func (e *Emulator) waitForKeyPress(op uint16) {
	x := (op & 0x0f00) >> 8

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
}

func (e *Emulator) loadDelayTimer(op uint16) {
	x := (op & 0x0f00) >> 8
	e.dt = e.v[x]
	e.pc += 2
}

func (e *Emulator) loadSoundTimer(op uint16) {
	x := (op & 0x0f00) >> 8
	e.st = e.v[x]
	e.pc += 2
}

func (e *Emulator) incrementIndex(op uint16) {
	x := (op & 0x0f00) >> 8
	e.i += uint16(e.v[x])
	e.pc += 2
}

func (e *Emulator) loadIndexFromSprite(op uint16) {
	x := (op & 0x0f00) >> 8
	e.i = uint16(5 * e.v[x])
	e.pc += 2
}

func (e *Emulator) loadMemoryFromBCD(op uint16) {
	x := (op & 0x0f00) >> 8
	e.memory[e.i] = e.v[x] / 100
	e.memory[e.i+1] = (e.v[x] % 100) / 10
	e.memory[e.i+2] = e.v[x] % 10
	e.pc += 2
}

func (e *Emulator) loadMemoryFromRegisters(op uint16) {
	x := (op & 0x0f00) >> 8

	for n := range x + 1 {
		e.memory[e.i] = e.v[n]
		e.i++
	}

	e.pc += 2
}

func (e *Emulator) loadRegistersFromMemory(op uint16) {
	x := (op & 0x0f00) >> 8

	for n := range x + 1 {
		e.v[n] = e.memory[e.i]
		e.i++
	}

	e.pc += 2
}
