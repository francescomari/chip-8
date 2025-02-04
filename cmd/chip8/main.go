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
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/francescomari/chip-8/emulator"
)

//go:embed beep.wav
var beep []byte

var mappings map[ebiten.Key]uint8 = map[ebiten.Key]uint8{
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

type State struct {
	V      emulator.Registers
	I      uint16
	SP     uint8
	DT     uint8
	ST     uint8
	PC     uint16
	Memory emulator.Memory
}

type Game struct {
	emulator     *emulator.Emulator
	debug        bool
	display      *emulator.Display
	state        *State
	once         sync.Once
	keyDownCh    chan uint8
	keyUpCh      chan uint8
	displayCh    chan chan *emulator.Display
	startDebugCh chan struct{}
	stopDebugCh  chan struct{}
	stepCh       chan time.Time
	clockCh      chan time.Time
	stateCh      chan chan *State
}

func (g *Game) Update() error {
	g.once.Do(func() {
		g.keyDownCh = make(chan uint8)
		g.keyUpCh = make(chan uint8)
		g.displayCh = make(chan chan *emulator.Display)
		g.startDebugCh = make(chan struct{})
		g.stopDebugCh = make(chan struct{})
		g.stepCh = make(chan time.Time)
		g.clockCh = make(chan time.Time)
		g.stateCh = make(chan chan *State)

		go func() {
			var (
				tickerStepCh  = time.Tick(time.Second / 500)
				tickerClockCh = time.Tick(time.Second / 60)
			)

			var (
				stepCh  <-chan time.Time
				clockCh <-chan time.Time
			)

			if g.debug {
				stepCh = g.stepCh
				clockCh = g.clockCh
			} else {
				stepCh = tickerStepCh
				clockCh = tickerClockCh
			}

			for {
				select {
				case <-stepCh:
					g.emulator.Step()
				case <-clockCh:
					g.emulator.Clock()
				case k := <-g.keyDownCh:
					g.emulator.KeyDown(k)
				case k := <-g.keyUpCh:
					g.emulator.KeyUp(k)
				case displayCh := <-g.displayCh:
					var d emulator.Display
					g.emulator.Display(&d)
					displayCh <- &d
				case <-g.startDebugCh:
					stepCh = g.stepCh
					clockCh = g.clockCh
				case <-g.stopDebugCh:
					stepCh = tickerStepCh
					clockCh = tickerClockCh
				case stateCh := <-g.stateCh:
					s := State{
						I:  g.emulator.I(),
						SP: g.emulator.SP(),
						DT: g.emulator.DT(),
						ST: g.emulator.ST(),
						PC: g.emulator.PC(),
					}
					g.emulator.V(&s.V)
					g.emulator.Memory(&s.Memory)
					stateCh <- &s
				}
			}
		}()
	})

	var keys [16]ebiten.Key

	for _, key := range inpututil.AppendJustPressedKeys(keys[:0]) {
		if value, ok := mappings[key]; ok {
			g.keyDownCh <- value
		}
	}

	for _, key := range inpututil.AppendJustReleasedKeys(keys[:0]) {
		if value, ok := mappings[key]; ok {
			g.keyUpCh <- value
		}
	}

	displayCh := make(chan *emulator.Display)
	g.displayCh <- displayCh
	g.display = <-displayCh

	if inpututil.IsKeyJustPressed(ebiten.KeyF12) {
		if g.debug {
			g.stopDebugCh <- struct{}{}
		} else {
			g.startDebugCh <- struct{}{}
		}

		g.debug = !g.debug
	}

	var printState bool

	if inpututil.IsKeyJustPressed(ebiten.KeyF10) && g.debug {
		g.stepCh <- time.Now()
		printState = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF11) && g.debug {
		g.clockCh <- time.Now()
		printState = true
	}

	if printState {
		stateCh := make(chan *State)
		g.stateCh <- stateCh
		g.state = <-stateCh

		logState(g.state)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// This uses the same color palette of the original Game Boy, as documented by
	// https://en.wikipedia.org/wiki/List_of_video_game_console_palettes.

	for y := range g.display {
		for x := range g.display[y] {
			if g.display[y][x] != 0 {
				screen.Set(x, y, color.RGBA{R: 0x29, G: 0x41, B: 0x39, A: 1})
			} else {
				screen.Set(x, y, color.RGBA{R: 0x7b, G: 0x82, B: 0x10, A: 1})
			}
		}
	}
}

func (g *Game) Layout(w, h int) (int, int) {
	return emulator.DisplayWidth, emulator.DisplayHeight
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

func logState(s *State) {
	log.Printf("---> %s", decodeOp(s))

	var b strings.Builder

	for i, v := range s.V {
		if i > 0 {
			fmt.Fprintf(&b, ", v%x = %02x", i, v)
		} else {
			fmt.Fprintf(&b, "v%x = %02x", i, v)
		}
	}

	log.Println(b.String())

	log.Printf("i = %04x sp = %02x, dt = %02x, st = %02x, pc = %04x", s.I, s.SP, s.DT, s.ST, s.PC)
}

func decodeOp(s *State) string {
	var (
		hi     = uint16(s.Memory[s.PC]) << 8
		lo     = uint16(s.Memory[s.PC+1])
		op     = hi | lo
		x      = fmt.Sprintf("v%x", (op&0x0f00)>>8)
		y      = fmt.Sprintf("v%x", (op&0x00f0)>>4)
		n      = fmt.Sprintf("%03x", op&0x0fff)
		k      = fmt.Sprintf("%02x", op&0x00ff)
		nibble = fmt.Sprintf("%x", op&0x000f)
	)

	switch op & 0xf000 {
	case 0x0000:
		switch op & 0x00ff {
		case 0x00e0:
			return "cls"
		case 0x00ee:
			return "ret"
		}
	case 0x1000:
		return fmt.Sprintf("jp %s", n)
	case 0x2000:
		return fmt.Sprintf("call %s", n)
	case 0x3000:
		return fmt.Sprintf("se %s, %s", x, k)
	case 0x4000:
		return fmt.Sprintf("sne %s, %s", x, k)
	case 0x5000:
		return fmt.Sprintf("se %s, %s", x, y)
	case 0x6000:
		return fmt.Sprintf("ld %s, %s", x, k)
	case 0x7000:
		return fmt.Sprintf("add %s, %s", x, k)
	case 0x8000:
		switch op & 0x000f {
		case 0x0:
			return fmt.Sprintf("ld %s, %s", x, y)
		case 0x1:
			return fmt.Sprintf("or %s, %s", x, y)
		case 0x2:
			return fmt.Sprintf("and %s, %s", x, y)
		case 0x3:
			return fmt.Sprintf("xor %s, %s", x, y)
		case 0x4:
			return fmt.Sprintf("add %s, %s", x, y)
		case 0x5:
			return fmt.Sprintf("sub %s, %s", x, y)
		case 0x6:
			return fmt.Sprintf("shr %s, %s", x, y)
		case 0x7:
			return fmt.Sprintf("subn %s, %s", x, y)
		case 0xe:
			return fmt.Sprintf("shl %s, %s", x, y)
		}
	case 0x9000:
		return fmt.Sprintf("sne %s, %s", x, y)
	case 0xa000:
		return fmt.Sprintf("ld i, %s", n)
	case 0xb000:
		return fmt.Sprintf("jp v0, %s", n)
	case 0xc000:
		return fmt.Sprintf("rnd %s, %s", x, k)
	case 0xd000:
		return fmt.Sprintf("draw %s, %s, %s", x, y, nibble)
	case 0xe000:
		switch op & 0xff {
		case 0x9e:
			return fmt.Sprintf("skp %s", x)
		case 0xa1:
			return fmt.Sprintf("sknp %s", x)
		}
	case 0xf000:
		switch op & 0xff {
		case 0x07:
			return fmt.Sprintf("ld %s, dt", x)
		case 0x0a:
			return fmt.Sprintf("ld %s, k", x)
		case 0x15:
			return fmt.Sprintf("ld dt, %s", x)
		case 0x18:
			return fmt.Sprintf("ld st, %s", x)
		case 0x1e:
			return fmt.Sprintf("add i, %s", x)
		case 0x29:
			return fmt.Sprintf("ld f, %s", x)
		case 0x33:
			return fmt.Sprintf("ld b, %s", x)
		case 0x55:
			return fmt.Sprintf("ld [i], %s", x)
		case 0x65:
			return fmt.Sprintf("ld %s, [i]", x)
		}
	}

	return fmt.Sprintf("unknown (%04x)", op)
}
