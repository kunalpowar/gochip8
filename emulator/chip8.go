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
	RAM [4096]uint8

	// 16 general purpose 8 bit registers
	V [16]uint8

	// 16 bit I register used to store memory address
	// Only the lowest 12 bits are used
	I uint16

	// DT is a delay timer
	DT uint8

	// ST is a sound timer
	ST uint8

	// PC is a 16 bit program counter
	PC uint16

	// SP is a 8 bit stackpointer
	SP uint8

	Stack [16]uint16

	Display [32]uint64
	Keys    [16]uint8

	// UpdateDisplay is set if display frame needs to be updated.
	UpdateDisplay bool
}

// New returns a new instance of emulator ready to load programs
func New() *Emulator {
	e := Emulator{PC: programStart}
	for i, b := range fontSet {
		e.RAM[interpreterStart+i] = b
	}

	return &e
}

// LoadROM takes a rom io.reader to load to memory
func (e *Emulator) LoadROM(r io.Reader) {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatalf("could not read from rom: %v", err)
	}

	for i, b := range bs {
		e.RAM[programStart+i] = b
	}

	log.Printf("emulator: loaded %d bytes of rom into memory", len(bs))
}

func (e *Emulator) updateTimers() {
	if e.DT > 0 {
		e.DT--
	}

	if e.ST > 0 {
		print("beep..")
		e.ST--
	}
}

func (e *Emulator) clearDisplay() {
	for i := range e.Display {
		e.Display[i] = 0x00
	}
	e.UpdateDisplay = true
}

func (e *Emulator) togglePixel(x, y int) {
	e.Display[y] ^= ((0x1 << 63) >> x)
}

func (e *Emulator) getPixel(x, y int) int {
	return int(e.Display[y] & ((0x1 << 63) >> x))
}

