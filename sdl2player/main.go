package main

import (
	"flag"
	"image/color"
	"log"
	"os"
	"time"

	chip8 "github.com/kunalpowar/gochip8"
	"github.com/veandco/go-sdl2/sdl"
)

type sdlDisplay struct {
	window *sdl.Window
}

func (w sdlDisplay) DrawFrame(data [32]uint64) {
	surface, err := w.window.GetSurface()
	if err != nil {
		panic(err)
	}

	for y, rowData := range data {
		for x := 0; x < 64; x++ {
			if rowData&((0x1<<63)>>x) == 0 {
				surface.Set(x, y, color.Black)
			} else {
				surface.Set(x, y, color.White)
			}
		}
	}

	w.window.UpdateSurface()
}

type sdlKeyboard struct {
	pressedKeys map[int]bool
}

var keyboardToChip8Map = map[int]int{
	sdl.K_1: 1,
	sdl.K_2: 2,
	sdl.K_3: 3,
	sdl.K_q: 4,
	sdl.K_w: 5,
	sdl.K_e: 6,
	sdl.K_a: 7,
	sdl.K_s: 8,
	sdl.K_d: 9,
	sdl.K_z: 0xa,
	sdl.K_x: 0,
	sdl.K_c: 0xb,
	sdl.K_4: 0xc,
	sdl.K_r: 0xd,
	sdl.K_f: 0xe,
	sdl.K_v: 0xf,
}

func (s sdlKeyboard) PressedKeys() map[int]bool {
	out := make(map[int]bool)
	for k := range s.pressedKeys {
		if outK, present := keyboardToChip8Map[k]; present {
			out[outK] = true
		}
	}

	return out
}

var rom = flag.String("rom", "roms/Chip8 Picture.ch8", "-rom path_to_rom")

func main() {
	flag.Parse()
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 64, 32, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	surface.FillRect(nil, 0)
	window.UpdateSurface()

	kb := sdlKeyboard{pressedKeys: make(map[int]bool)}
	c := chip8.New(sdlDisplay{window: window}, &kb, nil)

	flag.Parse()
	f, err := os.Open(*rom)
	if err != nil {
		log.Fatalf("could not open rom: %v", err)
	}
	defer f.Close()
	c.LoadROM(f)

	ticker := time.NewTicker(5 * time.Millisecond)

	running := true
	for range ticker.C {
		c.RunOnce()
		if !running {
			break
		}

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				ticker.Stop()
				running = false
				break

			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN {
					kb.pressedKeys[int(e.Keysym.Sym)] = true
				}
				if e.Type == sdl.KEYUP {
					delete(kb.pressedKeys, int(e.Keysym.Sym))
				}
			}
		}
	}
}
