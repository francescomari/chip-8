package emulator_test

import (
	"slices"
	"testing"

	"github.com/francescomari/chip-8/emulator"
)

func TestInit(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	memory := e.Memory()

	if !slices.Equal(memory[:len(emulator.Fonts)], emulator.Fonts[:]) {
		t.Fatalf("fonts not loaded")
	}

	for _, v := range memory[len(emulator.Fonts):] {
		if v != 0 {
			t.Fatalf("memory not zeroed")
		}
	}

	for i, v := range e.V() {
		if v != 0 {
			t.Fatalf("register %d not zeroed", i)
		}
	}

	if e.I() != 0 {
		t.Fatalf("index register not zeroed")
	}

	for _, v := range e.Stack() {
		if v != 0 {
			t.Fatalf("stack not zeroed")
		}
	}

	if e.SP() != 0 {
		t.Fatalf("stack pointer not zeroed")
	}

	if e.DT() != 0 {
		t.Fatalf("delay timer not zeroed")
	}

	if e.ST() != 0 {
		t.Fatalf("sound timer not zeroed")
	}

	if e.PC() != 0x200 {
		t.Fatalf("program counter not initialized")
	}
}

func TestLoad(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	// The maximum size of a program is the total memory size, minus the space
	// reserved for the virtual machine.

	var program [4096 - 512]uint8

	for i := range program {
		program[i] = 0xff
	}

	e.Load(program[:])

	if !slices.Equal(e.Memory()[0x200:], program[:]) {
		t.Fatalf("memory not loaded")
	}
}
func TestConstLoad(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0xff, // LD V0, 0xff
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0xff)
}

func TestConstIncrement(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x70, 0x01, // ADD V0, 0x00
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0x01)
}

func TestAssign(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0xff, // LD V0, 0xFF
		0x81, 0x00, // LD V1, V0
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0xff)
	checkRegister(t, e.V(), 0x1, 0xff)
}

func TestBitwiseOr(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0xf0, // LD V0, 0xf0
		0x61, 0x0f, // LD V1, 0x0f
		0x81, 0x01, // OR V1, V0
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0xf0)
	checkRegister(t, e.V(), 0x1, 0xff)
}

func TestBitwiseAnd(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0xf0, // LD V0, 0xf0
		0x61, 0x0f, // LD V1, 0x0f
		0x81, 0x02, // AND V1, V0
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0xf0)
	checkRegister(t, e.V(), 0x1, 0x00)
}

func TestBitwiseXor(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0xf0, // LD V0, 0xf0
		0x61, 0x0f, // LD V1, 0x0f
		0x81, 0x03, // XOR V1, V0
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0xf0)
	checkRegister(t, e.V(), 0x1, 0xff)
}

func TestAdd(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0x01, // LD V0, 0x01
		0x61, 0xfe, // LD V1, 0xfe
		0x81, 0x04, // ADD V1, V0
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0x01)
	checkRegister(t, e.V(), 0x1, 0xff)
	checkRegister(t, e.V(), 0xf, 0x00)
}

func TestAddOverflow(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0x0f, // LD V0, 0x0f
		0x61, 0xff, // LD V1, 0xff
		0x81, 0x04, // ADD V1, V0
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0x0f)
	checkRegister(t, e.V(), 0x1, 0x0e)
	checkRegister(t, e.V(), 0xf, 0x01)
}

func TestAddOverflowClear(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0x01, // LD V0, 0x01
		0x61, 0xff, // LD V1, 0xff
		0x81, 0x04, // ADD V1, V0
		0x81, 0x04, // ADD V1, V0
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0x01)
	checkRegister(t, e.V(), 0x1, 0x01)
	checkRegister(t, e.V(), 0xf, 0x00)
}

func TestSub(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0x01, // LD V0, 0x01
		0x61, 0xff, // LD V1, 0xff
		0x81, 0x05, // SUB V1, V0
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0x01)
	checkRegister(t, e.V(), 0x1, 0xfe)
	checkRegister(t, e.V(), 0xf, 0x01)
}

func TestSubUnderflow(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0x02, // LD V0, 0x02
		0x61, 0x01, // LD V1, 0x01
		0x81, 0x05, // SUB V1, V0
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0x02)
	checkRegister(t, e.V(), 0x1, 0xff)
	checkRegister(t, e.V(), 0xf, 0x00)
}