func (e *Emulator) EmulateCycle() {
	var opcode uint16
	opcode = uint16(e.RAM[e.PC])<<8 | uint16(e.RAM[e.PC+1])

	// nnn or addr - A 12-bit value, the lowest 12 bits of the instruction
	// n or nibble - A 4-bit value, the lowest 4 bits of the instruction
	// x - A 4-bit value, the lower 4 bits of the high byte of the instruction
	// y - A 4-bit value, the upper 4 bits of the low byte of the instruction
	// kk or byte - An 8-bit value, the lowest 8 bits of the instruction
	var (
		nnn = opcode & 0x0fff
		n   = opcode & 0x000f
		x   = (opcode >> 8) & 0x000f
		y   = (opcode >> 4) & 0x000f
		kk  = opcode & 0x00ff
	)

	log.Printf("opcode: %04x, PC: %04x", opcode, e.PC)

	switch opcode & 0xf000 {
	case 0x0000:
		switch kk {
		case 0x00e0:
			log.Print("clear display")
			e.clearDisplay()
			e.PC += 2
		case 0x00EE:
			log.Print("Return from a subroutine")
			e.PC = e.Stack[e.SP]
			e.SP--
		}
	case 0x1000:
		log.Printf("Jump to location nnn:%x.", nnn)
		e.PC = nnn
	case 0x2000:
		log.Printf("Call subroutine at nnn:%x", nnn)
		e.SP++
		e.Stack[e.SP] = e.PC + 2
		e.PC = nnn
	case 0x3000:
		log.Printf("Skip next instruction if Vx(%x) = kk(%x).", e.V[x], kk)
		if uint16(e.V[x]) == kk {
			e.PC += 4
		} else {
			e.PC += 2
		}
	case 0x4000:
		log.Printf("Skip next instruction if Vx(%x) ! kk(%x).", e.V[x], kk)
		if uint16(e.V[x]) != kk {
			e.PC += 4
		} else {
			e.PC += 2
		}
	case 0x5000:
		log.Printf("Skip next instruction if Vx(%x) = Vy(%x).", e.V[x], e.V[y])
		if e.V[x] == e.V[y] {
			e.PC += 4
		} else {
			e.PC += 2
		}
	case 0x6000:
		log.Printf("Set Vx(%x) = kk(%x).", e.V[x], kk)
		e.V[x] = uint8(kk)
		log.Printf("e.V[x]: %d %04x", e.V[x], e.V[x])
		e.PC += 2
	case 0x7000:
		log.Printf("Set Vx (%d) = Vx + kk(%d).", e.V[x], kk)
		e.V[x] += uint8(kk)
		log.Printf("e.V[x]: %d %04x", e.V[x], e.V[x])
		e.PC += 2
	case 0x8000:
		switch n {
		case 0x0000:
			log.Print("Set Vx = Vy.")
			e.V[x] = e.V[y]
			log.Printf("e.V[x]: %d %04x", e.V[x], e.V[x])
		case 0x0001:
			log.Print("Set Vx = Vx OR Vy.")
			e.V[x] |= e.V[y]
		case 0x0002:
			log.Print("Set Vx = Vx AND Vy.")
			e.V[x] &= e.V[y]
		case 0x0003:
			log.Print("Set Vx = Vx XOR Vy.")
			e.V[x] ^= e.V[y]
		case 0x0004:
			log.Print("Set Vx = Vx + Vy, set VF = carry")
			if int(e.V[x])+int(e.V[y]) > 255 {
				e.V[0xF] = 1
			} else {
				e.V[0xF] = 0
			}
			e.V[x] += e.V[y]
			log.Printf("e.V[x]: %d %04x", e.V[x], e.V[x])
		case 0x0005:
			log.Printf("Set Vx = Vx - Vy, set VF = NOT borrow.")
			if e.V[x] > e.V[y] {
				e.V[0xF] = 1
			} else {
				e.V[0xF] = 0
			}
			e.V[x] -= e.V[y]
			log.Printf("e.V[x]: %d %04x", e.V[x], e.V[x])
		case 0x0006:
			log.Print("Set Vx = Vx SHR 1.")
			e.V[0xF] = e.V[x] & 0x1
			e.V[x] >>= 1
			log.Printf("e.V[x]: %d %04x", e.V[x], e.V[x])
		case 0x0007:
			log.Print("Set Vx = Vy - Vx, set VF = NOT borrow.")
			if e.V[y] > e.V[x] {
				e.V[0xF] = 1
			} else {
				e.V[0xF] = 0
			}
			e.V[x] = e.V[y] - e.V[x]
			log.Printf("e.V[x]: %d %04x", e.V[x], e.V[x])
		case 0x000E:
			log.Print("Set Vx = Vx SHL 1.")
			e.V[0xF] = (e.V[x] >> 7) & 0x1
			e.V[x] <<= 1
			log.Printf("e.V[x]: %d %04x", e.V[x], e.V[x])
		}
		e.PC += 2
	case 0x9000:
		log.Print("Skip next instruction if Vx != Vy.")
		if n == 0x0 {
			if e.V[x] != e.V[y] {
				e.PC += 4
			} else {
				e.PC += 2
			}
		}
	case 0xa000:
		log.Print("Set I = nnn.")
		e.I = nnn
		e.PC += 2
	case 0xb000:
		log.Print("Jump to location nnn + V0")
		e.PC = nnn + uint16(e.V[0])
	case 0xc000:
		log.Print("Set Vx = random byte AND kk.")
		e.V[x] = uint8(rand.Intn(255)) & uint8(kk)
		log.Printf("e.V[x]: %d %04x", e.V[x], e.V[x])
		e.PC += 2
	case 0xd000:
		log.Printf("Display n(%d)-byte sprite starting at memory location I(%d) at (Vx:%d, Vy:%d), set VF = collision.", n, e.I, e.V[x], e.V[y])

		e.V[0xF] = 0
		for row := 0; row < int(n); row++ {
			y := int(e.V[y]) + row
			bt := e.RAM[int(e.I)+row]
			log.Printf("byte: %08b", bt)

			for col := 0; col < 8; col++ {
				x := int(e.V[x]) + col

				pix := e.getPixel(x, y)
				if bt&(0x80>>col) != 0 {
					e.togglePixel(x, y)
					if pix != 0 {
						e.V[0xF] = 1
					}
				}
			}
		}
		e.UpdateDisplay = true
		e.PC += 2
	case 0xe000:
		switch kk {
		case 0x009e:
			log.Print("Skip next instruction if key with the value of Vx is pressed.")
			if e.Keys[e.V[x]] > 0 {
				e.PC += 4
			} else {
				e.PC += 2
			}
		case 0x00a1:
			log.Print("Skip next instruction if key with the value of Vx is not pressed.")
			if e.Keys[e.V[x]] == 0 {
				e.PC += 4
			} else {
				e.PC += 2
			}
		}
	case 0xf000:
		switch kk {
		// Fx07 - LD Vx, DT
		// Set Vx = delay timer value.
		// The value of DT is placed into Vx.
		case 0x0007:
			log.Print("Set Vx = delay timer value.")
			e.V[x] = e.DT
			log.Printf("e.V[x]: %d %04x", e.V[x], e.V[x])
			e.PC += 2

		// Fx0A - LD Vx, K
		// Wait for a key press, store the value of the key in Vx.
		// All execution stops until a key is pressed, then the value of that key is stored in Vx.
		case 0x000a:
			log.Print("Wait for a key press, store the value of the key in Vx.")
			ticker := time.NewTicker(10 * time.Millisecond)
			for range ticker.C {
				keyPressed := -1
				for i, k := range e.Keys {
					if k > 0 {
						log.Printf("got key press: %d", i)
						keyPressed = i
						break
					}
				}

				if keyPressed == -1 {
					continue
				}

				e.V[x] = uint8(keyPressed)
				log.Printf("e.V[x]: %d %04x", e.V[x], e.V[x])
				ticker.Stop()
				break
			}

			e.PC += 2

		// Fx15 - LD DT, Vx
		// Set delay timer = Vx.
		// DT is set equal to the value of Vx.
		case 0x0015:
			e.DT = e.V[x]
			e.PC += 2

		// Fx18 - LD ST, Vx
		// Set sound timer = Vx.
		// ST is set equal to the value of Vx.
		case 0x0018:
			e.ST = e.V[x]
			e.PC += 2

		// Fx1E - ADD I, Vx
		// Set I = I + Vx.
		// The values of I and Vx are added, and the results are stored in I.
		case 0x001e:
			if int(e.I)+int(e.V[x]) > 255 {
				e.V[0xf] = 1
			} else {
				e.V[0xf] = 0
			}
			e.I += uint16(e.V[x])
			e.PC += 2

		// Fx29 - LD F, Vx
		// Set I = location of sprite for digit Vx.
		// The value of I is set to the location for the hexadecimal sprite corresponding to the value of Vx. See section 2.4, Display, for more information on the Chip-8 hexadecimal font.
		case 0x0029:
			e.I = uint16(bytesPerCharacter) * uint16(e.V[x])
			e.PC += 2

		// Fx33 - LD B, Vx
		// Store BCD representation of Vx in memory locations I, I+1, and I+2.
		// The interpreter takes the decimal value of Vx, and places the hundreds digit in memory at location in I, the tens digit at location I+1, and the ones digit at location I+2.
		case 0x0033:
			intVal := int(e.V[x])
			if intVal > 100 {
				e.RAM[e.I] = uint8((intVal % 1000) / 100)
				e.RAM[e.I+1] = uint8((intVal % 100) / 10)
				e.RAM[e.I+2] = uint8(intVal % 10)
			} else if intVal > 10 {
				e.RAM[e.I] = 0
				e.RAM[e.I+1] = uint8((intVal % 100) / 10)
				e.RAM[e.I+2] = uint8(intVal % 10)
			} else {
				e.RAM[e.I] = 0
				e.RAM[e.I+1] = 0
				e.RAM[e.I+2] = uint8(intVal)
			}
			e.PC += 2

		// Fx55 - LD [I], Vx
		// Store registers V0 through Vx in memory starting at location I.
		// The interpreter copies the values of registers V0 through Vx into memory, starting at the address in I.
		case 0x0055:
			var i uint16
			for i = 0; i < x; i++ {
				e.RAM[e.I+i] = e.V[i]
			}
			e.I += x + 1
			e.PC += 2

		// Fx65 - LD Vx, [I]
		// Read registers V0 through Vx from memory starting at location I.
		// The interpreter reads values from memory starting at location I into registers V0 through Vx.
		case 0x0065:
			var i uint16
			for i = 0; i < x; i++ {
				e.V[i] = e.RAM[e.I+i]
			}
			e.I += x + 1
			e.PC += 2
		default:
			log.Fatalf("unknown opcode")
		}

	default:
		log.Fatalf("unrecognised opcode")
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
