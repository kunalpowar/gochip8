package display

// Display defines the interface required
// to mimic the screen for a chip 8.
type Display interface {
	// DrawFrame draws a display frame based
	// on the 32*64 bit display data from emulator.
	DrawFrame([32]uint64)
}
