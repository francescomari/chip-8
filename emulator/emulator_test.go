package emulator_test

import (
	"testing"
	"time"

	"github.com/francescomari/chip-8/emulator"
)

func TestNew(t *testing.T) {
	e := emulator.New()

	var state emulator.State

	e.State(&state)

	// Memory location lower than 0x0200 are reserved to the interpreter and are
	// not guaranteed to be zeroed.

	for _, v := range state.Memory[0x200:] {
		if v != 0 {
			t.Fatalf("memory not zeroed")
		}
	}

	for i, v := range state.V {
		if v != 0 {
			t.Fatalf("register %d not zeroed", i)
		}
	}

	if state.I != 0 {
		t.Fatalf("index register not zeroed")
	}

	for _, v := range state.Stack {
		if v != 0 {
			t.Fatalf("stack not zeroed")
		}
	}

	if state.SP != 0 {
		t.Fatalf("stack pointer not zeroed")
	}

	if state.DT != 0 {
		t.Fatalf("delay timer not zeroed")
	}

	if state.ST != 0 {
		t.Fatalf("sound timer not zeroed")
	}

	if state.PC != 0x200 {
		t.Fatalf("program counter not initialized")
	}

	for y := range state.Display {
		for x := range state.Display[y] {
			if state.Display[y][x] != 0 {
				t.Fatalf("display not initialized")
			}
		}
	}
}

func TestConstLoad(t *testing.T) {
	e := run(t,
		0x60, 0xff, // LD V0, 0xff
	)

	check(t, e).
		register(0x0, 0xff)
}

func TestConstIncrement(t *testing.T) {
	e := run(t,
		0x70, 0x01, // ADD V0, 0x00
	)

	check(t, e).
		register(0x0, 0x01)
}

func TestAssign(t *testing.T) {
	e := run(t,
		0x60, 0xff, // LD V0, 0xFF
		0x81, 0x00, // LD V1, V0
	)

	check(t, e).
		register(0x0, 0xff).
		register(0x1, 0xff)
}

func TestBitwiseOr(t *testing.T) {
	e := run(t,
		0x60, 0xf0, // LD V0, 0xf0
		0x61, 0x0f, // LD V1, 0x0f
		0x81, 0x01, // OR V1, V0
	)

	check(t, e).
		register(0x0, 0xf0).
		register(0x1, 0xff)
}

func TestBitwiseOrFlagQuirk(t *testing.T) {
	e := run(t,
		0x60, 0xf0, // LD V0, 0xf0
		0x61, 0x0f, // LD V1, 0x0f
		0x6f, 0xff, // LD VF, 0xff
		0x81, 0x01, // OR V1, V0
	)

	check(t, e).
		register(0x0, 0xf0).
		register(0x1, 0xff).
		register(0xf, 0x00)
}

func TestBitwiseAnd(t *testing.T) {
	e := run(t,
		0x60, 0xf0, // LD V0, 0xf0
		0x61, 0x0f, // LD V1, 0x0f
		0x81, 0x02, // AND V1, V0
	)

	check(t, e).
		register(0x0, 0xf0).
		register(0x1, 0x00)
}

func TestBitwiseAndFlagQuirk(t *testing.T) {
	e := run(t,
		0x60, 0xf0, // LD V0, 0xf0
		0x61, 0x0f, // LD V1, 0x0f
		0x6f, 0xff, // LD VF, 0xff
		0x81, 0x02, // AND V1, V0
	)

	check(t, e).
		register(0x0, 0xf0).
		register(0x1, 0x00).
		register(0xf, 0x00)
}

func TestBitwiseXor(t *testing.T) {
	e := run(t,
		0x60, 0xf0, // LD V0, 0xf0
		0x61, 0x0f, // LD V1, 0x0f
		0x81, 0x03, // XOR V1, V0
	)

	check(t, e).
		register(0x0, 0xf0).
		register(0x1, 0xff)
}