func TestSubUnderflowClear(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0x01, // LD V0, 0x01
		0x61, 0x00, // LD V1, 0x00
		0x81, 0x05, // SUB V1, V0
		0x81, 0x05, // SUB V1, V0
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0x01)
	checkRegister(t, e.V(), 0x1, 0xfe)
	checkRegister(t, e.V(), 0xf, 0x01)
}

func TestShiftRight(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0x02, // LD V0, 0x02
		0x80, 0x06, // SHR V0
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0x01)
	checkRegister(t, e.V(), 0xf, 0x00)
}

func TestShiftRightCarry(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0x03, // LD V0, 0x03
		0x80, 0x06, // SHR V0
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0x01)
	checkRegister(t, e.V(), 0xf, 0x01)
}

func TestSubn(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0x03, // LD V0, 0x03
		0x61, 0x01, // LD V1, 0x01
		0x81, 0x07, // SUBN V1, V0
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0x03)
	checkRegister(t, e.V(), 0x1, 0x02)
	checkRegister(t, e.V(), 0xf, 0x01)
}

func TestSubnUnderflow(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0x01, // LD V0, 0x01
		0x61, 0x03, // LD V1, 0x03
		0x81, 0x07, // SUBN V1, V0
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0x01)
	checkRegister(t, e.V(), 0x1, 0xfe)
	checkRegister(t, e.V(), 0xf, 0x00)
}

func TestShiftLeft(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0x40, // LD V0, 0x40
		0x80, 0x0e, // SHL V0
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0x80)
	checkRegister(t, e.V(), 0xf, 0x00)
}

func TestShiftLeftCarry(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0xc0, // LD V0, 0xc0
		0x80, 0x0e, // SHL V0
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0x80)
	checkRegister(t, e.V(), 0xf, 0x01)
}

func TestSkipIfEqual(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x30, 0x00, // SE V0, 0
		0x00, 0x00, // HALT
		0x60, 0x01, // LD V0, 0x01
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0x01)
}

func TestSkipIfNotEqual(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x40, 0x01, // SNE V0, 0x01
		0x00, 0x00, // HALT
		0x60, 0x01, // LD V0, 0x01
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0x01)
}

func TestSkipIfEqualRegister(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0x01, // LD V0, 0x01
		0x61, 0x01, // LD V1, 0x01
		0x50, 0x10, // SE V0, V1
		0x00, 0x00, // HALT
		0x62, 0x02, // LD V2, 0x01
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0x01)
	checkRegister(t, e.V(), 0x1, 0x01)
	checkRegister(t, e.V(), 0x2, 0x02)
}

func TestSkipIfNotEqualRegister(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0x01, // LD V0, 0x01
		0x61, 0x02, // LD V1, 0x02
		0x90, 0x10, // SNE V0, V1
		0x00, 0x00, // HALT
		0x62, 0x03, // LD V2, 0x03
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0x01)
	checkRegister(t, e.V(), 0x1, 0x02)
	checkRegister(t, e.V(), 0x2, 0x03)
}

func TestJump(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x12, 0x04, // JP ...
		0x60, 0x01, // LD V0, 0x01
		0x61, 0x01, // LD V1, 0x01
	})

	e.Run()

	checkRegister(t, e.V(), 0x1, 0x01)
}

func TestCallAndReturn(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x22, 0x06, // CALL 0x206
		0x61, 0x01, // LD V1, 0x01
		0x00, 0x00, // HALT
		0x60, 0x01, // LD V0, 0x01
		0x00, 0xee, // RET
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0x01)
	checkRegister(t, e.V(), 0x1, 0x01)
}

func TestJumpRelative(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0x04, // LD V0, 0x04
		0xb2, 0x02, // JP V0, 0x202
		0x00, 0x00, // HALT
		0x61, 0x01, // LD V1, 0x01
	})

	e.Run()

	checkRegister(t, e.V(), 0x0, 0x04)
	checkRegister(t, e.V(), 0x1, 0x01)
}

func checkRegister(t *testing.T, registers []uint8, i int, want uint8) {
	t.Helper()

	if i >= len(registers) {
		t.Fatalf("invalid register index: %x", i)
	}

	if got := registers[i]; got != want {
		t.Fatalf("V%X: got %#x, want %#x", i, got, want)
	}
}
