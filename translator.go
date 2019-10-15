package lc3

import (
	"fmt"
	"github.com/pkg/errors"
)

type labelRegistry map[string]Word

type OperandType int

const (
	Register OperandType = iota
	Immediate
	Offset
	String
)

type InstructionSignature struct {
	opcode          string
	operands        []OperandType
	builderFunction interface{}
	writerFunction  func(pass int, labels labelRegistry, m *VM, currentAddress Word, signature InstructionSignature, line Line) (Word, error)
}

var signatures = []InstructionSignature{
	{stropBrn, []OperandType{Offset}, NewBR, brWriterFunction},
	{stropBrz, []OperandType{Offset}, NewBR, brWriterFunction},
	{stropBrp, []OperandType{Offset}, NewBR, brWriterFunction},
	{stropBrnp, []OperandType{Offset}, NewBR, brWriterFunction},
	{stropBrnz, []OperandType{Offset}, NewBR, brWriterFunction},
	{stropBrzp, []OperandType{Offset}, NewBR, brWriterFunction},
	{stropBrnzp, []OperandType{Offset}, NewBR, brWriterFunction},
	{stropAdd, []OperandType{Register, Register, Register}, NewAddRegister, simpleWriterFunction},
	{stropAdd, []OperandType{Register, Register, Immediate}, NewAddImmediate, simpleWriterFunction},
	{stropLd, []OperandType{Register, Offset}, NewLd, simpleWriterFunction},
	{stropSt, []OperandType{Register, Offset}, NewSt, simpleWriterFunction},
	{stropJsr, []OperandType{Offset}, NewJsr, simpleWriterFunction},
	{stropJsr, []OperandType{Register}, NewJsr, simpleWriterFunction},
	{stropJsrr, []OperandType{Register}, NewJsrr, simpleWriterFunction},
	{stropAnd, []OperandType{Register, Register, Immediate}, NewAndImmediate, simpleWriterFunction},
	{stropAnd, []OperandType{Register, Register, Register}, NewAndRegister, simpleWriterFunction},
	{stropLdr, []OperandType{Register, Register, Offset}, NewLdr, simpleWriterFunction},
	{stropStr, []OperandType{Register, Register, Offset}, NewStr, simpleWriterFunction},
	{stropNot, []OperandType{Register, Register}, NewNot, simpleWriterFunction},
	{stropLdi, []OperandType{Register, Offset}, NewLdi, simpleWriterFunction},
	{stropSti, []OperandType{Register, Offset}, NewSti, simpleWriterFunction},
	{stropJmp, []OperandType{Register}, NewJmp, simpleWriterFunction},
	{stropRet, []OperandType{}, NewRet, simpleWriterFunction},
	{stropLea, []OperandType{Register, Offset}, NewLea, simpleWriterFunction},
	{stropGetc, []OperandType{}, NewGetc, simpleWriterFunction},
	{stropOut, []OperandType{}, NewOut, simpleWriterFunction},
	{stropPuts, []OperandType{}, NewPuts, simpleWriterFunction},
	{stropIn, []OperandType{}, NewIn, simpleWriterFunction},
	{stropPutsp, []OperandType{}, NewPutsp, simpleWriterFunction},
	{stropHalt, []OperandType{}, NewHalt, simpleWriterFunction},
	{stropTrap, []OperandType{Offset}, NewTrap, simpleWriterFunction},
	{stropFill, []OperandType{Immediate}, nil, rawWriterFunction},
	{stropFill, []OperandType{Offset}, nil, rawWriterFunction},
	{stropOrig, []OperandType{Immediate}, nil, originWriterFunction},
	{stropStringZ, []OperandType{String}, nil, rawWriterFunction},
}

const (
	pass1 = 1
	pass2 = 2
)

func (lr labelRegistry) setLabelOffset(label string, offset Word) {
	lr[label] = offset
}

