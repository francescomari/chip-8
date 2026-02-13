package debug_test

import (
	"strings"
	"testing"

	"github.com/francescomari/chip-8/debug"
	"github.com/francescomari/chip-8/emulator"
)

func TestInstructionString(t *testing.T) {
	tests := []struct {
		op   uint16
		want string
	}{
		// Sys
		{emulator.OpCLS, "cls"},
		{emulator.OpRET, "ret"},

		// JP, CALL
		{0x1234, "jp 234"},
		{0x2456, "call 456"},

		// SE, SNE (constant)
		{0x3a12, "se va, 12"},
		{0x4b34, "sne vb, 34"},

		// SE, SNE (register)
		{0x5120, "se v1, v2"},
		{0x9340, "sne v3, v4"},

		// LD, ADD (constant)
		{0x6a42, "ld va, 42"},
		{0x7b10, "add vb, 10"},

		// ALU
		{0x8120, "ld v1, v2"},
		{0x8121, "or v1, v2"},
		{0x8122, "and v1, v2"},
		{0x8123, "xor v1, v2"},
		{0x8124, "add v1, v2"},
		{0x8125, "sub v1, v2"},
		{0x8126, "shr v1, v2"},
		{0x8127, "subn v1, v2"},
		{0x812e, "shl v1, v2"},

		// LD I, JP V0, RND, DRW
		{0xa123, "ld i, 123"},
		{0xb456, "jp v0, 456"},
		{0xc1ff, "rnd v1, ff"},
		{0xd125, "draw v1, v2, 5"},

		// Key
		{0xe19e, "skp v1"},
		{0xe1a1, "sknp v1"},

		// Misc
		{0xf107, "ld v1, dt"},
		{0xf10a, "ld v1, k"},
		{0xf115, "ld dt, v1"},
		{0xf118, "ld st, v1"},
		{0xf11e, "add i, v1"},
		{0xf129, "ld f, v1"},
		{0xf133, "ld b, v1"},
		{0xf155, "ld [i], v1"},
		{0xf165, "ld v1, [i]"},

		// Unknown
		{0x8009, "unknown (8009)"},
		{0xe1ff, "unknown (e1ff)"},
		{0xf1ff, "unknown (f1ff)"},
		{0x0001, "unknown (0001)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := debug.Instruction(tt.op).String()
			if got != tt.want {
				t.Errorf("Instruction(%04x).String() = %q, want %q", tt.op, got, tt.want)
			}
		})
	}
}

func TestPrintRegisters(t *testing.T) {
	var state emulator.State
	state.V[0] = 0x12
	state.V[1] = 0x34
	state.V[15] = 0xff

	var b strings.Builder
	debug.PrintRegisters(&b, &state)

	got := b.String()

	if !strings.Contains(got, "v0 = 12") {
		t.Errorf("expected v0 = 12 in output, got %q", got)
	}
	if !strings.Contains(got, "v1 = 34") {
		t.Errorf("expected v1 = 34 in output, got %q", got)
	}
	if !strings.Contains(got, "vf = ff") {
		t.Errorf("expected vf = ff in output, got %q", got)
	}
}

func TestPrintState(t *testing.T) {
	var state emulator.State
	state.I = 0x0abc
	state.SP = 0x02
	state.DT = 0x10
	state.ST = 0x20
	state.PC = 0x0200

	var b strings.Builder
	debug.PrintState(&b, &state)

	got := b.String()

	for _, want := range []string{"i = 0abc", "sp = 02", "dt = 10", "st = 20", "pc = 0200"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got %q", want, got)
		}
	}
}

func TestPrintInstruction(t *testing.T) {
	var state emulator.State
	state.PC = emulator.ProgramStart
	state.Memory[state.PC] = 0x00
	state.Memory[state.PC+1] = 0xe0

	var b strings.Builder
	debug.PrintInstruction(&b, &state)

	if got := b.String(); got != "cls" {
		t.Errorf("PrintInstruction = %q, want %q", got, "cls")
	}
}
