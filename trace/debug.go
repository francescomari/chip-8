package trace

import (
	"fmt"
	"io"

	"github.com/francescomari/chip-8/emulator"
)

func PrintRegisters(w io.Writer, state *emulator.State) {
	for i, v := range state.V {
		if i > 0 {
			fmt.Fprintf(w, ", v%x = %02x", i, v)
		} else {
			fmt.Fprintf(w, "v%x = %02x", i, v)
		}
	}
}

func PrintState(w io.Writer, state *emulator.State) {
	fmt.Fprintf(w, "i = %04x, ", state.I)
	fmt.Fprintf(w, "sp = %02x, ", state.SP)
	fmt.Fprintf(w, "dt = %02x, ", state.DT)
	fmt.Fprintf(w, "st = %02x, ", state.ST)
	fmt.Fprintf(w, "pc = %04x", state.PC)
}

func PrintInstruction(w io.Writer, state *emulator.State) {
	var (
		op = state.Instruction()
		x  = fmt.Sprintf("v%x", (op&0x0f00)>>8)
		y  = fmt.Sprintf("v%x", (op&0x00f0)>>4)
		n  = fmt.Sprintf("%03x", op&0x0fff)
		k  = fmt.Sprintf("%02x", op&0x00ff)
		b  = fmt.Sprintf("%x", op&0x000f)
	)

	switch op & 0xf000 {
	case 0x0000:
		switch op & 0x00ff {
		case 0x00e0:
			fmt.Fprintf(w, "cls")
		case 0x00ee:
			fmt.Fprintf(w, "ret")
		}
	case 0x1000:
		fmt.Fprintf(w, "jp %s", n)
	case 0x2000:
		fmt.Fprintf(w, "call %s", n)
	case 0x3000:
		fmt.Fprintf(w, "se %s, %s", x, k)
	case 0x4000:
		fmt.Fprintf(w, "sne %s, %s", x, k)
	case 0x5000:
		fmt.Fprintf(w, "se %s, %s", x, y)
	case 0x6000:
		fmt.Fprintf(w, "ld %s, %s", x, k)
	case 0x7000:
		fmt.Fprintf(w, "add %s, %s", x, k)
	case 0x8000:
		switch op & 0x000f {
		case 0x0:
			fmt.Fprintf(w, "ld %s, %s", x, y)
		case 0x1:
			fmt.Fprintf(w, "or %s, %s", x, y)
		case 0x2:
			fmt.Fprintf(w, "and %s, %s", x, y)
		case 0x3:
			fmt.Fprintf(w, "xor %s, %s", x, y)
		case 0x4:
			fmt.Fprintf(w, "add %s, %s", x, y)
		case 0x5:
			fmt.Fprintf(w, "sub %s, %s", x, y)
		case 0x6:
			fmt.Fprintf(w, "shr %s, %s", x, y)
		case 0x7:
			fmt.Fprintf(w, "subn %s, %s", x, y)
		case 0xe:
			fmt.Fprintf(w, "shl %s, %s", x, y)
		}
	case 0x9000:
		fmt.Fprintf(w, "sne %s, %s", x, y)
	case 0xa000:
		fmt.Fprintf(w, "ld i, %s", n)
	case 0xb000:
		fmt.Fprintf(w, "jp v0, %s", n)
	case 0xc000:
		fmt.Fprintf(w, "rnd %s, %s", x, k)
	case 0xd000:
		fmt.Fprintf(w, "draw %s, %s, %s", x, y, b)
	case 0xe000:
		switch op & 0xff {
		case 0x9e:
			fmt.Fprintf(w, "skp %s", x)
		case 0xa1:
			fmt.Fprintf(w, "sknp %s", x)
		}
	case 0xf000:
		switch op & 0xff {
		case 0x07:
			fmt.Fprintf(w, "ld %s, dt", x)
		case 0x0a:
			fmt.Fprintf(w, "ld %s, k", x)
		case 0x15:
			fmt.Fprintf(w, "ld dt, %s", x)
		case 0x18:
			fmt.Fprintf(w, "ld st, %s", x)
		case 0x1e:
			fmt.Fprintf(w, "add i, %s", x)
		case 0x29:
			fmt.Fprintf(w, "ld f, %s", x)
		case 0x33:
			fmt.Fprintf(w, "ld b, %s", x)
		case 0x55:
			fmt.Fprintf(w, "ld [i], %s", x)
		case 0x65:
			fmt.Fprintf(w, "ld %s, [i]", x)
		}
	default:
		fmt.Fprintf(w, "unknown (%04x)", op)
	}
}
