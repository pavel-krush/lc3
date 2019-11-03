package lc3

import (
	"errors"
	"math"
)

// CPU word size
type Word uint16

const WordMax = math.MaxUint16
const MrKbsr = 0xFE00 // keyboard status
const MrKbdr = 0xFE02 // keyboard data

const (
	TrapVectGetc  = 0x20
	TrapVectOut   = 0x21
	TrapVectPuts  = 0x22
	TrapVectIn    = 0x23
	TrapVectPutsp = 0x24
	TrapVectHalt  = 0x25
)

const IOChannelsBufferSize = 64

// registers
const (
	RegR0 = iota // general purpose registers
	RegR1
	RegR2
	RegR3
	RegR4
	RegR5
	RegR6
	RegR7
	RegPC   // program counter
	RegCond // flags register
)

// flags
const (
	FlP Word = 1 << iota // result is positive
	FlZ                  // result is zero
	FlN                  // result is negative
)

// opcodes
const (
	OpBr   = iota // branch
	OpAdd         // add
	OpLd          // load
	OpSt          // store
	OpJsr         // jump register
	OpAnd         // bitwise and
	OpLdr         // load register
	OpStr         // store register
	OpRti         // todo
	OpNot         // bitwise not
	OpLdi         // load indirect
	OpSti         // store indirect
	OpJmp         // jump
	OpRes         // reserved
	OpLea         // load effective address
	OpTrap        // execute trap
)

var ErrBadInstruction = errors.New("bad instruction")
var ErrNotRunning = errors.New("vm is not running")

type VM struct {
	memory               [math.MaxUint16]Word
	registers            [10]Word
	running              bool
	instructionsExecuted uint

	Stdin  chan Word
	Stdout chan Word
}

func NewVM() *VM {
	ret := &VM{
		Stdin:  make(chan Word, IOChannelsBufferSize),
		Stdout: make(chan Word, IOChannelsBufferSize),
	}
	return ret
}

func (m *VM) setFlags(result Word) {
	if result == 0 {
		m.registers[RegCond] = FlZ
		return
	}
	if IsNegative(result) {
		m.registers[RegCond] = FlN
	} else {
		m.registers[RegCond] = FlP
	}
}

func (m *VM) Start() {
	m.running = true
}

func (m *VM) Stop() {
	m.running = false
}

func (m *VM) getCurrentInstruction() Word {
	return m.ReadMem(m.registers[RegPC])
}

