package lc3

import (
	"fmt"
	"strconv"
	"strings"
)

func (m *VM) Dump(dumpMemory bool) {
	const instructionsPerLine = 8
	fmt.Printf("Executed:  %d\n", m.instructionsExecuted)

	fmt.Printf("Registers: ")
	fmt.Printf("PC %s ", m.registers[RegPC].AsString())
	fmt.Printf("Flags [%s] ", m.registers[RegCond].FlagsAsString())

	for i := 0; i < 8; i++ {
		fmt.Printf("r%d=%s ", i, m.registers[i].AsString())
	}
	fmt.Printf("\n")

	instruction := m.getCurrentInstruction()
	fmt.Printf("Instruction: %s; %s\n", EncodeInstruction(instruction), instruction.AsString())

	if dumpMemory {
		var address Word
		allowEmpty := false
		emptyStreak := false
		for address = 0; address < 0xffff-instructionsPerLine && int(address) < len(m.memory); address += instructionsPerLine {
			var words []Word
			isEmpty := true
			var j Word
			for j = 0; j < instructionsPerLine; j++ {
				word := m.ReadMem(address + j)
				if word != 0 {
					isEmpty = false
				}
				words = append(words, word)
			}

			if allowEmpty {
				if isEmpty {
					if emptyStreak {
						continue
					}
					emptyStreak = true
					fmt.Printf("   *\n")
					continue
				} else {
					emptyStreak = false
				}
			}

			fmt.Printf("0x%04X  ", address)

			for _, word := range words {
				fmt.Printf("%02X %02X  ", word>>8&0xff, word&0xff)
			}

			for _, word := range words {
				out := func(c byte) {
					if strconv.IsPrint(rune(c)) {
						fmt.Printf("%c", c)
					} else {
						fmt.Printf(".")
					}
				}

				out(byte((word >> 8) & 0xff))
				out(byte(word & 0xff))
			}

			fmt.Printf(" ")

			var asmInst []string
			for _, word := range words {
				asmInst = append(asmInst, EncodeInstruction(word))
			}

			fmt.Printf("%s\n", strings.Join(asmInst, "; "))
			allowEmpty = true
		}
	}
}

func (w Word) AsString() string {
	// type conversion to avoid recursion
	return fmt.Sprintf("%02X %02X(%d)", int((w>>8)&0xff), int(w&0xff), int(w))
}

func (w Word) FlagsAsString() string {
	var builder strings.Builder
	if w&FlN > 0 {
		builder.WriteRune('N')
	} else {
		builder.WriteRune('_')
	}
	if w&FlZ > 0 {
		builder.WriteRune('Z')
	} else {
		builder.WriteRune('_')
	}
	if w&FlP > 0 {
		builder.WriteRune('P')
	} else {
		builder.WriteRune('_')
	}
	return builder.String()
}
