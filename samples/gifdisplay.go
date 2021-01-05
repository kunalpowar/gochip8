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

	d := GIFDisplay{currentFrame: newPalettedImage()}
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

	currentFrame *image.Paletted
}

func newPalettedImage() *image.Paletted {
	palette := []color.Color{color.Black, color.White}
	rect := image.Rect(0, 0, 64, 32)
	return image.NewPaletted(rect, palette)
}

// DrawFrame will add the currentframe as an image to the list
// to be encoded as gif the end.
func (d *GIFDisplay) DrawFrame() {
	if d.currentFrame == nil {
		return
	}

	d.images = append(d.images, d.currentFrame)
	d.delays = append(d.delays, 0)

	img := newPalettedImage()
	rect := img.Bounds()
	for row := 0; row < rect.Dy(); row++ {
		for col := 0; col < rect.Dx(); col++ {
			img.Set(col, row, d.currentFrame.At(col, row))
		}
	}

	d.currentFrame = img
}

// ClearPixel will set the pixel at location to black.
func (d *GIFDisplay) ClearPixel(x, y int) {
	d.currentFrame.Set(x, y, color.Black)
}

// SetPixel will set the pixel at location to white.
func (d *GIFDisplay) SetPixel(x, y int) {
	d.currentFrame.Set(x, y, color.White)
}

// ClearAll will clear all pixels by using a new paletted image.
func (d *GIFDisplay) ClearAll() {
	d.currentFrame = newPalettedImage()
}