func (m *VM) Step() error {
	if !m.running {
		return ErrNotRunning
	}

	m.instructionsExecuted++

	instruction := m.getCurrentInstruction()

	// advance PC immediately all relative PC use incremented PC
	m.registers[RegPC]++

	opcode := getOpcode(instruction)
	switch opcode {
	case OpBr:
		// | 15 | 14 | 13 | 12 | 11 | 10 |  9 |  8 |  7 |  6 |  5 |  4 |  3 |  2 |  1 |  0 |
		// |  0    0    0    0 |  n |  z |  p |                  PCOffset9                 |
		instructionFlags := getNBits(instruction, 9, 3)

		nCond := instructionFlags & m.registers[RegCond] & FlN
		zCond := instructionFlags & m.registers[RegCond] & FlZ
		pCond := instructionFlags & m.registers[RegCond] & FlP

		if pCond|zCond|nCond > 0 {
			m.registers[RegPC] = m.registers[RegPC] + getNBitsExtended(instruction, 0, 9)
		}
	case OpAdd:
		dr := getNBits(instruction, 9, 3)
		sr1 := getNBits(instruction, 6, 3)
		if getNBits(instruction, 5, 1) == 0 {
			// | 15 | 14 | 13 | 12 | 11 | 10 |  9 |  8 |  7 |  6 |  5 |  4 |  3 |  2 |  1 |  0 |
			// |  0    0    0    1 |      DR      |      SR1     |  0 |  0    0 |      SR2     |
			sr2 := getNBits(instruction, 0, 3)
			m.registers[dr] = m.registers[sr1] + m.registers[sr2]

		} else {
			// | 15 | 14 | 13 | 12 | 11 | 10 |  9 |  8 |  7 |  6 |  5 |  4 |  3 |  2 |  1 |  0 |
			// |  0    0    0    1 |      DR      |      SR      |  1 |         Imm5           |
			imm := getNBitsExtended(instruction, 0, 5)
			m.registers[dr] = m.registers[sr1] + imm
		}
		m.setFlags(m.registers[dr])
	case OpLd:
		// | 15 | 14 | 13 | 12 | 11 | 10 |  9 |  8 |  7 |  6 |  5 |  4 |  3 |  2 |  1 |  0 |
		// |  0    0    1    0 |      DR      |               PCOffset9                    |
		dr := getNBits(instruction, 9, 3)
		m.registers[dr] = m.ReadMem(m.registers[RegPC] + getNBitsExtended(instruction, 0, 9))
		m.setFlags(m.registers[dr])
	case OpSt:
		// | 15 | 14 | 13 | 12 | 11 | 10 |  9 |  8 |  7 |  6 |  5 |  4 |  3 |  2 |  1 |  0 |
		// |  0    0    1    1 |      SR      |               PCOffset9                    |
		sr := getNBits(instruction, 9, 3)
		offset := getNBitsExtended(instruction, 0, 9)
		m.WriteMem(m.registers[RegPC]+offset, m.registers[sr])
	case OpJsr:
		m.registers[RegR7] = m.registers[RegPC]
		if getNBits(instruction, 11, 1) == 1 {
			// | 15 | 14 | 13 | 12 | 11 | 10 |  9 |  8 |  7 |  6 |  5 |  4 |  3 |  2 |  1 |  0 |
			// |  0    1    0    0 |  1 |                    PCOffset11                        |
			m.registers[RegPC] = m.registers[RegPC] + getNBitsExtended(instruction, 0, 11)
		} else {
			// | 15 | 14 | 13 | 12 | 11 | 10 |  9 |  8 |  7 |  6 |  5 |  4 |  3 |  2 |  1 |  0 |
			// |  0    1    0    0 |  0 |  0    0 |     BaseR    |  0    0    0    0    0    0 |
			baseR := getNBits(instruction, 6, 3)
			m.registers[RegPC] = m.registers[baseR]
		}
	case OpAnd:
		dr := getNBits(instruction, 9, 3)
		sr1 := getNBits(instruction, 6, 3)
		if getNBits(instruction, 5, 1) == 0 {
			// | 15 | 14 | 13 | 12 | 11 | 10 |  9 |  8 |  7 |  6 |  5 |  4 |  3 |  2 |  1 |  0 |
			// |  0    1    0    1 |      DR      |      SR1     |  0 |  0    0 |      SR2     |
			sr2 := getNBits(instruction, 0, 3)
			m.registers[dr] = m.registers[sr1] & m.registers[sr2]

		} else {
			// | 15 | 14 | 13 | 12 | 11 | 10 |  9 |  8 |  7 |  6 |  5 |  4 |  3 |  2 |  1 |  0 |
			// |  0    1    0    1 |      DR      |      SR      |  1 |         Imm5           |
			imm := getNBitsExtended(instruction, 0, 5)
			m.registers[dr] = m.registers[sr1] & imm
		}
		m.setFlags(m.registers[dr])
	case OpLdr:
		// | 15 | 14 | 13 | 12 | 11 | 10 |  9 |  8 |  7 |  6 |  5 |  4 |  3 |  2 |  1 |  0 |
		// |  0    1    1    0 |      DR      |     BaseR    |          PCOffset6          |
		dr := getNBits(instruction, 9, 3)
		baseR := getNBits(instruction, 6, 3)
		offset := getNBitsExtended(instruction, 0, 6)
		m.registers[dr] = m.ReadMem(m.registers[baseR] + offset)
		m.setFlags(m.registers[dr])
	case OpStr:
		// | 15 | 14 | 13 | 12 | 11 | 10 |  9 |  8 |  7 |  6 |  5 |  4 |  3 |  2 |  1 |  0 |
		// |  0    1    1    1 |      SR      |     BaseR    |           offset6           |
		sr := getNBits(instruction, 9, 3)
		baseR := getNBits(instruction, 6, 3)
		offset := getNBitsExtended(instruction, 0, 6)
		m.WriteMem(m.registers[baseR]+offset, m.registers[sr])
	case OpRti:
		return ErrBadInstruction
	case OpNot:
		// | 15 | 14 | 13 | 12 | 11 | 10 |  9 |  8 |  7 |  6 |  5 |  4 |  3 |  2 |  1 |  0 |
		// |  1    0    0    1 |      DR      |       SR     |  1    1    1    1    1    1 |
		dr := getNBits(instruction, 9, 3)
		sr := getNBits(instruction, 6, 3)
		m.registers[dr] = ^m.registers[sr]
		m.setFlags(m.registers[dr])
	case OpLdi:
		// | 15 | 14 | 13 | 12 | 11 | 10 |  9 |  8 |  7 |  6 |  5 |  4 |  3 |  2 |  1 |  0 |
		// |  1    0    1    0 |      DR      |                 PCOffset9                  |
		dr := getNBits(instruction, 9, 3)
		m.registers[dr] = m.ReadMem(m.ReadMem(m.registers[RegPC] + getNBitsExtended(instruction, 0, 9)))
		m.setFlags(m.registers[dr])
	case OpSti:
		// | 15 | 14 | 13 | 12 | 11 | 10 |  9 |  8 |  7 |  6 |  5 |  4 |  3 |  2 |  1 |  0 |
		// |  1    0    1    1 |      SR      |                 PCOffset9                  |
		sr := getNBits(instruction, 9, 3)
		m.WriteMem(m.ReadMem(m.registers[RegPC]+getNBitsExtended(instruction, 0, 9)), m.registers[sr])
	case OpJmp:
		// | 15 | 14 | 13 | 12 | 11 | 10 |  9 |  8 |  7 |  6 |  5 |  4 |  3 |  2 |  1 |  0 |
		// |  1    1    0    0 |  0    0    0 |     BaseR    |  0    0    0    0    0    0 |
		baseR := getNBits(instruction, 6, 3)
		m.registers[RegPC] = m.registers[baseR]
	case OpRes:
		return ErrBadInstruction
	case OpLea:
		// | 15 | 14 | 13 | 12 | 11 | 10 |  9 |  8 |  7 |  6 |  5 |  4 |  3 |  2 |  1 |  0 |
		// |  1    1    1    0 |      DR      |                 PCOffset9                  |
		dr := getNBits(instruction, 9, 3)
		pcOffset9 := getNBitsExtended(instruction, 0, 9)
		m.registers[dr] = m.registers[RegPC] + pcOffset9
		m.setFlags(m.registers[dr])
	case OpTrap:
		// | 15 | 14 | 13 | 12 | 11 | 10 |  9 |  8 |  7 |  6 |  5 |  4 |  3 |  2 |  1 |  0 |
		// |  1    1    1    1 |  0    0    0    0 |              trapvect8                |
		vector := getNBitsExtended(instruction, 0, 8)
		switch vector {
		case TrapVectGetc:
			//ch := <-m.Stdin
			//m.registers[RegR0] = ch
		case TrapVectOut:
			//fmt.Printf("%c", m.registers[RegR0] & 0xff)
		case TrapVectPuts:
			/*ptr := m.registers[RegR0]
			for {
				word := m.ReadMem(ptr)
				char := byte(word & 0xff)
				if char == 0 {
					break
				}
				fmt.Printf("%c", char)
				ptr++
			}*/
		case TrapVectIn:
		case TrapVectPutsp:
		case TrapVectHalt:
			m.Stop()
		default:
			m.registers[RegR7] = m.registers[RegPC]
			m.registers[RegPC] = m.ReadMem(vector)
		}
	}

	return nil
}

func (m *VM) WriteMem(address Word, value Word) {
	m.memory[address] = value
}

func (m *VM) ReadMem(address Word) Word {
	if address == MrKbsr {
		if len(m.Stdin) > 0 {
			m.memory[MrKbsr] = 1 << 15
			m.memory[MrKbdr] = <-m.Stdin
		} else {
			m.memory[MrKbsr] = 0
		}
	}

	return m.memory[address]
}

func (m *VM) SetOrigin(origin Word) {
	if m.running {
		return
	}
	m.registers[RegPC] = origin
}

func (m *VM) GetInstructionsExecuted() uint {
	return m.instructionsExecuted
}

func (m *VM) IsRunning() bool {
	return m.running
}

func (m *VM) GetRegister(register int) Word {
	return m.registers[register]
}
