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
		0x60, 0xff, // LD V0, 0xFF
		0x61, 0xff, // LD V1, 0xFF
		0x62, 0xff, // LD V2, 0xFF
		0x63, 0xff, // LD V3, 0xFF
		0x64, 0xff, // LD V4, 0xFF
		0x65, 0xff, // LD V5, 0xFF
		0x66, 0xff, // LD V6, 0xFF
		0x67, 0xff, // LD V7, 0xFF
		0x68, 0xff, // LD V8, 0xFF
		0x69, 0xff, // LD V9, 0xFF
		0x6a, 0xff, // LD VA, 0xFF
		0x6b, 0xff, // LD VB, 0xFF
		0x6c, 0xff, // LD VC, 0xFF
		0x6d, 0xff, // LD VD, 0xFF
		0x6e, 0xff, // LD VE, 0xFF
		0x6f, 0xff, // LD VF, 0xFF
		0x00, 0x00, // HALT
	})

	e.Run()

	v := []uint8{
		0xff, // V0
		0xff, // V1
		0xff, // V2
		0xff, // V3
		0xff, // V4
		0xff, // V5
		0xff, // V6
		0xff, // V7
		0xff, // V8
		0xff, // V9
		0xff, // VA
		0xff, // VB
		0xff, // VC
		0xff, // VD
		0xff, // VE
		0xff, // VF
	}

	if !slices.Equal(e.V(), v) {
		t.Fatalf("invalid registers: %v", e.V())
	}
}

func TestConstIncrement(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x70, 0x00, // ADD V0, 0x00
		0x71, 0x01, // ADD V1, 0x01
		0x72, 0x02, // ADD V2, 0x02
		0x73, 0x03, // ADD V3, 0x03
		0x74, 0x04, // ADD V4, 0x04
		0x75, 0x05, // ADD V5, 0x05
		0x76, 0x06, // ADD V6, 0x06
		0x77, 0x07, // ADD V7, 0x07
		0x78, 0x08, // ADD V8, 0x08
		0x79, 0x09, // ADD V9, 0x09
		0x7a, 0x0a, // ADD VA, 0x0A
		0x7b, 0x0b, // ADD VB, 0x0B
		0x7c, 0x0c, // ADD VC, 0x0C
		0x7d, 0x0d, // ADD VD, 0x0D
		0x7e, 0x0e, // ADD VE, 0x0E
		0x7f, 0x0f, // ADD VF, 0x0F
		0x00, 0x00, // HALT
	})

	e.Run()

	checkRegisters(t, e.V(), []uint8{
		0x00, // V0
		0x01, // V1
		0x02, // V2
		0x03, // V3
		0x04, // V4
		0x05, // V5
		0x06, // V6
		0x07, // V7
		0x08, // V8
		0x09, // V9
		0x0a, // VA
		0x0b, // VB
		0x0c, // VC
		0x0d, // VD
		0x0e, // VE
		0x0f, // VF
	})
}

func TestAssign(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0xff, // LD V0, 0xFF
		0x81, 0x00, // LD V1, V0
		0x82, 0x10, // LD V2, V1
		0x83, 0x20, // LD V3, V2
		0x84, 0x30, // LD V4, V3
		0x85, 0x40, // LD V5, V4
		0x86, 0x50, // LD V6, V5
		0x87, 0x60, // LD V7, V6
		0x88, 0x70, // LD V8, V7
		0x89, 0x80, // LD V9, V8
		0x8a, 0x90, // LD VA, V9
		0x8b, 0xa0, // LD VB, VA
		0x8c, 0xb0, // LD VC, VB
		0x8d, 0xc0, // LD VD, VC
		0x8e, 0xd0, // LD VE, VD
		0x8f, 0xe0, // LD VF, VE
		0x80, 0xf0, // LD V0, VF
		0x00, 0x00, // HALT
	})

	e.Run()

	checkRegisters(t, e.V(), []uint8{
		0xff, // V0
		0xff, // V1
		0xff, // V2
		0xff, // V3
		0xff, // V4
		0xff, // V5
		0xff, // V6
		0xff, // V7
		0xff, // V8
		0xff, // V9
		0xff, // VA
		0xff, // VB
		0xff, // VC
		0xff, // VD
		0xff, // VE
		0xff, // VF
	})
}

