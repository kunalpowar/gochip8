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

	// Display holds the data abotu current display state
	Display [32]uint64

	keys [16]uint8

	// UpdateDisplay is set if display frame needs to be updated.
	UpdateDisplay bool
}

// New returns a new instance of emulator ready to load programs
func New() *Emulator {
	e := Emulator{pc: programStart}
	for i, b := range fontSet {
		e.ram[interpreterStart+i] = b
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
		e.ram[programStart+i] = b
	}

	log.Printf("emulator: loaded %d bytes of rom into memory", len(bs))
}

func (e *Emulator) updateTimers() {
	if e.dt > 0 {
		e.dt--
	}

	if e.st > 0 {
		print("beep..")
		e.st--
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

// EmulateCycle runs the next opcode and updates the timers accordingly
func (e *Emulator) EmulateCycle() {
	var opcode uint16
	opcode = uint16(e.ram[e.pc])<<8 | uint16(e.ram[e.pc+1])

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

	log.Printf("opcode: %04x, PC: %04x", opcode, e.pc)

	switch opcode & 0xf000 {
	case 0x0000:
		switch kk {
		case 0x00e0:
			log.Print("clear display")
			e.clearDisplay()
			e.pc += 2
		case 0x00EE:
			log.Print("Return from a subroutine")
			e.pc = e.stack[e.sp]
			e.sp--
		}
	case 0x1000:
		log.Printf("Jump to location nnn:%x.", nnn)
		e.pc = nnn
	case 0x2000:
		log.Printf("Call subroutine at nnn:%x", nnn)
		e.sp++
		e.stack[e.sp] = e.pc + 2
		e.pc = nnn
	case 0x3000:
		log.Printf("Skip next instruction if Vx(%x) = kk(%x).", e.v[x], kk)
		if uint16(e.v[x]) == kk {
			e.pc += 4
		} else {
			e.pc += 2
		}
	case 0x4000:
		log.Printf("Skip next instruction if Vx(%x) ! kk(%x).", e.v[x], kk)
		if uint16(e.v[x]) != kk {
			e.pc += 4
		} else {
			e.pc += 2
		}
	case 0x5000:
		log.Printf("Skip next instruction if Vx(%x) = Vy(%x).", e.v[x], e.v[y])
		if e.v[x] == e.v[y] {
			e.pc += 4
		} else {
			e.pc += 2
		}
	case 0x6000:
		log.Printf("Set Vx(%x) = kk(%x).", e.v[x], kk)
		e.v[x] = uint8(kk)
		log.Printf("e.V[x]: %d %04x", e.v[x], e.v[x])
		e.pc += 2
	case 0x7000:
		log.Printf("Set Vx (%d) = Vx + kk(%d).", e.v[x], kk)
		e.v[x] += uint8(kk)
		log.Printf("e.V[x]: %d %04x", e.v[x], e.v[x])
		e.pc += 2
	case 0x8000:
		switch n {
		case 0x0000:
			log.Print("Set Vx = Vy.")
			e.v[x] = e.v[y]
			log.Printf("e.V[x]: %d %04x", e.v[x], e.v[x])
		case 0x0001:
			log.Print("Set Vx = Vx OR Vy.")
			e.v[x] |= e.v[y]
		case 0x0002:
			log.Print("Set Vx = Vx AND Vy.")
			e.v[x] &= e.v[y]
		case 0x0003:
			log.Print("Set Vx = Vx XOR Vy.")
			e.v[x] ^= e.v[y]
		case 0x0004:
			log.Print("Set Vx = Vx + Vy, set VF = carry")
			if int(e.v[x])+int(e.v[y]) > 255 {
				e.v[0xF] = 1
			} else {
				e.v[0xF] = 0
			}
			e.v[x] += e.v[y]
			log.Printf("e.V[x]: %d %04x", e.v[x], e.v[x])
		case 0x0005:
			log.Printf("Set Vx = Vx - Vy, set VF = NOT borrow.")
			if e.v[x] > e.v[y] {
				e.v[0xF] = 1
			} else {
				e.v[0xF] = 0
			}
			e.v[x] -= e.v[y]
			log.Printf("e.V[x]: %d %04x", e.v[x], e.v[x])
		case 0x0006:
			log.Print("Set Vx = Vx SHR 1.")
			e.v[0xF] = e.v[x] & 0x1
			e.v[x] >>= 1
			log.Printf("e.V[x]: %d %04x", e.v[x], e.v[x])
		case 0x0007:
			log.Print("Set Vx = Vy - Vx, set VF = NOT borrow.")
			if e.v[y] > e.v[x] {
				e.v[0xF] = 1
			} else {
				e.v[0xF] = 0
			}
			e.v[x] = e.v[y] - e.v[x]
			log.Printf("e.V[x]: %d %04x", e.v[x], e.v[x])
		case 0x000E:
			log.Print("Set Vx = Vx SHL 1.")
			e.v[0xF] = (e.v[x] >> 7) & 0x1
			e.v[x] <<= 1
			log.Printf("e.V[x]: %d %04x", e.v[x], e.v[x])
		}
		e.pc += 2
	case 0x9000:
		log.Print("Skip next instruction if Vx != Vy.")
		if n == 0x0 {
			if e.v[x] != e.v[y] {
				e.pc += 4
			} else {
				e.pc += 2
			}
		}
	case 0xa000:
		log.Print("Set I = nnn.")
		e.i = nnn
		e.pc += 2
	case 0xb000:
		log.Print("Jump to location nnn + V0")
		e.pc = nnn + uint16(e.v[0])
	case 0xc000:
		log.Print("Set Vx = random byte AND kk.")
		e.v[x] = uint8(rand.Intn(255)) & uint8(kk)
		log.Printf("e.V[x]: %d %04x", e.v[x], e.v[x])
		e.pc += 2
	case 0xd000:
		log.Printf("Display n(%d)-byte sprite starting at memory location I(%d) at (Vx:%d, Vy:%d), set VF = collision.", n, e.i, e.v[x], e.v[y])

		e.v[0xF] = 0
		for row := 0; row < int(n); row++ {
			y := int(e.v[y]) + row
			bt := e.ram[int(e.i)+row]
			log.Printf("byte: %08b", bt)

			for col := 0; col < 8; col++ {
				x := int(e.v[x]) + col

				pix := e.getPixel(x, y)
				if bt&(0x80>>col) != 0 {
					e.togglePixel(x, y)
					if pix != 0 {
						e.v[0xF] = 1
					}
				}
			}
		}
		e.UpdateDisplay = true
		e.pc += 2
	case 0xe000:
		switch kk {
		case 0x009e:
			log.Print("Skip next instruction if key with the value of Vx is pressed.")
			if e.keys[e.v[x]] > 0 {
				e.pc += 4
			} else {
				e.pc += 2
			}
		case 0x00a1:
			log.Print("Skip next instruction if key with the value of Vx is not pressed.")
			if e.keys[e.v[x]] == 0 {
				e.pc += 4
			} else {
				e.pc += 2
			}
		}
	case 0xf000:
		switch kk {
		// Fx07 - LD Vx, DT
		// Set Vx = delay timer value.
		// The value of DT is placed into Vx.
		case 0x0007:
			log.Print("Set Vx = delay timer value.")
			e.v[x] = e.dt
			log.Printf("e.V[x]: %d %04x", e.v[x], e.v[x])
			e.pc += 2

		// Fx0A - LD Vx, K
		// Wait for a key press, store the value of the key in Vx.
		// All execution stops until a key is pressed, then the value of that key is stored in Vx.
		case 0x000a:
			log.Print("Wait for a key press, store the value of the key in Vx.")
			ticker := time.NewTicker(10 * time.Millisecond)
			for range ticker.C {
				keyPressed := -1
				for i, k := range e.keys {
					if k > 0 {
						log.Printf("got key press: %d", i)
						keyPressed = i
						break
					}
				}

				if keyPressed == -1 {
					continue
				}

				e.v[x] = uint8(keyPressed)
				log.Printf("e.V[x]: %d %04x", e.v[x], e.v[x])
				ticker.Stop()
				break
			}

			e.pc += 2

		// Fx15 - LD DT, Vx
		// Set delay timer = Vx.
		// DT is set equal to the value of Vx.
		case 0x0015:
			e.dt = e.v[x]
			e.pc += 2

		// Fx18 - LD ST, Vx
		// Set sound timer = Vx.
		// ST is set equal to the value of Vx.
		case 0x0018:
			e.st = e.v[x]
			e.pc += 2

		// Fx1E - ADD I, Vx
		// Set I = I + Vx.
		// The values of I and Vx are added, and the results are stored in I.
		case 0x001e:
			if int(e.i)+int(e.v[x]) > 255 {
				e.v[0xf] = 1
			} else {
				e.v[0xf] = 0
			}
			e.i += uint16(e.v[x])
			e.pc += 2

		// Fx29 - LD F, Vx
		// Set I = location of sprite for digit Vx.
		// The value of I is set to the location for the hexadecimal sprite corresponding to the value of Vx. See section 2.4, Display, for more information on the Chip-8 hexadecimal font.
		case 0x0029:
			e.i = uint16(bytesPerCharacter) * uint16(e.v[x])
			e.pc += 2

		// Fx33 - LD B, Vx
		// Store BCD representation of Vx in memory locations I, I+1, and I+2.
		// The interpreter takes the decimal value of Vx, and places the hundreds digit in memory at location in I, the tens digit at location I+1, and the ones digit at location I+2.
		case 0x0033:
			intVal := int(e.v[x])
			if intVal > 100 {
				e.ram[e.i] = uint8((intVal % 1000) / 100)
				e.ram[e.i+1] = uint8((intVal % 100) / 10)
				e.ram[e.i+2] = uint8(intVal % 10)
			} else if intVal > 10 {
				e.ram[e.i] = 0
				e.ram[e.i+1] = uint8((intVal % 100) / 10)
				e.ram[e.i+2] = uint8(intVal % 10)
			} else {
				e.ram[e.i] = 0
				e.ram[e.i+1] = 0
				e.ram[e.i+2] = uint8(intVal)
			}
			e.pc += 2

		// Fx55 - LD [I], Vx
		// Store registers V0 through Vx in memory starting at location I.
		// The interpreter copies the values of registers V0 through Vx into memory, starting at the address in I.
		case 0x0055:
			var i uint16
			for i = 0; i < x; i++ {
				e.ram[e.i+i] = e.v[i]
			}
			e.i += x + 1
			e.pc += 2

		// Fx65 - LD Vx, [I]
		// Read registers V0 through Vx from memory starting at location I.
		// The interpreter reads values from memory starting at location I into registers V0 through Vx.
		case 0x0065:
			var i uint16
			for i = 0; i < x; i++ {
				e.v[i] = e.ram[e.i+i]
			}
			e.i += x + 1
			e.pc += 2
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