func (lr labelRegistry) getLabelOffset(label string) (Word, error) {
	value, ok := lr[label]
	if !ok {
		return value, fmt.Errorf("unknown label %s", label)
	}
	return value, nil
}

func makeRegisterFromString(register string) Word {
	mapping := map[string]Word{
		strReg0: RegR0,
		strReg1: RegR1,
		strReg2: RegR2,
		strReg3: RegR3,
		strReg4: RegR4,
		strReg5: RegR5,
		strReg6: RegR6,
		strReg7: RegR7,
	}
	if ret, ok := mapping[register]; ok {
		return ret
	}
	panic("unknown register " + register)
}

func resolveLabelOrImmediate(currentAddress Word, labels labelRegistry, opType OperandType, operand Operand) (Word, error) {
	var value Word
	var err error

	if opType == Immediate {
		if operand.isNumber() {
			// leave number as number
			value = *operand.number
		} else if operand.isLabel() {
			// absolute address
			value, err = labels.getLabelOffset(*operand.label)
			if err != nil {
				return 0, err
			}
		}
	} else if opType == Offset {
		// calculate relative offset for label
		if operand.isLabel() {
			value, err = labels.getLabelOffset(*operand.label)
			if err != nil {
				return 0, err
			}
			// relative to incremented PC
			value = value - (currentAddress + 1)
		} else if operand.isNumber() {
			value = *operand.number
		} else {
			panic(operand)
		}

	} else {
		panic(opType)
	}

	return value, nil
}

// write instruction to vm's memory
// all instructions except .ORIG, .STRINGZ are handled by this functions
func simpleWriterFunction(pass int, labels labelRegistry, m *VM, currentAddress Word, signature InstructionSignature, line Line) (Word, error) {
	// nothing to do on the first pass
	if pass != pass2 {
		return currentAddress + 1, nil
	}

	var args = make([]Word, len(signature.operands))
	// build instruction arguments
	for i := range signature.operands {
		if line.Operands[i].isRegister() {
			args[i] = makeRegisterFromString(*line.Operands[i].register)
		} else if line.Operands[i].isLabel() || line.Operands[i].isNumber() {
			arg, err := resolveLabelOrImmediate(currentAddress, labels, signature.operands[i], line.Operands[i])
			if err != nil {
				return currentAddress, err
			}
			args[i] = arg
		} else if line.Operands[i].isString() {
			// todo: allow string operands.
			// to allow string operands, we should calculate string length
			// on the first pass and write instruction on second pass
			return 0, errors.Errorf("string operands are not allowed")
		}
	}

	var instruction Word
	var err error
	if len(signature.operands) == 0 {
		instruction = signature.builderFunction.(func() Word)()
	} else if len(signature.operands) == 1 {
		instruction = signature.builderFunction.(func(Word) Word)(args[0])
	} else if len(signature.operands) == 2 {
		instruction = signature.builderFunction.(func(Word, Word) Word)(args[0], args[1])
	} else if len(signature.operands) == 3 {
		instruction = signature.builderFunction.(func(Word, Word, Word) Word)(args[0], args[1], args[2])
	}
	if err != nil {
		return currentAddress, err
	}
	m.WriteMem(currentAddress, instruction)

	return currentAddress + 1, nil
}

func originWriterFunction(pass int, labels labelRegistry, m *VM, currentAddress Word, signature InstructionSignature, line Line) (Word, error) {
	if !line.Operands[0].isNumber() {
		return currentAddress, errors.Errorf("number expected for .ORIG")
	}
	m.SetOrigin(*line.Operands[0].number)
	// .origin resets currentAddress to origin's's absolute value
	return *line.Operands[0].number, nil
}

