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
		program[i] = 0xFF
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
		0x60, 0xFF, // V0 = 0xFF,
		0x61, 0xFF, // V1 = 0xFF,
		0x62, 0xFF, // V2 = 0xFF,
		0x63, 0xFF, // V3 = 0xFF,
		0x64, 0xFF, // V4 = 0xFF,
		0x65, 0xFF, // V5 = 0xFF,
		0x66, 0xFF, // V6 = 0xFF,
		0x67, 0xFF, // V7 = 0xFF,
		0x68, 0xFF, // V8 = 0xFF,
		0x69, 0xFF, // V9 = 0xFF,
		0x6A, 0xFF, // VA = 0xFF,
		0x6B, 0xFF, // VB = 0xFF,
		0x6C, 0xFF, // VC = 0xFF,
		0x6D, 0xFF, // VD = 0xFF,
		0x6E, 0xFF, // VE = 0xFF,
		0x6F, 0xFF, // VF = 0xFF,
	})

	for i := 0; i < 16; i++ {
		e.Step()

		if e.V()[i] != 0xFF {
			t.Fatalf("register %d not loaded", i)
		}
	}
}

func TestConstIncrement(t *testing.T) {
	var e emulator.Emulator

	e.Init()

	e.Load([]uint8{
		0x70, 0x00, // V0 += 0x00,
		0x71, 0x01, // V1 += 0x01,
		0x72, 0x02, // V2 += 0x02,
		0x73, 0x03, // V3 += 0x03,
		0x74, 0x04, // V4 += 0x04,
		0x75, 0x05, // V5 += 0x05,
		0x76, 0x06, // V6 += 0x06,
		0x77, 0x07, // V7 += 0x07,
		0x78, 0x08, // V8 += 0x08,
		0x79, 0x09, // V9 += 0x09,
		0x7A, 0x0A, // VA += 0x0A,
		0x7B, 0x0B, // VB += 0x0B,
		0x7C, 0x0C, // VC += 0x0C,
		0x7D, 0x0D, // VD += 0x0D,
		0x7E, 0x0E, // VE += 0x0E,
		0x7F, 0x0F, // VF += 0x0F,
	})

	for i := 0; i < 16; i++ {
		e.Step()

		if e.V()[i] != uint8(i) {
			t.Fatalf("register %d not loaded", i)
		}
	}
}