func TestBitwiseOr(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0xf0, // LD V0, 0xf0
		0x61, 0x0f, // LD V1, 0x0f
		0x62, 0x0f, // LD V2, 0x0f
		0x63, 0x0f, // LD V3, 0x0f
		0x64, 0x0f, // LD V4, 0x0f
		0x65, 0x0f, // LD V5, 0x0f
		0x66, 0x0f, // LD V6, 0x0f
		0x67, 0x0f, // LD V7, 0x0f
		0x68, 0x0f, // LD V8, 0x0f
		0x69, 0x0f, // LD V9, 0x0f
		0x6a, 0x0f, // LD VA, 0x0f
		0x6b, 0x0f, // LD VB, 0x0f
		0x6c, 0x0f, // LD VC, 0x0f
		0x6d, 0x0f, // LD VD, 0x0f
		0x6e, 0x0f, // LD VE, 0x0f
		0x6f, 0x0f, // LD VF, 0x0f
		0x81, 0x01, // OR V1, V0
		0x82, 0x01, // OR V2, V0
		0x83, 0x01, // OR V3, V0
		0x84, 0x01, // OR V4, V0
		0x85, 0x01, // OR V5, V0
		0x86, 0x01, // OR V6, V0
		0x87, 0x01, // OR V7, V0
		0x88, 0x01, // OR V8, V0
		0x89, 0x01, // OR V9, V0
		0x8a, 0x01, // OR VA, V0
		0x8b, 0x01, // OR VB, V0
		0x8c, 0x01, // OR VC, V0
		0x8d, 0x01, // OR VD, V0
		0x8e, 0x01, // OR VE, V0
		0x8f, 0x01, // OR VF, V0
		0x00, 0x00, // HALT
	})

	e.Run()

	checkRegisters(t, e.V(), []uint8{
		0xf0, // V0
		0xff, // V1
		0xff, // V2
		0xff, // V3
		0xff, // V4
		0xff, // V5
		0xff, // V6
		0xff, // V7
		0xff, // V8
		0xff, // V9
		0xff, // VA
		0xff, // VB
		0xff, // VC
		0xff, // VD
		0xff, // VE
		0xff, // VF
	})
}

func TestBitwiseAnd(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0xf0, // LD V0, 0xf0
		0x61, 0x0f, // LD V1, 0x0f
		0x62, 0x0f, // LD V2, 0x0f
		0x63, 0x0f, // LD V3, 0x0f
		0x64, 0x0f, // LD V4, 0x0f
		0x65, 0x0f, // LD V5, 0x0f
		0x66, 0x0f, // LD V6, 0x0f
		0x67, 0x0f, // LD V7, 0x0f
		0x68, 0x0f, // LD V8, 0x0f
		0x69, 0x0f, // LD V9, 0x0f
		0x6a, 0x0f, // LD VA, 0x0f
		0x6b, 0x0f, // LD VB, 0x0f
		0x6c, 0x0f, // LD VC, 0x0f
		0x6d, 0x0f, // LD VD, 0x0f
		0x6e, 0x0f, // LD VE, 0x0f
		0x6f, 0x0f, // LD VF, 0x0f
		0x81, 0x02, // AND V1, V0
		0x82, 0x02, // AND V2, V0
		0x83, 0x02, // AND V3, V0
		0x84, 0x02, // AND V4, V0
		0x85, 0x02, // AND V5, V0
		0x86, 0x02, // AND V6, V0
		0x87, 0x02, // AND V7, V0
		0x88, 0x02, // AND V8, V0
		0x89, 0x02, // AND V9, V0
		0x8a, 0x02, // AND VA, V0
		0x8b, 0x02, // AND VB, V0
		0x8c, 0x02, // AND VC, V0
		0x8d, 0x02, // AND VD, V0
		0x8e, 0x02, // AND VE, V0
		0x8f, 0x02, // AND VF, V0
		0x00, 0x00, // HALT
	})

	e.Run()

	checkRegisters(t, e.V(), []uint8{
		0xf0, // V0
		0x00, // V1
		0x00, // V2
		0x00, // V3
		0x00, // V4
		0x00, // V5
		0x00, // V6
		0x00, // V7
		0x00, // V8
		0x00, // V9
		0x00, // VA
		0x00, // VB
		0x00, // VC
		0x00, // VD
		0x00, // VE
		0x00, // VF
	})
}