func TestBitwiseXorFlagQuirk(t *testing.T) {
	e := run(t,
		0x60, 0xf0, // LD V0, 0xf0
		0x61, 0x0f, // LD V1, 0x0f
		0x6f, 0xff, // LD VF, 0xff
		0x81, 0x03, // XOR V1, V0
	)

	check(t, e).
		register(0x0, 0xf0).
		register(0x1, 0xff).
		register(0xf, 0x00)
}

func TestAdd(t *testing.T) {
	e := run(t,
		0x60, 0x01, // LD V0, 0x01
		0x61, 0xfe, // LD V1, 0xfe
		0x81, 0x04, // ADD V1, V0
	)

	check(t, e).
		register(0x0, 0x01).
		register(0x1, 0xff).
		register(0xf, 0x00)
}

func TestAddOverflow(t *testing.T) {
	e := run(t,
		0x60, 0x0f, // LD V0, 0x0f
		0x61, 0xff, // LD V1, 0xff
		0x81, 0x04, // ADD V1, V0
	)

	check(t, e).
		register(0x0, 0x0f).
		register(0x1, 0x0e).
		register(0xf, 0x01)
}

func TestAddOverflowClear(t *testing.T) {
	e := run(t,
		0x60, 0x01, // LD V0, 0x01
		0x61, 0xff, // LD V1, 0xff
		0x81, 0x04, // ADD V1, V0
		0x81, 0x04, // ADD V1, V0
	)

	check(t, e).
		register(0x0, 0x01).
		register(0x1, 0x01).
		register(0xf, 0x00)
}

func TestAddOverflowRegister(t *testing.T) {
	// VF should be set after the result of the addition is saved in the X
	// register. Otherwise, when Y is VF, VF is modified before the additional
	// takes places.

	e := run(t,
		0x6f, 0x01, // LD VF, 0x01
		0x80, 0xf4, // ADD V0, VF
	)

	check(t, e).
		register(0x0, 0x01).
		register(0xf, 0x00)
}

func TestSub(t *testing.T) {
	e := run(t,
		0x60, 0x01, // LD V0, 0x01
		0x61, 0xff, // LD V1, 0xff
		0x81, 0x05, // SUB V1, V0
	)

	check(t, e).
		register(0x0, 0x01).
		register(0x1, 0xfe).
		register(0xf, 0x01)
}

func TestSubUnderflow(t *testing.T) {
	e := run(t,
		0x60, 0x02, // LD V0, 0x02
		0x61, 0x01, // LD V1, 0x01
		0x81, 0x05, // SUB V1, V0
	)

	check(t, e).
		register(0x0, 0x02).
		register(0x1, 0xff).
		register(0xf, 0x00)
}

func TestSubUnderflowClear(t *testing.T) {
	e := run(t,
		0x60, 0x01, // LD V0, 0x01
		0x61, 0x00, // LD V1, 0x00
		0x81, 0x05, // SUB V1, V0
		0x81, 0x05, // SUB V1, V0
	)

	check(t, e).
		register(0x0, 0x01).
		register(0x1, 0xfe).
		register(0xf, 0x01)
}

func TestSubUnderflowRegister(t *testing.T) {
	e := run(t,
		0x60, 0x05, // LD V0, 0x05
		0x6f, 0x02, // LD VF, 0x02
		0x80, 0xf5, // SUB V0, VF
	)

	check(t, e).
		register(0x0, 0x3).
		register(0xf, 0x1)
}

func TestShiftRight(t *testing.T) {
	e := run(t,
		0x60, 0x02, // LD V0, 0x02
		0x80, 0x06, // SHR V0
	)

	check(t, e).
		register(0x0, 0x01).
		register(0xf, 0x00)
}

func TestShiftRightCarry(t *testing.T) {
	e := run(t,
		0x60, 0x03, // LD V0, 0x03
		0x80, 0x06, // SHR V0
	)

	check(t, e).
		register(0x0, 0x01).
		register(0xf, 0x01)
}

