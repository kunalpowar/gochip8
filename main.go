package main

import (
	"flag"
	"log"
	"os"

	"github.com/kunalpowar/gochip8/emulator"
)

var rom = flag.String("rom", "roms/Chip8 Picture.ch8", "-rom path_to_rom")

func main() {
	flag.Parse()

	e := emulator.New()

	f, err := os.Open(*rom)
	if err != nil {
		log.Fatalf("could not open rom: %v", err)
	}
	defer f.Close()
	e.LoadROM(f)

	e.Run()
}
