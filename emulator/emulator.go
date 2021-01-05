package emulator

import (
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"time"
)

// References:
// http://devernay.free.fr/hacks/chip8/C8TECH10.HTM
// http://www.multigesture.net/articles/how-to-write-an-emulator-chip-8-interpreter/

const (
	interpreterStart  = 0x000
	programStart      = 0x200
	bytesPerCharacter = 5
)

var fontSet = []uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

// Emulator for chip8
type Emulator struct {
	ram [4096]uint8

	// 16 general purpose 8 bit registers
	v [16]uint8

	// 16 bit i register used to store memory address
	// Only the lowest 12 bits are used
	i uint16

	// dt is a delay timer
	dt uint8

	// st is a sound timer
	st uint8

	// pc is a 16 bit program counter
	pc uint16

	// sp is a 8 bit stackpointer
	sp uint8

	stack [16]uint16

	Keys [16]uint8

	// Display holds the data abotu current display state
	Display [32]uint64
	// UpdateDisplay is set if display frame needs to be updated.
	UpdateDisplay bool

	// Helper flags for efficient rendering
	ClearDisplay  bool
	ClearedPixels []Pixel
	SetPixels     []Pixel

	// Beep when set to true
	Beep bool
}

// Pixel represents one pixel on the 64X32 display
type Pixel struct{ X, Y int }

func dispPix(x, y int) Pixel {
	return Pixel{X: x % 64, Y: y % 32}
}

// New returns a new instance of emulator ready to load programs
func New() *Emulator {
	e := Emulator{pc: programStart}
	for i, b := range fontSet {
		e.ram[interpreterStart+i] = b
	}

	rand.Seed(time.Now().Unix())

	return &e
}

// LoadROM takes a rom io.reader to load to memory
func (e *Emulator) LoadROM(r io.Reader) {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatalf("could not read from rom: %v", err)
	}

	for i, b := range bs {
		e.ram[programStart+i] = b
	}

	log.Printf("emulator: loaded %d bytes of rom into memory", len(bs))
}

func (e *Emulator) updateTimers() {
	if e.dt > 0 {
		e.dt--
	}

	e.Beep = false
	if e.st > 0 {
		e.Beep = true
		e.st--
	}
}

func (e *Emulator) clearDisplay() {
	for i := range e.Display {
		e.Display[i] = 0x00
	}

	e.UpdateDisplay = true
	e.ClearDisplay = true
}

func (e *Emulator) togglePixel(p Pixel) {
	e.Display[p.Y] ^= ((0x1 << 63) >> p.X)
}

func (e *Emulator) getPixel(p Pixel) int {
	return int(e.Display[p.Y] & ((0x1 << 63) >> p.X))
}