func TestShiftRightCarryRegister(t *testing.T) {
	e := run(t,
		0x60, 0x07, // LD V0, 0x07
		0x8f, 0x06, // SHR V0
	)

	check(t, e).
		register(0xf, 0x01)
}

func TestSubn(t *testing.T) {
	e := run(t,
		0x60, 0x03, // LD V0, 0x03
		0x61, 0x01, // LD V1, 0x01
		0x81, 0x07, // SUBN V1, V0
	)

	check(t, e).
		register(0x0, 0x03).
		register(0x1, 0x02).
		register(0xf, 0x01)
}

func TestSubnUnderflow(t *testing.T) {
	e := run(t,
		0x60, 0x01, // LD V0, 0x01
		0x61, 0x03, // LD V1, 0x03
		0x81, 0x07, // SUBN V1, V0
	)

	check(t, e).
		register(0x0, 0x01).
		register(0x1, 0xfe).
		register(0xf, 0x00)
}

func TestSubnUnderflowRegister(t *testing.T) {
	e := run(t,
		0x60, 0x02, // LD V0, 0x02
		0x6f, 0x05, // LD VF, 0x05
		0x80, 0xf7, // SUBN V0, VF
	)

	check(t, e).
		register(0x0, 0x03).
		register(0xf, 0x01)
}

func TestShiftLeft(t *testing.T) {
	e := run(t,
		0x60, 0x40, // LD V0, 0x40
		0x80, 0x0e, // SHL V0
	)

	check(t, e).
		register(0x0, 0x80).
		register(0xf, 0x00)
}

func TestShiftLeftCarry(t *testing.T) {
	e := run(t,
		0x60, 0xc0, // LD V0, 0xc0
		0x80, 0x0e, // SHL V0
	)

	check(t, e).
		register(0x0, 0x80).
		register(0xf, 0x01)
}

func TestShiftLeftCarryRegister(t *testing.T) {
	e := run(t,
		0x60, 0xf0, // LD V0, 0xc0
		0x8f, 0x0e, // SHL V0
	)

	check(t, e).
		register(0xf, 0x01)
}

func TestSkipIfEqual(t *testing.T) {
	e := run(t,
		0x30, 0x00, // SE V0, 0
		0x00, 0x00, // HALT
		0x60, 0x01, // LD V0, 0x01
	)

	check(t, e).
		register(0x0, 0x01)
}

func TestSkipIfNotEqual(t *testing.T) {
	e := run(t,
		0x40, 0x01, // SNE V0, 0x01
		0x00, 0x00, // HALT
		0x60, 0x01, // LD V0, 0x01
	)

	check(t, e).
		register(0x0, 0x01)
}

func TestSkipIfEqualRegister(t *testing.T) {
	e := run(t,
		0x60, 0x01, // LD V0, 0x01
		0x61, 0x01, // LD V1, 0x01
		0x50, 0x10, // SE V0, V1
		0x00, 0x00, // HALT
		0x62, 0x02, // LD V2, 0x01
	)

	check(t, e).
		register(0x0, 0x01).
		register(0x1, 0x01).
		register(0x2, 0x02)
}

func TestSkipIfNotEqualRegister(t *testing.T) {
	e := run(t,
		0x60, 0x01, // LD V0, 0x01
		0x61, 0x02, // LD V1, 0x02
		0x90, 0x10, // SNE V0, V1
		0x00, 0x00, // HALT
		0x62, 0x03, // LD V2, 0x03
	)

	check(t, e).
		register(0x0, 0x01).
		register(0x1, 0x02).
		register(0x2, 0x03)
}

func TestJump(t *testing.T) {
	e := run(t,
		0x12, 0x04, // JP 0x204
		0x60, 0x01, // LD V0, 0x01
		0x61, 0x01, // LD V1, 0x01
	)

	check(t, e).
		register(0x1, 0x01)
}

