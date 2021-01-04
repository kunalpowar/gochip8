package speaker

// Speaker acts as a sound device for the emulator
type Speaker interface {
	// Beep should play some sound
	Beep()

	// Pause should pause any sound that's playing
	Pause()
}
