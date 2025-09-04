package main

import (
	_ "embed"
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/francescomari/chip-8/emulator"
	"github.com/francescomari/chip-8/trace"
)

//go:embed beep.wav
var beep []byte

var mappings = map[ebiten.Key]uint8{
	ebiten.Key1: 0x1,
	ebiten.Key2: 0x2,
	ebiten.Key3: 0x3,
	ebiten.Key4: 0xc,
	ebiten.KeyQ: 0x4,
	ebiten.KeyW: 0x5,
	ebiten.KeyE: 0x6,
	ebiten.KeyR: 0xd,
	ebiten.KeyA: 0x7,
	ebiten.KeyS: 0x8,
	ebiten.KeyD: 0x9,
	ebiten.KeyF: 0xe,
	ebiten.KeyZ: 0xa,
	ebiten.KeyX: 0x0,
	ebiten.KeyC: 0xb,
	ebiten.KeyV: 0xf,
}

type Game struct {
	emulator *emulator.Emulator
	debug    bool
	state    emulator.State
	once     sync.Once
}

func (g *Game) Update() error {
	g.once.Do(func() {
		if g.debug {
			g.logDebugBanner()
		}
	})

	var keys [16]ebiten.Key

	for _, key := range inpututil.AppendJustPressedKeys(keys[:0]) {
		if value, ok := mappings[key]; ok {
			g.emulator.KeyDown(value)
		}
	}

	for _, key := range inpututil.AppendJustReleasedKeys(keys[:0]) {
		if value, ok := mappings[key]; ok {
			g.emulator.KeyUp(value)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF12) {
		g.debug = !g.debug

		if g.debug {
			g.logDebugBanner()
		}
	}

	if g.debug {
		var printState bool

		if inpututil.IsKeyJustPressed(ebiten.KeyF10) {
			g.emulator.Clock()
			printState = true
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
			g.emulator.Step()
			printState = true
		}

		if printState {
			g.logState()
		}
	} else {

		// Ebitengine calls this function (by default) every 1/60 seconds. This
		// frequency is determined by the Ticks Per Second (TPS) configuration
		// option. Ebitengine will adjust the time between calls to Update() so that
		// the code can assume a constant TPS and doesn't have to track the time
		// internally.

		g.emulator.Clock()

		// Experimentally, 530 Instructions Per Second (IPS) seems to be a good
		// speed to emulate CHIP-8 at. The number of instructions to run in a single
		// call to Update() has been determined by dividing the IPS by the TPS, and
		// truncating the result.

		for range 8 {
			g.emulator.Step()
		}
	}

	g.emulator.State(&g.state)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	// This uses the same color palette of the original Game Boy, as documented by
	// https://en.wikipedia.org/wiki/List_of_video_game_console_palettes.

	for y := range g.state.Display {
		for x := range g.state.Display[y] {
			if g.state.Display[y][x] != 0 {
				screen.Set(x, y, color.RGBA{R: 0x29, G: 0x41, B: 0x39, A: 0xff})
			} else {
				screen.Set(x, y, color.RGBA{R: 0x7b, G: 0x82, B: 0x10, A: 0xff})
			}
		}
	}
}

func (g *Game) Layout(_, _ int) (int, int) {
	return emulator.DisplayWidth, emulator.DisplayHeight
}

func (g *Game) logState() {
	var s strings.Builder

	trace.PrintInstruction(&s, &g.state)
	s.WriteString("\n")

	trace.PrintRegisters(&s, &g.state)
	s.WriteString("\n")

	trace.PrintState(&s, &g.state)
	s.WriteString("\n")

	log.Print(s.String())
}

func (g *Game) logDebugBanner() {
	log.Printf("Debug mode:")
	log.Printf(" [F10] Advance the clock")
	log.Printf(" [F11] Next instruction")
	log.Printf(" [F12] Exit debug mode")
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run() error {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "Start the emulator in debug mode")
	flag.Parse()

	if flag.NArg() != 1 {
		return fmt.Errorf("invalid number of arguments")
	}

	rom, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		return fmt.Errorf("read file: %v", err)
	}

	context := audio.NewContext(44100)

	e := emulator.New()

	e.Load(rom)

	e.SetSound(func() {
		context.NewPlayerFromBytes(beep).Play()
	})

	game := Game{
		emulator: e,
		debug:    debug,
	}

	ebiten.SetWindowSize(10*emulator.DisplayWidth, 10*emulator.DisplayHeight)
	ebiten.SetWindowTitle("CHIP-8 Emulator")

	if err := ebiten.RunGame(&game); err != nil {
		return fmt.Errorf("run game: %v", err)
	}

	return nil
}
