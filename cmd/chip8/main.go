package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"sync"
	"time"

	"github.com/francescomari/chip-8/emulator"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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

type Game struct {
	emulator *emulator.Emulator
	once     sync.Once
}

func (g *Game) Update() error {
	g.once.Do(func() {
		go func() {
			for {
				g.emulator.Step()
			}
		}()
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

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	var display emulator.Display

	g.emulator.Display(&display)

	// This uses the same color palette of the original Game Boy, as documented by
	// https://en.wikipedia.org/wiki/List_of_video_game_console_palettes.

	for y := range display {
		for x := range display[y] {
			if display[y][x] != 0 {
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
	flag.Parse()

	if flag.NArg() != 1 {
		return fmt.Errorf("invalid number of arguments")
	}

	rom, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		return fmt.Errorf("read file: %v", err)
	}

	context := audio.NewContext(44100)

	stream, err := wav.DecodeWithoutResampling(bytes.NewReader(beep))
	if err != nil {
		return fmt.Errorf("decode beep audio: %v", err)
	}

	player, err := context.NewPlayer(stream)
	if err != nil {
		return fmt.Errorf("create beep player: %v", err)
	}

	e := emulator.New()

	e.Load(rom)

	e.SetSound(func() {
		player.Rewind()
		player.Play()
		waitForPlayer(player)
	})

	game := Game{
		emulator: e,
	}

	ebiten.SetWindowSize(10*emulator.DisplayWidth, 10*emulator.DisplayHeight)
	ebiten.SetWindowTitle("CHIP-8 Emulator")

	if err := ebiten.RunGame(&game); err != nil {
		return fmt.Errorf("run game: %v", err)
	}

	return nil
}

func waitForPlayer(p *audio.Player) {
	var last time.Duration

	for {
		curr := p.Position()

		if curr == last {
			break
		}

		last = curr
	}
}
