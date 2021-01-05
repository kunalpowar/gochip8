package main

import (
	"flag"
	"image/color"
	"log"
	"os"
	"time"

	"github.com/kunalpowar/gochip8/pkg/chip8"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	displayScale = 10
	c8DispWidth  = 64
	c8DispHeight = 32
)

type sdlDisplay struct {
	window *sdl.Window
}

// DrawFrame will call the update surface which should update
// the current frame with all the previous changes.
func (d sdlDisplay) DrawFrame() {
	d.window.UpdateSurface()
}

// ClearPixel will set the pixel at location to black.
func (d sdlDisplay) ClearPixel(x, y int) {
	surface, err := d.window.GetSurface()
	if err != nil {
		panic(err)
	}

	x = x * displayScale
	y = y * displayScale
	for row := 0; row < displayScale; row++ {
		for col := 0; col < displayScale; col++ {
			surface.Set(x+col, y+row, color.Black)
		}
	}
}

// ClearPixel will set the pixel at location to white.
func (d sdlDisplay) SetPixel(x, y int) {
	surface, err := d.window.GetSurface()
	if err != nil {
		panic(err)
	}

	x = x * displayScale
	y = y * displayScale
	for row := 0; row < displayScale; row++ {
		for col := 0; col < displayScale; col++ {
			surface.Set(x+col, y+row, color.White)
		}
	}
}

// ClearAll adds a black frame to the display.
func (d sdlDisplay) ClearAll() {
	surface, err := d.window.GetSurface()
	if err != nil {
		panic(err)
	}

	surface.FillRect(nil, 0)
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

	window, err := sdl.CreateWindow(
		"test",
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		c8DispWidth*displayScale, c8DispHeight*displayScale,
		sdl.WINDOW_SHOWN)
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

	ticker := time.NewTicker(1 * time.Millisecond)

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
