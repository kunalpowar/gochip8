package keyboard

type Keyboard interface {
	PressedKeys() map[int]bool
}
