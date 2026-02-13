package emulator

import (
	"fmt"
	"math/rand/v2"
)

var fonts = [16 * FontSize]uint8{
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

// ProgramStart is the address in memory where programs are loaded and executed.
const ProgramStart = 0x200

// FontSize is the number of bytes in each font sprite.
const FontSize = 5

// Display and sprite geometry.
const (
	DisplayWidth  = 64 // Width of the display in pixels.
	DisplayHeight = 32 // Height of the display in pixels.
	SpriteWidth   = 8  // Width of a sprite in pixels.
)

// Masks for extracting parts of an opcode.
const (
	MaskFamily = 0xf000 // High nibble: instruction family
	MaskX      = 0x0f00 // Second nibble: first register index (Vx)
	MaskY      = 0x00f0 // Third nibble: second register index (Vy)
	MaskN      = 0x000f // Low nibble: 4-bit constant or ALU sub-opcode
	MaskKK     = 0x00ff // Low byte: 8-bit constant or sys/key/misc sub-opcode
	MaskNNN    = 0x0fff // Low 12 bits: memory address
)

// Shifts for converting a masked opcode field to an integer index.
const (
	ShiftX = 8 // MaskX >> ShiftX yields the Vx register index
	ShiftY = 4 // MaskY >> ShiftY yields the Vy register index
)

// Instruction families, matched against op & [MaskFamily].
const (
	OpTypeSys  = 0x0000 // System instructions: CLS, RET, and HALT.
	OpTypeJP   = 0x1000 // JP addr: jump to address NNN.
	OpTypeCALL = 0x2000 // CALL addr: call subroutine at address NNN.
	OpTypeSE   = 0x3000 // SE Vx, byte: skip next instruction if Vx == KK.
	OpTypeSNE  = 0x4000 // SNE Vx, byte: skip next instruction if Vx != KK.
	OpTypeSEV  = 0x5000 // SE Vx, Vy: skip next instruction if Vx == Vy.
	OpTypeLD   = 0x6000 // LD Vx, byte: load KK into Vx.
	OpTypeADD  = 0x7000 // ADD Vx, byte: add KK to Vx.
	OpTypeALU  = 0x8000 // Register-to-register ALU instructions.
	OpTypeSNEV = 0x9000 // SNE Vx, Vy: skip next instruction if Vx != Vy.
	OpTypeLDI  = 0xa000 // LD I, addr: load address NNN into I.
	OpTypeJPV  = 0xb000 // JP V0, addr: jump to address NNN + V0.
	OpTypeRND  = 0xc000 // RND Vx, byte: load a random byte AND KK into Vx.
	OpTypeDRW  = 0xd000 // DRW Vx, Vy, nibble: draw an N-byte sprite at (Vx, Vy).
	OpTypeKey  = 0xe000 // Key-state instructions: SKP and SKNP.
	OpTypeMisc = 0xf000 // Miscellaneous instructions operating on timers, memory, and I.
)

// Sub-opcodes for [OpTypeSys], matched against op & [MaskKK].
const (
	OpHALT = 0x0000 // Halt the emulator. Non-standard extension.
	OpCLS  = 0x00e0 // CLS: clear the display.
	OpRET  = 0x00ee // RET: return from a subroutine.
)

// Sub-opcodes for [OpTypeALU], matched against op & [MaskN].
const (
	OpLDVV  = 0x0000 // LD Vx, Vy: set Vx = Vy.
	OpORVV  = 0x0001 // OR Vx, Vy: set Vx = Vx | Vy.
	OpANDVV = 0x0002 // AND Vx, Vy: set Vx = Vx & Vy.
	OpXORVV = 0x0003 // XOR Vx, Vy: set Vx = Vx ^ Vy.
	OpADDVV = 0x0004 // ADD Vx, Vy: set Vx = Vx + Vy, VF = carry.
	OpSUBVV = 0x0005 // SUB Vx, Vy: set Vx = Vx - Vy, VF = not borrow.
	OpSHR   = 0x0006 // SHR Vx, Vy: set Vx = Vy >> 1, VF = shifted-out bit.
	OpSUBN  = 0x0007 // SUBN Vx, Vy: set Vx = Vy - Vx, VF = not borrow.
	OpSHL   = 0x000e // SHL Vx, Vy: set Vx = Vy << 1, VF = shifted-out bit.
)

// Sub-opcodes for [OpTypeKey], matched against op & [MaskKK].
const (
	OpSKP  = 0x009e // SKP Vx: skip next instruction if key Vx is pressed.
	OpSKNP = 0x00a1 // SKNP Vx: skip next instruction if key Vx is not pressed.
)

// Sub-opcodes for [OpTypeMisc], matched against op & [MaskKK].
const (
	OpLDVDT = 0x0007 // LD Vx, DT: load the delay timer value into Vx.
	OpLDVK  = 0x000a // LD Vx, K: wait for a key press and load the key into Vx.
	OpLDDTV = 0x0015 // LD DT, Vx: load Vx into the delay timer.
	OpLDSTV = 0x0018 // LD ST, Vx: load Vx into the sound timer.
	OpADDIV = 0x001e // ADD I, Vx: set I = I + Vx.
	OpLDF   = 0x0029 // LD F, Vx: load the address of the sprite for digit Vx into I.
	OpLDB   = 0x0033 // LD B, Vx: store the BCD representation of Vx at I, I+1, I+2.
	OpSTMV  = 0x0055 // LD [I], Vx: store registers V0 through Vx in memory starting at I.
	OpLDVM  = 0x0065 // LD Vx, [I]: load registers V0 through Vx from memory starting at I.
)

type (
	// Memory is the 4096-byte addressable memory of the CHIP-8.
	Memory [4096]uint8
	// Registers holds the 16 general-purpose 8-bit registers V0 through VF.
	Registers [16]uint8
	// Stack holds the up to 16 return addresses pushed by CALL instructions.
	Stack [16]uint16
	// Display is the 64×32 monochrome pixel framebuffer.
	Display [DisplayHeight][DisplayWidth]uint8
	// Keys holds the pressed state of the 16 keys of the hexadecimal keypad.
	Keys [16]bool
)

// State is a snapshot of the complete CHIP-8 machine state.
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

// Instruction returns the 16-bit opcode at the current program counter.
func (s *State) Instruction() uint16 {
	return uint16(s.Memory[s.PC])<<8 | uint16(s.Memory[s.PC+1])
}

// Emulator is a CHIP-8 interpreter. Use [New] to create one.
type Emulator struct {
	state           State
	waitKey         bool          // Waiting for a key press?
	waitKeyRegister uint8         // Where to store the pressed key, if waiting
	rng             func() uint32 // Random number generator
	sound           func()        // Callback called when the sound timer expires
}

// New returns a new Emulator ready to execute a program loaded with [Emulator.Load].
func New() *Emulator {
	var e Emulator

	// Copy the fonts to the beginning of the memory.
	copy(e.state.Memory[:], fonts[:])

	// Set the program counter to the beginning of the program's memory.
	e.state.PC = ProgramStart

	return &e
}

// State copies the current machine state into the provided [State].
func (e *Emulator) State(state *State) {
	*state = e.state
}

// Clock advances the delay and sound timers by one tick. When the sound timer
// reaches zero, the sound callback registered with [Emulator.SetSound] is called.
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

// KeyDown records that key has been pressed. Only the low four bits of key are used.
func (e *Emulator) KeyDown(key uint8) {
	e.state.Keys[key&0xf] = true
}

// KeyUp records that key has been released. If the emulator is waiting for a key
// press (LD Vx, K), execution resumes and the key value is stored in Vx.
func (e *Emulator) KeyUp(key uint8) {
	e.state.Keys[key&0xf] = false

	if e.waitKey {
		e.state.V[e.waitKeyRegister] = key
		e.waitKey = false
		e.state.PC += 2
	}
}

// SetRNG sets the random number generator used by the RND instruction. If not
// set, the emulator uses the default source from math/rand/v2.
func (e *Emulator) SetRNG(rng func() uint32) {
	e.rng = rng
}

// SetSound registers a callback that is called once when the sound timer expires.
func (e *Emulator) SetSound(sound func()) {
	e.sound = sound
}

// Load copies program into memory starting at [ProgramStart].
func (e *Emulator) Load(program []uint8) {
	copy(e.state.Memory[ProgramStart:], program)
}

// Step decodes and executes the instruction at the current program counter.
// It returns true if execution should continue, or false if the emulator has
// halted. It returns an error if the instruction is not recognized.
func (e *Emulator) Step() (bool, error) {
	op := e.state.Instruction()

	// The opcode 0NNN jumps to a machine code routine at address NNN, but it is
	// only used on the computers on which CHIP-8 was implemented. This
	// interpreter implements an opcode of this form as a HALT instruction.

	switch op & MaskFamily {
	case OpTypeSys:
		switch op & MaskKK {
		case OpCLS:
			e.clearDisplay()
		case OpRET:
			e.functionReturn()
		case OpHALT:
			return false, nil
		default:
			return false, fmt.Errorf("invalid opcode: %04x", op)
		}
	case OpTypeJP:
		e.jump(op)
	case OpTypeCALL:
		e.functionCall(op)
	case OpTypeSE:
		e.skipIfConstantEqual(op)
	case OpTypeSNE:
		e.skipIfConstantNotEqual(op)
	case OpTypeSEV:
		e.skipIfRegisterEqual(op)
	case OpTypeLD:
		e.loadRegisterFromConstant(op)
	case OpTypeADD:
		e.incrementRegister(op)
	case OpTypeALU:
		switch op & MaskN {
		case OpLDVV:
			e.loadRegisterFromRegister(op)
		case OpORVV:
			e.bitwiseOr(op)
		case OpANDVV:
			e.bitwiseAnd(op)
		case OpXORVV:
			e.bitwiseXor(op)
		case OpADDVV:
			e.addWithCarry(op)
		case OpSUBVV:
			e.subtractRightWithBorrow(op)
		case OpSHR:
			e.shiftRight(op)
		case OpSUBN:
			e.subtractLeftWithBorrow(op)
		case OpSHL:
			e.shiftLeft(op)
		default:
			return false, fmt.Errorf("invalid opcode: %04x", op)
		}
	case OpTypeSNEV:
		e.skipIfRegisterNotEqual(op)
	case OpTypeLDI:
		e.loadIndex(op)
	case OpTypeJPV:
		e.jumpRelative(op)
	case OpTypeRND:
		e.generateRandomNumber(op)
	case OpTypeDRW:
		e.draw(op)
	case OpTypeKey:
		switch op & MaskKK {
		case OpSKP:
			e.skipIfKeyPressed(op)
		case OpSKNP:
			e.skipIfKeyNotPressed(op)
		default:
			return false, fmt.Errorf("invalid opcode: %04x", op)
		}
	case OpTypeMisc:
		switch op & MaskKK {
		case OpLDVDT:
			e.loadRegisterFromDelayTimer(op)
		case OpLDVK:
			e.waitForKeyPress(op)
		case OpLDDTV:
			e.loadDelayTimer(op)
		case OpLDSTV:
			e.loadSoundTimer(op)
		case OpADDIV:
			e.incrementIndex(op)
		case OpLDF:
			e.loadIndexFromSprite(op)
		case OpLDB:
			e.loadMemoryFromBCD(op)
		case OpSTMV:
			e.loadMemoryFromRegisters(op)
		case OpLDVM:
			e.loadRegistersFromMemory(op)
		default:
			return false, fmt.Errorf("invalid opcode: %04x", op)
		}
	}

	return true, nil
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
	e.state.PC = op & MaskNNN
}

func (e *Emulator) functionCall(op uint16) {
	e.state.Stack[e.state.SP] = e.state.PC
	e.state.SP++
	e.state.PC = op & 0xfff
}

func (e *Emulator) skipIfConstantEqual(op uint16) {
	x := (op & MaskX) >> ShiftX
	n := uint8(op & MaskKK)

	if e.state.V[x] == n {
		e.state.PC += 4
	} else {
		e.state.PC += 2
	}
}

func (e *Emulator) skipIfConstantNotEqual(op uint16) {
	x := (op & MaskX) >> ShiftX
	n := uint8(op & MaskKK)

	if e.state.V[x] != n {
		e.state.PC += 4
	} else {
		e.state.PC += 2
	}
}

func (e *Emulator) skipIfRegisterEqual(op uint16) {
	x := (op & MaskX) >> ShiftX
	y := (op & MaskY) >> ShiftY

	if e.state.V[x] == e.state.V[y] {
		e.state.PC += 4
	} else {
		e.state.PC += 2
	}
}

func (e *Emulator) loadRegisterFromConstant(op uint16) {
	x := (op & MaskX) >> ShiftX
	v := uint8(op & MaskKK)
	e.state.V[x] = v
	e.state.PC += 2
}

func (e *Emulator) incrementRegister(op uint16) {
	x := (op & MaskX) >> ShiftX
	v := uint8(op & MaskKK)
	e.state.V[x] += v
	e.state.PC += 2
}

func (e *Emulator) loadRegisterFromRegister(op uint16) {
	x := (op & MaskX) >> ShiftX
	y := (op & MaskY) >> ShiftY
	e.state.V[x] = e.state.V[y]
	e.state.PC += 2
}

func (e *Emulator) bitwiseOr(op uint16) {
	x := (op & MaskX) >> ShiftX
	y := (op & MaskY) >> ShiftY
	e.state.V[x] |= e.state.V[y]
	e.state.V[0xf] = 0
	e.state.PC += 2
}

func (e *Emulator) bitwiseAnd(op uint16) {
	x := (op & MaskX) >> ShiftX
	y := (op & MaskY) >> ShiftY
	e.state.V[x] &= e.state.V[y]
	e.state.V[0xf] = 0
	e.state.PC += 2
}

func (e *Emulator) bitwiseXor(op uint16) {
	x := (op & MaskX) >> ShiftX
	y := (op & MaskY) >> ShiftY
	e.state.V[x] ^= e.state.V[y]
	e.state.V[0xf] = 0
	e.state.PC += 2
}

func (e *Emulator) addWithCarry(op uint16) {
	x := (op & MaskX) >> ShiftX
	y := (op & MaskY) >> ShiftY

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
	x := (op & MaskX) >> ShiftX
	y := (op & MaskY) >> ShiftY

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
	x := (op & MaskX) >> ShiftX
	y := (op & MaskY) >> ShiftY

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
	x := (op & MaskX) >> ShiftX
	y := (op & MaskY) >> ShiftY

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
	x := (op & MaskX) >> ShiftX
	y := (op & MaskY) >> ShiftY

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
	x := (op & MaskX) >> ShiftX
	y := (op & MaskY) >> ShiftY

	if e.state.V[x] != e.state.V[y] {
		e.state.PC += 4
	} else {
		e.state.PC += 2
	}
}

func (e *Emulator) loadIndex(op uint16) {
	n := op & MaskNNN
	e.state.I = n
	e.state.PC += 2
}

func (e *Emulator) jumpRelative(op uint16) {
	n := op & MaskNNN
	e.state.PC = uint16(e.state.V[0]) + n
}

func (e *Emulator) generateRandomNumber(op uint16) {
	x := (op & MaskX) >> ShiftX
	n := op & MaskKK

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
	x := (op & MaskX) >> ShiftX
	y := (op & MaskY) >> ShiftY
	n := op & MaskN

	e.state.V[0xf] = 0

	bx := e.state.V[x] % DisplayWidth
	by := e.state.V[y] % DisplayHeight

	for dy := range n {
		sprite := e.state.Memory[e.state.I+dy]

		py := int(by) + int(dy)

		if py >= DisplayHeight {
			break
		}

		for dx := range SpriteWidth {
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
	x := (op & MaskX) >> ShiftX
	k := e.state.V[x] & 0xf

	if e.state.Keys[k] {
		e.state.PC += 4
	} else {
		e.state.PC += 2
	}
}

func (e *Emulator) skipIfKeyNotPressed(op uint16) {
	x := (op & MaskX) >> ShiftX
	k := e.state.V[x] & 0xf

	if e.state.Keys[k] {
		e.state.PC += 2
	} else {
		e.state.PC += 4
	}
}

func (e *Emulator) loadRegisterFromDelayTimer(op uint16) {
	x := (op & MaskX) >> ShiftX
	e.state.V[x] = e.state.DT
	e.state.PC += 2
}

func (e *Emulator) waitForKeyPress(op uint16) {
	x := (op & MaskX) >> ShiftX
	e.waitKey = true
	e.waitKeyRegister = uint8(x)
}

func (e *Emulator) loadDelayTimer(op uint16) {
	x := (op & MaskX) >> ShiftX
	e.state.DT = e.state.V[x]
	e.state.PC += 2
}

func (e *Emulator) loadSoundTimer(op uint16) {
	x := (op & MaskX) >> ShiftX
	e.state.ST = e.state.V[x]
	e.state.PC += 2
}

func (e *Emulator) incrementIndex(op uint16) {
	x := (op & MaskX) >> ShiftX
	e.state.I += uint16(e.state.V[x])
	e.state.PC += 2
}

func (e *Emulator) loadIndexFromSprite(op uint16) {
	x := (op & MaskX) >> ShiftX
	e.state.I = uint16(FontSize * e.state.V[x])
	e.state.PC += 2
}

func (e *Emulator) loadMemoryFromBCD(op uint16) {
	x := (op & MaskX) >> ShiftX
	e.state.Memory[e.state.I] = e.state.V[x] / 100
	e.state.Memory[e.state.I+1] = (e.state.V[x] % 100) / 10
	e.state.Memory[e.state.I+2] = e.state.V[x] % 10
	e.state.PC += 2
}

func (e *Emulator) loadMemoryFromRegisters(op uint16) {
	x := (op & MaskX) >> ShiftX

	for n := range x + 1 {
		e.state.Memory[e.state.I] = e.state.V[n]
		e.state.I++
	}

	e.state.PC += 2
}

func (e *Emulator) loadRegistersFromMemory(op uint16) {
	x := (op & MaskX) >> ShiftX

	for n := range x + 1 {
		e.state.V[n] = e.state.Memory[e.state.I]
		e.state.I++
	}

	e.state.PC += 2
}
