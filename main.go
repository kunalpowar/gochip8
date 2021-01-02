package main

import (
	"log"
	"os"

	"github.com/kunalpowar/gochip8/emulator"
)

func main() {
	f, err := os.Open("roms/Chip8 Picture.ch8")
	if err != nil {
		log.Fatalf("could not open rom: %v", err)
	}
	defer f.Close()

	emulator.Run(f)
}