func TestBitwiseXor(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0xf0, // LD V0, 0xf0
		0x61, 0x0f, // LD V1, 0x0f
		0x62, 0x0f, // LD V2, 0x0f
		0x63, 0x0f, // LD V3, 0x0f
		0x64, 0x0f, // LD V4, 0x0f
		0x65, 0x0f, // LD V5, 0x0f
		0x66, 0x0f, // LD V6, 0x0f
		0x67, 0x0f, // LD V7, 0x0f
		0x68, 0x0f, // LD V8, 0x0f
		0x69, 0x0f, // LD V9, 0x0f
		0x6a, 0x0f, // LD VA, 0x0f
		0x6b, 0x0f, // LD VB, 0x0f
		0x6c, 0x0f, // LD VC, 0x0f
		0x6d, 0x0f, // LD VD, 0x0f
		0x6e, 0x0f, // LD VE, 0x0f
		0x6f, 0x0f, // LD VF, 0x0f
		0x81, 0x03, // XOR V1, V0
		0x82, 0x03, // XOR V2, V0
		0x83, 0x03, // XOR V3, V0
		0x84, 0x03, // XOR V4, V0
		0x85, 0x03, // XOR V5, V0
		0x86, 0x03, // XOR V6, V0
		0x87, 0x03, // XOR V7, V0
		0x88, 0x03, // XOR V8, V0
		0x89, 0x03, // XOR V9, V0
		0x8a, 0x03, // XOR VA, V0
		0x8b, 0x03, // XOR VB, V0
		0x8c, 0x03, // XOR VC, V0
		0x8d, 0x03, // XOR VD, V0
		0x8e, 0x03, // XOR VE, V0
		0x8f, 0x03, // XOR VF, V0
		0x00, 0x00, // HALT
	})

	e.Run()

	checkRegisters(t, e.V(), []uint8{
		0xf0, // V0
		0xff, // V1
		0xff, // V2
		0xff, // V3
		0xff, // V4
		0xff, // V5
		0xff, // V6
		0xff, // V7
		0xff, // V8
		0xff, // V9
		0xff, // VA
		0xff, // VB
		0xff, // VC
		0xff, // VD
		0xff, // VE
		0xff, // VF
	})
}
func TestAdd(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0x01, // LD V0, 0x01
		0x61, 0xfe, // LD V1, 0xfe
		0x81, 0x04, // ADD V1, V0
		0x00, 0x00, // HALT
	})

	e.Run()

	checkRegisters(t, e.V(), []uint8{
		0x01, // V0
		0xff, // V1
		0x00, // V2
		0x00, // V3
		0x00, // V4
		0x00, // V5
		0x00, // V6
		0x00, // V7
		0x00, // V8
		0x00, // V9
		0x00, // VA
		0x00, // VB
		0x00, // VC
		0x00, // VD
		0x00, // VE
		0x00, // VF
	})
}

func TestAddOverflow(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0x0f, // LD V0, 0x0f
		0x61, 0xff, // LD V1, 0xff
		0x81, 0x04, // ADD V1, V0
		0x00, 0x00, // HALT
	})

	e.Run()

	checkRegisters(t, e.V(), []uint8{
		0x0f, // V0
		0x0e, // V1
		0x00, // V2
		0x00, // V3
		0x00, // V4
		0x00, // V5
		0x00, // V6
		0x00, // V7
		0x00, // V8
		0x00, // V9
		0x00, // VA
		0x00, // VB
		0x00, // VC
		0x00, // VD
		0x00, // VE
		0x01, // VF
	})
}

func TestSub(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0x01, // LD V0, 0x01
		0x61, 0xff, // LD V1, 0xff
		0x81, 0x05, // SUB V1, V0
		0x00, 0x00, // HALT
	})

	e.Run()

	checkRegisters(t, e.V(), []uint8{
		0x01, // V0
		0xfe, // V1
		0x00, // V2
		0x00, // V3
		0x00, // V4
		0x00, // V5
		0x00, // V6
		0x00, // V7
		0x00, // V8
		0x00, // V9
		0x00, // VA
		0x00, // VB
		0x00, // VC
		0x00, // VD
		0x00, // VE
		0x01, // VF
	})
}

func TestSubUnderflow(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x60, 0x02, // LD V0, 0x02
		0x61, 0x01, // LD V1, 0x01
		0x81, 0x05, // SUB V1, V0
		0x00, 0x00, // HALT
	})

	e.Run()

	checkRegisters(t, e.V(), []uint8{
		0x02, // V0
		0xff, // V1
		0x00, // V2
		0x00, // V3
		0x00, // V4
		0x00, // V5
		0x00, // V6
		0x00, // V7
		0x00, // V8
		0x00, // V9
		0x00, // VA
		0x00, // VB
		0x00, // VC
		0x00, // VD
		0x00, // VE
		0x00, // VF
	})
}

func checkRegisters(t *testing.T, got, want []uint8) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("invalid number of registers provided")
	}

	for i := range got {
		if got[i] != want[i] {
			t.Errorf("V%X: got %#x, want %#x", i, got[i], want[i])
		}
	}
}
