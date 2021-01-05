# Chip-8 Emulator with Go 
[![Go Report Card](https://goreportcard.com/badge/github.com/kunalpowar/gochip8)](https://goreportcard.com/report/github.com/kunalpowar/gochip8)

Pet project to write a Chip-8 emulator with go.

## References
* [Cowdgod's technical reference](http://devernay.free.fr/hacks/chip8/C8TECH10.HTM)
* [How to write a Chip-8 emulator](http://www.multigesture.net/articles/how-to-write-an-emulator-chip-8-interpreter/)
* [Game ROMs](https://github.com/mir3z/chip8-emu/tree/master/roms)

## Description

The idea of the repository is to provide a chip8 emulator to be used with any display, keyboard and speaker interface. This is done using by satisfying the interfaces as shown below.

```go
// Display defines the interface required
// to mimic the screen for a chip 8.
type Display interface {
	// DrawFrame is called when once frame of display is done setting/clearing
	DrawFrame()

	// ClearPixel clears the pixel at location x,y or (row,col)
	ClearPixel(x, y int)

	// SetPixel sets the pixel at location x,y or (row,col)
	SetPixel(x, y int)

	// ClearAll should clear all the pixels
	ClearAll()
}

// Keyboard exposes mapping to emulator defined keys
// Refer http://devernay.free.fr/hacks/chip8/C8TECH10.HTM#keyboard
type Keyboard interface {
	PressedKeys() map[int]bool
}

// Speaker acts as a sound device for the emulator
type Speaker interface {
	// Beep should play some sound
	Beep()

	// Pause should pause any sound that's playing
	Pause()
}
```

With the interfaces, you can create a Chip-8 instance and run it as follows

```go
import chip8 "github.com/kunalpowar/gochip8"

c := chip8.New(myDisplay, myKeyboard, mySpeaker)
c.RunOnce()
```

## Examples

### SDL2 Player
Refer to [`sdl2player`](/sdl2player) to which uses [sdl2](https://github.com/veandco/go-sdl2) based Display and Keyboard.

### GIF
This is a sample code to generate a gif of the display by using the `Display` interface. The code runs the emulator for 1000 cycles to generate a gif.