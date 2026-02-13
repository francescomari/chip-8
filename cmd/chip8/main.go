package main

import (
	_ "embed"
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/francescomari/chip-8/debug"
	"github.com/francescomari/chip-8/emulator"
)

const (
	displayScale  = 10
	displayWidth  = displayScale * emulator.DisplayWidth
	displayHeight = displayScale * emulator.DisplayHeight
)

const (
	debugCharacterWidth  = 6
	debugCharacterHeight = 16
	debugColumns         = 60
	debugRows            = 14
	debugPanelScale      = 2
	debugPanelWidth      = debugPanelScale * debugColumns * debugCharacterWidth
	debugPanelHeight     = debugPanelScale * debugRows * debugCharacterHeight
)

const (
	regularWindowWidth  = displayWidth
	regularWindowHeight = displayHeight
	debugWindowWidth    = displayWidth
	debugWindowHeight   = displayHeight + debugPanelHeight
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
	emulator   *emulator.Emulator
	debug      bool
	state      emulator.State
	display    *ebiten.Image
	debugPanel *ebiten.Image
}

func NewGame(e *emulator.Emulator) (*Game, error) {
	if e == nil {
		return nil, fmt.Errorf("no emulator provided")
	}

	return &Game{
		emulator:   e,
		display:    ebiten.NewImage(emulator.DisplayWidth, emulator.DisplayHeight),
		debugPanel: ebiten.NewImage(debugPanelWidth, debugPanelHeight),
	}, nil
}

func (g *Game) SetDebug(debug bool) {
	g.debug = debug
	g.adjustWindowSize()
}

func (g *Game) toggleDebug() {
	g.debug = !g.debug
	g.adjustWindowSize()
}

func (g *Game) adjustWindowSize() {
	if g.debug {
		ebiten.SetWindowSize(debugWindowWidth, debugWindowHeight)
	} else {
		ebiten.SetWindowSize(regularWindowWidth, regularWindowHeight)
	}
}

func (g *Game) Update() error {
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

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.toggleDebug()
	}

	if g.debug {
		if inpututil.IsKeyJustPressed(ebiten.KeyI) {
			g.emulator.Clock()
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyO) {
			g.emulator.Step()
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
	g.drawDisplay()

	var screenOptions ebiten.DrawImageOptions
	screenOptions.GeoM.Scale(displayScale, displayScale)

	screen.DrawImage(g.display, &screenOptions)

	if g.debug {
		g.drawDebugPanel()

		var debugPanelOptions ebiten.DrawImageOptions
		debugPanelOptions.GeoM.Scale(debugPanelScale, debugPanelScale)
		debugPanelOptions.GeoM.Translate(0, displayHeight)

		screen.DrawImage(g.debugPanel, &debugPanelOptions)
	}
}

func (g *Game) drawDisplay() {

	// This uses the same color palette of the original Game Boy, as documented by
	// https://en.wikipedia.org/wiki/List_of_video_game_console_palettes.

	for y := range g.state.Display {
		for x := range g.state.Display[y] {
			if g.state.Display[y][x] != 0 {
				g.display.Set(x, y, color.RGBA{R: 0x29, G: 0x41, B: 0x39, A: 0xff})
			} else {
				g.display.Set(x, y, color.RGBA{R: 0x7b, G: 0x82, B: 0x10, A: 0xff})
			}
		}
	}
}

func (g *Game) drawDebugPanel() {
	var w strings.Builder

	out := func(s string, args ...any) {
		_, _ = fmt.Fprintf(&w, s, args...)
	}

	out("Instruction:\n")
	out("> %v\n\n", debug.Instruction(g.state.Instruction()))
	out("Registers:")

	for i, v := range g.state.V {
		if i%8 == 0 {
			out("\n")
		} else {
			out(" ")
		}
		out("v%x=%02x", i, v)
	}

	out("\n\n")
	out("State:\n")
	out("pc=%04x ", g.state.PC)
	out("i=%04x ", g.state.I)
	out("sp=%02x ", g.state.SP)
	out("dt=%02x ", g.state.DT)
	out("st=%02x\n\n", g.state.ST)
	out("[I] Advance time\n")
	out("[O] Step instruction\n")
	out("[P] Toggle debug mode\n")

	g.debugPanel.Clear()

	ebitenutil.DebugPrint(g.debugPanel, w.String())
}

func (g *Game) Layout(_, _ int) (int, int) {
	if g.debug {
		return displayWidth, displayHeight + debugPanelHeight
	}

	return displayWidth, displayHeight
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

	g, err := NewGame(e)
	if err != nil {
		return fmt.Errorf("create game: %v", err)
	}

	g.SetDebug(debug)

	ebiten.SetWindowTitle("CHIP-8 Emulator")

	if err := ebiten.RunGame(g); err != nil {
		return fmt.Errorf("run game: %v", err)
	}

	return nil
}
