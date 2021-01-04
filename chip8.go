package chip8

import (
	"io"
	"log"

	"github.com/kunalpowar/gochip8/emulator"
)

// Chip8 packages all the intefaces and the emulator itself.
// Use New(...) to create a instance.
type Chip8 struct {
	emulator *emulator.Emulator
	display  Display
	keyboard Keyboard
	speaker  Speaker
}

// New creates and initialises the emulator and all the interfaces.
func New(disp Display, kb Keyboard, sp Speaker) *Chip8 {
	if disp == nil {
		log.Fatalf("chip8: need a non-empty display")
	}

	return &Chip8{emulator: emulator.New(), display: disp, keyboard: kb, speaker: sp}
}

// LoadROM loads the bytes from reader into chip8 RAM
func (c *Chip8) LoadROM(r io.Reader) {
	c.emulator.LoadROM(r)
}

// RunCycles runs the emulator for a limited set of cycles
// Use this to test simple roms like chip8 logo.
func (c *Chip8) RunCycles(limit int) {
	for i := 0; i < limit; i++ {
		c.RunOnce()
	}
}

// RunOnce runs just 1 emulator cycle
func (c *Chip8) RunOnce() {
	c.emulator.UpdateDisplay = false

	c.emulator.EmulateCycle()

	if c.emulator.UpdateDisplay {
		c.updateDisplay()
	}

	c.setKeys()

	if c.speaker != nil {
		if c.emulator.Beep {
			c.speaker.Beep()
		} else {
			c.speaker.Pause()
		}
	}

}

func (c *Chip8) setKeys() {
	if c.keyboard == nil {
		return
	}

	keys := c.keyboard.PressedKeys()
	for i := 0; i < 16; i++ {
		if keys[i] {
			c.emulator.Keys[i] = 1
		} else {
			c.emulator.Keys[i] = 0
		}
	}
}

func (c *Chip8) updateDisplay() {
	for y, rowData := range c.emulator.Display {
		for x := 0; x < 64; x++ {
			if rowData&((0x1<<63)>>x) == 0 {
				c.display.ClearPixel(x, y)
				// img.Set(x, y, color.Black)
			} else {
				c.display.SetPixel(x, y)
				// img.Set(x, y, color.White)
			}
		}
	}

	c.display.DrawFrame()
}
