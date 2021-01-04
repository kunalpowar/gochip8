package chip8

import (
	"io"
	"log"

	"github.com/kunalpowar/gochip8/display"
	"github.com/kunalpowar/gochip8/emulator"
	"github.com/kunalpowar/gochip8/keyboard"
	"github.com/kunalpowar/gochip8/speaker"
)

type Chip8 struct {
	emulator *emulator.Emulator
	display  display.Display
	keyboard keyboard.Keyboard
	speaker  speaker.Speaker
}

func New(disp display.Display, kb keyboard.Keyboard) *Chip8 {
	if disp == nil {
		log.Fatalf("chip8: need a non-empty display")
	}

	return &Chip8{emulator: emulator.New(), display: disp, keyboard: kb}
}

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

// RunCycles runs just 1 emulator cycle
func (c *Chip8) RunOnce() {
	c.emulator.UpdateDisplay = false
	c.emulator.EmulateCycle()

	if c.emulator.UpdateDisplay {
		c.display.DrawFrame(c.emulator.Display)
	}

	c.SetKeys()
	if c.emulator.Beep && c.speaker != nil {
		c.speaker.Beep()
	}
}

func (c *Chip8) SetKeys() {
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
