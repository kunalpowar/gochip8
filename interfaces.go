package chip8

// Display defines the interface required
// to mimic the screen for a chip 8.
type Display interface {
	// DrawFrame is called when once frame of display is done setting/clearing
	DrawFrame()

	// ClearPixel clears the pixel at location x,y or (row,col)
	ClearPixel(x, y int)

	// SetPixel sets the pixel at location x,y or (row,col)
	SetPixel(x, y int)
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