func brWriterFunction(pass int, labels labelRegistry, m *VM, currentAddress Word, signature InstructionSignature, line Line) (Word, error) {
	if pass != pass2 {
		return currentAddress + 1, nil
	}
	var flags Word
	switch line.Opcode {
	case stropBrn:
		flags |= FlN
	case stropBrz:
		flags |= FlZ
	case stropBrp:
		flags |= FlP
	case stropBrnp:
		flags |= FlN | FlP
	case stropBrnz:
		flags |= FlN | FlZ
	case stropBrzp:
		flags |= FlZ | FlP
	case stropBrnzp:
		flags |= FlN | FlZ | FlP
	}

	if !line.Operands[0].isLabel() && !line.Operands[0].isNumber() {
		return currentAddress, errors.Errorf("label or immediate number expected")
	}

	value, err := resolveLabelOrImmediate(currentAddress, labels, signature.operands[0], line.Operands[0])
	if err != nil {
		return currentAddress, err
	}

	if flags == 0 {
		return 0, errors.Errorf("br with no flags")
	}

	instruction := NewBR(flags, value)
	m.WriteMem(currentAddress, instruction)
	return currentAddress + 1, nil
}

// write raw value. handler for .STRINGZ and .FILL
func rawWriterFunction(pass int, labels labelRegistry, m *VM, currentAddress Word, signature InstructionSignature, line Line) (Word, error) {
	var advancement Word = 0

	if signature.opcode == stropStringZ {
		if !line.Operands[0].isString() {
			return currentAddress, errors.Errorf("string expected for .STRINGZ")
		}

		str := *line.Operands[0].string

		for i := 0; i < len(str); i++ {
			m.WriteMem(currentAddress+advancement, Word(str[i]))
			advancement++
		}

		// termination zero
		m.WriteMem(currentAddress+advancement, 0)
		advancement++

		return currentAddress + advancement, nil
	} else if signature.opcode == stropFill {
		if pass != pass2 {
			return currentAddress + 1, nil
		}
		value, err := resolveLabelOrImmediate(currentAddress, labels, signature.operands[0], line.Operands[0])
		if err != nil {
			return currentAddress, err
		}
		m.WriteMem(currentAddress, value)
		advancement++
	}

	return currentAddress + advancement, nil
}

func assembleVM(lines []Line) (*VM, error) {
	ret := NewVM()
	labels := make(labelRegistry)

	for pass := pass1; pass <= pass2; pass++ {
		var currentAddress Word = 0
		for lineno, line := range lines {
			// save label position
			//fmt.Printf("pass %d line %d\n", pass, lineno)
			if line.Label != "" {
				if line.Label != "" {
					labels.setLabelOffset(line.Label, currentAddress)
				}
			}

			// handle opcode
			if line.Opcode == "" {
				continue
			}

			if line.Opcode == stropEnd {
				break
			}

			var foundSignature = false
			var signature InstructionSignature

			// find matching signature
			for _, signature = range signatures {
				if signature.opcode != line.Opcode {
					continue
				}
				if len(signature.operands) != len(line.Operands) {
					continue
				}
				operandsMatch := true
				for i, lineOp := range line.Operands {
					if signature.operands[i] == Register && !lineOp.isRegister() {
						operandsMatch = false
					}

					if signature.operands[i] == Immediate && !lineOp.isNumber() {
						operandsMatch = false
					}

					if signature.operands[i] == Offset && !(lineOp.isNumber() || lineOp.isLabel()) {
						operandsMatch = false
					}
				}
				if !operandsMatch {
					continue
				}
				foundSignature = true
				break
			}
			if !foundSignature {
				return nil, errors.Errorf("unknown opcode signature at line %d", lineno+1)
			}

			var err error
			currentAddress, err = signature.writerFunction(pass, labels, ret, currentAddress, signature, line)
			if err != nil {
				return nil, errors.Errorf("%s at line %d", err.Error(), lineno+1)
			}
		}
	}

	//for label, address := range labels {
	//	fmt.Printf("%04X %6d %s\n", address, address, label)
	//}

	return ret, nil
}
