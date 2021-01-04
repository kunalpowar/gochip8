package main

import (
	"flag"
	"image"
	"image/color"
	"image/gif"
	"log"
	"os"

	chip8 "github.com/kunalpowar/gochip8"
)

var rom = flag.String("rom", "roms/Chip8 Picture.ch8", "-rom path_to_rom")

func main() {
	flag.Parse()

	var d GIFDisplay
	c := chip8.New(&d, nil, nil)

	f, err := os.Open(*rom)
	if err != nil {
		log.Fatalf("could not open rom: %v", err)
	}
	defer f.Close()
	c.LoadROM(f)

	c.RunCycles(1000)

	out, err := os.Create("out.gif")
	if err != nil {
		log.Fatalf("could not create gif: %v", err)
	}
	defer out.Close()

	if err := gif.EncodeAll(out, &gif.GIF{Image: d.images, Delay: d.delays}); err != nil {
		log.Fatalf("could not encode to gif: %v", err)
	}
}

// GIFDisplay creates a gif frame for every display update
// At the end of limited cycles, a gif file will be encoded
// with all the frames.
type GIFDisplay struct {
	images []*image.Paletted
	delays []int
}

func (d *GIFDisplay) DrawFrame(data [32]uint64) {
	palette := []color.Color{color.White, color.Black}
	rect := image.Rect(0, 0, 64, 32)

	img := image.NewPaletted(rect, palette)
	for y, rowData := range data {
		for x := 0; x < 64; x++ {
			if rowData&((0x1<<63)>>x) == 0 {
				img.Set(x, y, color.Black)
			} else {
				img.Set(x, y, color.White)
			}
		}
	}

	d.images = append(d.images, img)
	d.delays = append(d.delays, 0)
}