func TestCallAndReturn(t *testing.T) {
	e := run(t,
		0x22, 0x06, // CALL 0x206
		0x61, 0x01, // LD V1, 0x01
		0x00, 0x00, // HALT
		0x60, 0x01, // LD V0, 0x01
		0x00, 0xee, // RET
	)

	check(t, e).
		register(0x0, 0x01).
		register(0x1, 0x01)
}

func TestJumpRelative(t *testing.T) {
	e := run(t,
		0x60, 0x04, // LD V0, 0x04
		0xb2, 0x02, // JP V0, 0x202
		0x00, 0x00, // HALT
		0x61, 0x01, // LD V1, 0x01
	)

	check(t, e).
		register(0x0, 0x04).
		register(0x0, 0x04)
}

func TestLoadIndex(t *testing.T) {
	e := run(t,
		0xa2, 0xff, // LD I, 0x2ff
	)

	check(t, e).
		index(0x2ff)
}

func TestAddIndex(t *testing.T) {
	e := run(t,
		0x60, 0x01, // LD V0, 0x01
		0xa0, 0x02, // LD I, 0x02
		0xf0, 0x1e, // ADD I, V0
	)

	check(t, e).
		register(0x0, 0x01).
		index(0x03)
}

func TestStoreMemory(t *testing.T) {
	e := run(t,
		0x60, 0x01, // LD V0, 0x01
		0x61, 0x02, // LD V1, 0x01
		0xa3, 0x00, // LD I, 0x300
		0xf1, 0x55, // LD [I], V1
	)

	check(t, e).
		register(0x0, 0x01).
		register(0x1, 0x02).
		index(0x0302).
		memory(0x0300, 0x01).
		memory(0x0301, 0x02)
}

func TestLoadMemory(t *testing.T) {
	e := run(t,
		0x60, 0x01, // LD V0, 0x01
		0x61, 0x02, // LD V1, 0x01
		0xa3, 0x00, // LD I, 0x300
		0xf1, 0x55, // LD [I], V1
		0x80, 0x03, // XOR V0, V0
		0x81, 0x13, // XOR V0, V0
		0xa3, 0x00, // LD I, 0x300
		0xf1, 0x65, // LD V1, [I]
	)

	check(t, e).
		register(0x0, 0x01).
		register(0x1, 0x02).
		index(0x0302).
		memory(0x0300, 0x01).
		memory(0x0301, 0x02)
}

func TestStoreBCD(t *testing.T) {
	e := run(t,
		0x60, 0xfe, // LD V0, 0xfe
		0xa3, 0x00, // LD I, 0x300
		0xf0, 0x33, // LD B, V0
	)

	check(t, e).
		register(0x0, 0xfe).
		index(0x300).
		memory(0x300, 2).
		memory(0x301, 5).
		memory(0x302, 4)
}

func TestDraw(t *testing.T) {
	e := run(t,
		0x60, 0x01, // LD V0, 0x01
		0x61, 0x02, // LD V1, 0x02
		0xa2, 0x0a, // LD I, 0x20a
		0xd0, 0x12, // DRW V0, V1, 0x02
		0x00, 0x00, // HALT
		0x80, // Bitmap, *.......
		0x01, // Bitmap, .......*
	)

	check(t, e).
		register(0x0, 0x01).
		register(0x01, 0x02).
		register(0x0f, 0x00).
		index(0x20a).
		display(1, 2, true).
		display(8, 3, true)
}

func TestDrawCollision(t *testing.T) {
	e := run(t,
		0x60, 0x01, // LD V0, 0x01
		0x61, 0x02, // LD V1, 0x02
		0xa2, 0x0c, // LD I, 0x20c
		0xd0, 0x12, // DRW V0, V1, 0x02
		0xd0, 0x11, // DRW V0, V1, 0x01
		0x00, 0x00, // HALT
		0x80, // Bitmap, *.......
		0x01, // Bitmap, .......*
	)

	check(t, e).
		register(0x0, 0x01).
		register(0x01, 0x02).
		register(0x0f, 0x01).
		index(0x20c).
		display(1, 2, false).
		display(8, 3, true)
}

