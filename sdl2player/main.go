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

type SDLWindow struct {
	window *sdl.Window
}

func (w SDLWindow) DrawFrame(data [32]uint64) {
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

type SDLKeyboard struct {
	pressedKeys map[int]bool
}

func (s SDLKeyboard) PressedKeys() map[int]bool {
	out := make(map[int]bool)
	for k := range s.pressedKeys {
		switch k {
		case sdl.K_1:
			out[1] = true
		case sdl.K_2:
			out[2] = true
		case sdl.K_3:
			out[3] = true
		case sdl.K_q:
			out[4] = true
		case sdl.K_w:
			out[5] = true
		case sdl.K_e:
			out[6] = true
		case sdl.K_a:
			out[7] = true
		case sdl.K_s:
			out[8] = true
		case sdl.K_d:
			out[9] = true
		case sdl.K_z:
			out[0xa] = true
		case sdl.K_x:
			out[0] = true
		case sdl.K_c:
			out[0xb] = true
		case sdl.K_4:
			out[0xc] = true
		case sdl.K_r:
			out[0xd] = true
		case sdl.K_f:
			out[0xe] = true
		case sdl.K_v:
			out[0xf] = true
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

	kb := SDLKeyboard{pressedKeys: make(map[int]bool)}
	c := chip8.New(SDLWindow{window: window}, &kb, nil)

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