// EmulateCycle runs the next opcode and updates the timers accordingly
func (e *Emulator) EmulateCycle() {
	var opcode uint16
	opcode = uint16(e.ram[e.pc])<<8 | uint16(e.ram[e.pc+1])

	e.ClearedPixels = make([]Pixel, 0)
	e.SetPixels = make([]Pixel, 0)
	e.UpdateDisplay = false
	e.ClearDisplay = false

	// nnn or addr - A 12-bit value, the lowest 12 bits of the instruction
	// n or nibble - A 4-bit value, the lowest 4 bits of the instruction
	// x - A 4-bit value, the lower 4 bits of the high byte of the instruction
	// y - A 4-bit value, the upper 4 bits of the low byte of the instruction
	// kk or byte - An 8-bit value, the lowest 8 bits of the instruction
	var (
		nnn       = opcode & 0x0fff
		n   uint8 = uint8(opcode & 0x000f)
		x   uint8 = uint8((opcode >> 8) & 0x000f)
		y   uint8 = uint8((opcode >> 4) & 0x000f)
		kk  uint8 = uint8(opcode & 0x00ff)
	)

	switch opcode & 0xf000 {
	case 0x0000:
		switch kk {
		case 0x00e0:
			e.clearDisplay()
			e.pc += 2
		case 0x00EE:
			e.sp--
			e.pc = e.stack[e.sp]
		}
	case 0x1000:
		e.pc = nnn
	case 0x2000:
		e.stack[e.sp] = e.pc + 2
		e.sp++
		e.pc = nnn
	case 0x3000:
		if e.v[x] == kk {
			e.pc += 4
		} else {
			e.pc += 2
		}
	case 0x4000:
		if e.v[x] != kk {
			e.pc += 4
		} else {
			e.pc += 2
		}
	case 0x5000:
		if e.v[x] == e.v[y] {
			e.pc += 4
		} else {
			e.pc += 2
		}
	case 0x6000:
		e.v[x] = kk
		e.pc += 2
	case 0x7000:
		e.v[x] += kk
		e.pc += 2
	case 0x8000:
		switch n {
		case 0x0000:
			e.v[x] = e.v[y]
		case 0x0001:
			e.v[x] |= e.v[y]
		case 0x0002:
			e.v[x] &= e.v[y]
		case 0x0003:
			e.v[x] ^= e.v[y]
		case 0x0004:
			if int(e.v[x])+int(e.v[y]) > 255 {
				e.v[0xF] = 1
			} else {
				e.v[0xF] = 0
			}
			e.v[x] += e.v[y]
		case 0x0005:
			if e.v[x] > e.v[y] {
				e.v[0xF] = 1
			} else {
				e.v[0xF] = 0
			}
			e.v[x] -= e.v[y]
		case 0x0006:
			e.v[0xF] = e.v[x] & 0x1
			e.v[x] >>= 1
		case 0x0007:
			if e.v[y] > e.v[x] {
				e.v[0xF] = 1
			} else {
				e.v[0xF] = 0
			}
			e.v[x] = e.v[y] - e.v[x]
		case 0x000E:
			e.v[0xF] = (e.v[x] >> 7) & 0x1
			e.v[x] <<= 1
		default:
			log.Fatalf("unknown opcode: %04x", e.pc)
		}
		e.pc += 2
	case 0x9000:
		if n == 0x0 {
			if e.v[x] != e.v[y] {
				e.pc += 4
			} else {
				e.pc += 2
			}
		} else {
			log.Fatalf("unknown opcode: %04x", e.pc)
		}
	case 0xa000:
		e.i = nnn
		e.pc += 2
	case 0xb000:
		e.pc = nnn + uint16(e.v[0])
	case 0xc000:
		e.v[x] = uint8(rand.Intn(255)) & kk
		e.pc += 2
	case 0xd000:
		e.v[0xF] = 0
		for row := 0; row < int(n); row++ {
			y := int(e.v[y]) + row
			bt := e.ram[int(e.i)+row]

			for col := 0; col < 8; col++ {
				x := int(e.v[x]) + col

				pix := e.getPixel(dispPix(x, y))
				if bt&(0x80>>col) != 0 {
					e.togglePixel(dispPix(x, y))
					if pix != 0 {
						e.v[0xF] = 1
						e.ClearedPixels = append(e.ClearedPixels, dispPix(x, y))
					} else {
						e.SetPixels = append(e.SetPixels, dispPix(x, y))
					}
				}
			}
		}
		e.UpdateDisplay = true
		e.pc += 2
	case 0xe000:
		switch kk {
		case 0x9e:
			if e.Keys[e.v[x]] != 0 {
				e.pc += 4
			} else {
				e.pc += 2
			}
		case 0xa1:
			if e.Keys[e.v[x]] == 0 {
				e.pc += 4
			} else {
				e.pc += 2
			}
		default:
			log.Fatalf("unknown opcode: %04x", e.pc)
		}
	case 0xf000:
		switch kk {
		case 0x07:
			e.v[x] = e.dt
			e.pc += 2

		case 0x0a:
			ticker := time.NewTicker(1 * time.Millisecond)
			log.Printf("emulator: waiting for key stroke...")
			for range ticker.C {
				keyPressed := -1
				for i, k := range e.Keys {
					if k > 0 {
						keyPressed = i
						break
					}
				}

				if keyPressed == -1 {
					continue
				}

				e.v[x] = uint8(keyPressed)
				ticker.Stop()
				break
			}
			log.Printf("emulator: got key stroke")
			e.pc += 2

		case 0x15:
			e.dt = e.v[x]
			e.pc += 2

		case 0x18:
			e.st = e.v[x]
			e.pc += 2

		case 0x1e:
			if (e.i + uint16(e.v[x])) > 0xfff {
				e.v[0xf] = 1
			} else {
				e.v[0xf] = 0
			}
			e.i += uint16(e.v[x])
			e.pc += 2

		case 0x29:
			e.i = uint16(bytesPerCharacter) * uint16(e.v[x])
			e.pc += 2

		case 0x33:
			intVal := int(e.v[x])
			e.ram[e.i] = uint8((intVal % 1000) / 100)
			e.ram[e.i+1] = uint8((intVal % 100) / 10)
			e.ram[e.i+2] = uint8(intVal % 10)
			e.pc += 2

		case 0x55:
			var i uint8
			for i = 0; i < x; i++ {
				e.ram[uint8(e.i&0xff)+i] = e.v[i]
			}
			e.i += uint16(x + 1)
			e.pc += 2

		case 0x65:
			var i uint8
			for i = 0; i < x; i++ {
				e.v[i] = e.ram[uint8(e.i&0xff)+i]
			}
			e.i += uint16(x + 1)
			e.pc += 2
		default:
			log.Fatalf("unknown opcode: %04x", e.pc)
		}

	default:
		log.Fatalf("unknown opcode: %04x", e.pc)
	}

	e.updateTimers()
}

func cicularShiftLeftUint8(in uint8, shiftBy int) uint8 {
	const bitLength = 8

	if shiftBy%bitLength == 0 {
		return in
	}

	by := shiftBy
	if by > bitLength {
		by = shiftBy % bitLength
	}

	out := in
	for i := 1; i <= by; i++ {
		msb := (out >> (bitLength - 1)) & 1
		out <<= 1
		out |= msb
	}
	return out
}

func cicularShiftLeftUint64(in uint64, shiftBy int) uint64 {
	const bitLength = 64

	if shiftBy%bitLength == 0 {
		return in
	}

	by := shiftBy
	if by > bitLength {
		by = shiftBy % bitLength
	}

	out := in
	for i := 1; i <= by; i++ {
		msb := (out >> (bitLength - 1)) & 1
		out <<= 1
		out |= msb
	}
	return out
}