func TestClearDisplay(t *testing.T) {
	e := run(t,
		0x60, 0x01, // LD V0, 0x01
		0x61, 0x02, // LD V1, 0x02
		0xa2, 0x0a, // LD I, 0x20c
		0xd0, 0x12, // DRW V0, V1, 0x01
		0x00, 0xe0, // CLS
		0x00, 0x00, // HALT
		0x80, // Bitmap, *.......
	)

	check(t, e).
		register(0x0, 0x01).
		register(0x01, 0x02).
		register(0x0f, 0x00).
		index(0x20a).
		display(1, 2, false)
}

func TestCharacterAddress(t *testing.T) {
	e := run(t,
		0x60, 0x0f, // LD V0, 0x0f
		0xf0, 0x29, // LD F, V0
		0xd1, 0x15, // DRW V1, V1, 0x05
	)

	check(t, e).
		register(0x0, 0x0f).
		register(0xf, 0).
		display(0, 0, true).
		display(1, 0, true).
		display(2, 0, true).
		display(3, 0, true).
		display(0, 1, true).
		display(0, 2, true).
		display(1, 2, true).
		display(2, 2, true).
		display(3, 2, true).
		display(0, 3, true).
		display(0, 4, true)
}

func TestSoundTimer(t *testing.T) {
	e := emulator.New()

	var sound bool

	e.SetSound(func() {
		sound = true
	})

	e.Load([]uint8{
		0x60, 0x0f, // LD V0, 0x0f
		0xf0, 0x18, // LD ST, V0
		0xf0, 0x15, // LD DT, V0
		0xf1, 0x07, // LD V1, DT
		0x31, 0x00, // SE V1, 0x00
		0x12, 0x06, // JP 0x206
	})

	clock := time.Tick(time.Second / 60)

	for {
		ok, err := e.Step()
		if err != nil {
			t.Fatalf("step: %v", err)
		}
		if !ok {
			break
		}
		select {
		case <-clock:
			e.Clock()
		default:
			continue
		}
	}

	if !sound {
		t.Fatalf("sound timer not triggered")
	}
}

func TestDelayTimer(t *testing.T) {
	e := run(t,
		0x60, 0x0f, // LD V0, 0x0f
		0xf0, 0x15, // LD DT, V0
		0xf1, 0x07, // LD V1, DT
		0x31, 0x00, // SE V1, 0x00
		0x12, 0x04, // JP 0x204
	)

	check(t, e).
		register(0x0, 0x0f).
		register(0x1, 0x00).
		delayTimer(0x00)
}

func TestSkipOnKeyDown(t *testing.T) {
	e := emulator.New()

	e.KeyDown(0xf)

	e.Load([]uint8{
		0x60, 0x0f, // LD V0, 0x0f
		0xe0, 0x9e, // SKP V0
		0x61, 0x01, // LD V1, 0x01
		0x60, 0x0e, // LD V0, 0x0e
		0xe0, 0x9e, // SKP V0
		0x62, 0x01, // LD V2, 0x02
	})

	for {
		ok, err := e.Step()
		if err != nil {
			t.Fatalf("step: %v", err)
		}
		if !ok {
			break
		}
	}

	check(t, e).
		register(0x1, 0x00).
		register(0x2, 0x01)
}

func TestSkipOnKeyNotDown(t *testing.T) {
	e := emulator.New()

	e.KeyDown(0xf)

	e.Load([]uint8{
		0x60, 0x0f, // LD V0, 0x0f
		0xe0, 0xa1, // SKPN V0
		0x61, 0x01, // LD V1, 0x01
		0x60, 0x0e, // LD V0, 0x0e
		0xe0, 0xa1, // SKPN V0
		0x62, 0x01, // LD V2, 0x02
	})

	for {
		ok, err := e.Step()
		if err != nil {
			t.Fatalf("step: %v", err)
		}
		if !ok {
			break
		}
	}

	check(t, e).
		register(0x1, 0x01).
		register(0x2, 0x00)
}

