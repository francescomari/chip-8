package trace

import (
	"fmt"
	"io"

	"github.com/francescomari/chip-8/emulator"
)

func printer(w io.Writer) func(string, ...any) {
	return func(s string, args ...any) {
		_, _ = fmt.Fprintf(w, s, args...)
	}
}

// PrintRegisters writes the values of all general-purpose registers (V0–VF)
// from state to w as a comma-separated list.
func PrintRegisters(w io.Writer, state *emulator.State) {
	out := printer(w)

	for i, v := range state.V {
		if i > 0 {
			out(", v%x = %02x", i, v)
		} else {
			out("v%x = %02x", i, v)
		}
	}
}

// PrintState writes the values of the special-purpose registers (I, SP, DT, ST,
// and PC) from state to w as a comma-separated list.
func PrintState(w io.Writer, state *emulator.State) {
	out := printer(w)

	out("i = %04x, ", state.I)
	out("sp = %02x, ", state.SP)
	out("dt = %02x, ", state.DT)
	out("st = %02x, ", state.ST)
	out("pc = %04x", state.PC)
}

// PrintInstruction writes the assembly mnemonic for the instruction at the
// current program counter of state to w.
func PrintInstruction(w io.Writer, state *emulator.State) {
	out := printer(w)

	out("%v", Instruction(state.Instruction()))
}

// Instruction wraps a raw instruction from the emulator's state and returns a
// printable representation of the opcode and its arguments.
type Instruction uint16

func (i Instruction) String() string {
	var (
		op = uint16(i)
		x  = fmt.Sprintf("v%x", (op&emulator.MaskX)>>emulator.ShiftX)
		y  = fmt.Sprintf("v%x", (op&emulator.MaskY)>>emulator.ShiftY)
		n  = fmt.Sprintf("%03x", op&emulator.MaskNNN)
		k  = fmt.Sprintf("%02x", op&emulator.MaskKK)
		b  = fmt.Sprintf("%x", op&emulator.MaskN)
	)

	switch op & emulator.MaskFamily {
	case emulator.OpTypeSys:
		switch op & emulator.MaskKK {
		case emulator.OpCLS:
			return "cls"
		case emulator.OpRET:
			return "ret"
		}
	case emulator.OpTypeJP:
		return fmt.Sprintf("jp %s", n)
	case emulator.OpTypeCALL:
		return fmt.Sprintf("call %s", n)
	case emulator.OpTypeSE:
		return fmt.Sprintf("se %s, %s", x, k)
	case emulator.OpTypeSNE:
		return fmt.Sprintf("sne %s, %s", x, k)
	case emulator.OpTypeSEV:
		return fmt.Sprintf("se %s, %s", x, y)
	case emulator.OpTypeLD:
		return fmt.Sprintf("ld %s, %s", x, k)
	case emulator.OpTypeADD:
		return fmt.Sprintf("add %s, %s", x, k)
	case emulator.OpTypeALU:
		switch op & emulator.MaskN {
		case emulator.OpLDVV:
			return fmt.Sprintf("ld %s, %s", x, y)
		case emulator.OpORVV:
			return fmt.Sprintf("or %s, %s", x, y)
		case emulator.OpANDVV:
			return fmt.Sprintf("and %s, %s", x, y)
		case emulator.OpXORVV:
			return fmt.Sprintf("xor %s, %s", x, y)
		case emulator.OpADDVV:
			return fmt.Sprintf("add %s, %s", x, y)
		case emulator.OpSUBVV:
			return fmt.Sprintf("sub %s, %s", x, y)
		case emulator.OpSHR:
			return fmt.Sprintf("shr %s, %s", x, y)
		case emulator.OpSUBN:
			return fmt.Sprintf("subn %s, %s", x, y)
		case emulator.OpSHL:
			return fmt.Sprintf("shl %s, %s", x, y)
		}
	case emulator.OpTypeSNEV:
		return fmt.Sprintf("sne %s, %s", x, y)
	case emulator.OpTypeLDI:
		return fmt.Sprintf("ld i, %s", n)
	case emulator.OpTypeJPV:
		return fmt.Sprintf("jp v0, %s", n)
	case emulator.OpTypeRND:
		return fmt.Sprintf("rnd %s, %s", x, k)
	case emulator.OpTypeDRW:
		return fmt.Sprintf("draw %s, %s, %s", x, y, b)
	case emulator.OpTypeKey:
		switch op & emulator.MaskKK {
		case emulator.OpSKP:
			return fmt.Sprintf("skp %s", x)
		case emulator.OpSKNP:
			return fmt.Sprintf("sknp %s", x)
		}
	case emulator.OpTypeMisc:
		switch op & emulator.MaskKK {
		case emulator.OpLDVDT:
			return fmt.Sprintf("ld %s, dt", x)
		case emulator.OpLDVK:
			return fmt.Sprintf("ld %s, k", x)
		case emulator.OpLDDTV:
			return fmt.Sprintf("ld dt, %s", x)
		case emulator.OpLDSTV:
			return fmt.Sprintf("ld st, %s", x)
		case emulator.OpADDIV:
			return fmt.Sprintf("add i, %s", x)
		case emulator.OpLDF:
			return fmt.Sprintf("ld f, %s", x)
		case emulator.OpLDB:
			return fmt.Sprintf("ld b, %s", x)
		case emulator.OpSTMV:
			return fmt.Sprintf("ld [i], %s", x)
		case emulator.OpLDVM:
			return fmt.Sprintf("ld %s, [i]", x)
		}
	}

	return fmt.Sprintf("unknown (%04x)", op)
}
