package main

import (
	"flag"
	"image/color"
	"log"
	"os"

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

	c := chip8.New(SDLWindow{window: window})

	flag.Parse()
	f, err := os.Open(*rom)
	if err != nil {
		log.Fatalf("could not open rom: %v", err)
	}
	defer f.Close()
	c.LoadROM(f)

	running := true
	for running {
		c.RunOnce()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			}
		}
	}
}