func TestSkipWithoutKeyDown(t *testing.T) {
	e := run(t,
		0x60, 0x0f, // LD V0, 0x0f
		0xe0, 0x9e, // SKP V0
		0x61, 0x01, // LD V1, 0x01
		0xe0, 0xa1, // SKPN V0
		0x62, 0x01, // LD V2, 0x01
	)

	check(t, e).
		register(0x1, 0x01).
		register(0x2, 0x00)
}

func TestWaitKeyPress(t *testing.T) {
	e := emulator.New()

	e.Load([]uint8{
		0xf0, 0x0a, // LD V0, K
	})

	ok, err := e.Step()
	if err != nil {
		t.Fatalf("step: %v", err)
	}
	if !ok {
		t.Fatal("should continue")
	}

	check(t, e).
		register(0x0, 0x00)

	e.KeyDown(0xf)
	e.KeyUp(0xf)
	e.KeyDown(0x1)
	e.KeyUp(0x1)

	check(t, e).
		register(0x0, 0x0f)

	ok, err = e.Step()
	if err != nil {
		t.Fatalf("step: %v", err)
	}
	if ok {
		t.Fatal("should stop")
	}
}

func TestRandom(t *testing.T) {
	e := emulator.New()

	e.SetRNG(func() uint32 {
		return 0x88
	})

	e.Load([]uint8{
		0xc0, 0x0f, // RND V0, 0xff
	})

	for {
		ok, err := e.Step()
		if err != nil {
			t.Fatalf("step: %v", err)
		}
		if !ok {
			break
		}
	}

	check(t, e).
		register(0x0, 0x08)
}

func TestStepInvalidOpcode(t *testing.T) {
	e := emulator.New()

	e.Load([]uint8{
		0x80, 0x09, // Invalid ALU opcode
	})

	if _, err := e.Step(); err == nil {
		t.Fatal("expected error for invalid opcode")
	}
}

func check(t *testing.T, e *emulator.Emulator) checks {
	return checks{t: t, e: e}
}

type checks struct {
	t *testing.T
	e *emulator.Emulator
}

func (c checks) register(i int, want uint8) checks {
	c.t.Helper()

	var state emulator.State

	c.e.State(&state)

	if got := state.V[i]; got != want {
		c.t.Fatalf("V%X: got %#x, want %#x", i, got, want)
	}

	return c
}

func (c checks) index(want uint16) checks {
	c.t.Helper()

	var state emulator.State

	c.e.State(&state)

	if got := state.I; got != want {
		c.t.Fatalf("I: got %#x, want %#x", got, want)
	}

	return c
}

func (c checks) memory(address uint16, want uint8) checks {
	c.t.Helper()

	var state emulator.State

	c.e.State(&state)

	if got := state.Memory[address]; got != want {
		c.t.Fatalf("memory[%x]: got %#x, want %#x", address, got, want)
	}

	return c
}

func (c checks) display(x, y int, on bool) checks {
	c.t.Helper()

	var state emulator.State

	c.e.State(&state)

	if got := state.Display[y][x]; on && got == 0 {
		c.t.Fatalf("display[%d,%d]: pixel should be on, but it is off", x, y)
	} else if !on && got != 0 {
		c.t.Fatalf("display[%d,%d]: pixel should be off, but it is on", x, y)
	}

	return c
}

func (c checks) delayTimer(want uint8) checks {
	var state emulator.State

	c.e.State(&state)

	if got := state.DT; got != want {
		c.t.Fatalf("delay timer: got %#x, want %#x", got, want)
	}

	return c
}

func run(t *testing.T, data ...uint8) *emulator.Emulator {
	t.Helper()

	e := emulator.New()

	e.Load(data)

	clock := time.Tick(time.Second / 60)

	for {
		ok, err := e.Step()
		if err != nil {
			t.Fatalf("step: %v", err)
		}
		if !ok {
			break
		}
		select {
		case <-clock:
			e.Clock()
		default:
			continue
		}
	}

	return e
}
