package chip8

import (
	"io"
	"log"

	"github.com/kunalpowar/gochip8/display"
	"github.com/kunalpowar/gochip8/emulator"
)

type Chip8 struct {
	emulator *emulator.Emulator
	display  display.Display
}

func New(display display.Display) *Chip8 {
	if display == nil {
		log.Fatalf("chip8: need a non-empty display")
	}

	return &Chip8{emulator: emulator.New(), display: display}
}

func (c *Chip8) LoadROM(r io.Reader) {
	c.emulator.LoadROM(r)
}

// RunCycles runs the emulator for a limited set of cycles
// Use this to test simple roms like chip8 logo.
func (c *Chip8) RunCycles(limit int) {
	for i := 0; i < limit; i++ {
		c.emulator.UpdateDisplay = false
		c.emulator.EmulateCycle()

		if c.emulator.UpdateDisplay {
			c.display.DrawFrame(c.emulator.Display)
		}
	}
}
