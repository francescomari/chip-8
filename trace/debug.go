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

func PrintState(w io.Writer, state *emulator.State) {
	out := printer(w)

	out("i = %04x, ", state.I)
	out("sp = %02x, ", state.SP)
	out("dt = %02x, ", state.DT)
	out("st = %02x, ", state.ST)
	out("pc = %04x", state.PC)
}

func PrintInstruction(w io.Writer, state *emulator.State) {
	var (
		op  = state.Instruction()
		x   = fmt.Sprintf("v%x", (op&0x0f00)>>8)
		y   = fmt.Sprintf("v%x", (op&0x00f0)>>4)
		n   = fmt.Sprintf("%03x", op&0x0fff)
		k   = fmt.Sprintf("%02x", op&0x00ff)
		b   = fmt.Sprintf("%x", op&0x000f)
		out = printer(w)
	)

	switch op & 0xf000 {
	case 0x0000:
		switch op & 0x00ff {
		case 0x00e0:
			out("cls")
		case 0x00ee:
			out("ret")
		}
	case 0x1000:
		out("jp %s", n)
	case 0x2000:
		out("call %s", n)
	case 0x3000:
		out("se %s, %s", x, k)
	case 0x4000:
		out("sne %s, %s", x, k)
	case 0x5000:
		out("se %s, %s", x, y)
	case 0x6000:
		out("ld %s, %s", x, k)
	case 0x7000:
		out("add %s, %s", x, k)
	case 0x8000:
		switch op & 0x000f {
		case 0x0:
			out("ld %s, %s", x, y)
		case 0x1:
			out("or %s, %s", x, y)
		case 0x2:
			out("and %s, %s", x, y)
		case 0x3:
			out("xor %s, %s", x, y)
		case 0x4:
			out("add %s, %s", x, y)
		case 0x5:
			out("sub %s, %s", x, y)
		case 0x6:
			out("shr %s, %s", x, y)
		case 0x7:
			out("subn %s, %s", x, y)
		case 0xe:
			out("shl %s, %s", x, y)
		}
	case 0x9000:
		out("sne %s, %s", x, y)
	case 0xa000:
		out("ld i, %s", n)
	case 0xb000:
		out("jp v0, %s", n)
	case 0xc000:
		out("rnd %s, %s", x, k)
	case 0xd000:
		out("draw %s, %s, %s", x, y, b)
	case 0xe000:
		switch op & 0xff {
		case 0x9e:
			out("skp %s", x)
		case 0xa1:
			out("sknp %s", x)
		}
	case 0xf000:
		switch op & 0xff {
		case 0x07:
			out("ld %s, dt", x)
		case 0x0a:
			out("ld %s, k", x)
		case 0x15:
			out("ld dt, %s", x)
		case 0x18:
			out("ld st, %s", x)
		case 0x1e:
			out("add i, %s", x)
		case 0x29:
			out("ld f, %s", x)
		case 0x33:
			out("ld b, %s", x)
		case 0x55:
			out("ld [i], %s", x)
		case 0x65:
			out("ld %s, [i]", x)
		}
	default:
		out("unknown (%04x)", op)
	}
}
