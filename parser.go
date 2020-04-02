package lc3

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// http://people.cs.georgetown.edu/~squier/Teaching/HardwareFundamentals/LC3-trunk/docs/LC3-AssemblyManualAndExamples.pdf

const (
	stropBrn   = "BRN"
	stropBrz   = "BRZ"
	stropBrp   = "BRP"
	stropBrzp  = "BRZP"
	stropBrnp  = "BRNP"
	stropBrnz  = "BRNZ"
	stropBrnzp = "BRNZP"
	stropAdd   = "ADD"
	stropLd    = "LD"
	stropSt    = "ST"
	stropJsr   = "JSR"
	stropJsrr  = "JSRR"
	stropAnd   = "AND"
	stropLdr   = "LDR"
	stropStr   = "STR"
	stropRti   = "RTI"
	stropNot   = "NOT"
	stropLdi   = "LDI"
	stropSti   = "STI"
	stropJmp   = "JMP"
	stropRet   = "RET"
	stropRes   = "RES"
	stropLea   = "LEA"
	stropTrap  = "TRAP"

	stropGetc  = "GETC"
	stropOut   = "OUT"
	stropPuts  = "PUTS"
	stropIn    = "IN"
	stropPutsp = "PUTSP"
	stropHalt  = "HALT"

	stropEnd     = ".END"
	stropFill    = ".FILL"
	stropOrig    = ".ORIG"
	stropStringZ = ".STRINGZ"
)

var strOps = []string{stropBrn, stropBrz, stropBrp, stropBrzp, stropBrnp, stropBrnz, stropBrnzp,
	stropAdd, stropLd, stropSt, stropJsr, stropJsrr, stropAnd, stropLdr, stropStr, stropRti,
	stropNot, stropLdi, stropSti, stropJmp, stropRet, stropRes, stropLea, stropTrap,
	stropGetc, stropOut, stropPuts, stropIn, stropPutsp, stropHalt,
	stropEnd, stropFill, stropOrig, stropStringZ}

const (
	strReg0 = "R0"
	strReg1 = "R1"
	strReg2 = "R2"
	strReg3 = "R3"
	strReg4 = "R4"
	strReg5 = "R5"
	strReg6 = "R6"
	strReg7 = "R7"
)

var strRegs = []string{strReg0, strReg1, strReg2, strReg3, strReg4, strReg5, strReg6, strReg7}

type Line struct {
	Label    string
	Opcode   string
	Operands []Operand
	Comment  string
}

func (l *Line) String() string {
	var buffer strings.Builder

	putSpaceBeforeComment := false

	if len(l.Label) > 0 {
		buffer.WriteString(l.Label)
		buffer.WriteByte(' ')
		putSpaceBeforeComment = true
	}

	if len(l.Opcode) > 0 {
		buffer.WriteString(l.Opcode)
		putSpaceBeforeComment = true
	}

	var operands []string
	for _, operand := range l.Operands {
		operands = append(operands, operand.String())
	}

	if len(operands) > 0 {
		buffer.WriteByte(' ')
		buffer.WriteString(strings.Join(operands, ", "))
	}

	if len(l.Comment) > 0 {
		if putSpaceBeforeComment {
			buffer.WriteString(" ")
		}
		buffer.WriteString(";")
		buffer.WriteString(l.Comment)
	}
	return buffer.String()
}

// operand can be one of: label, register, string, number
type Operand struct {
	register *string
	number   *Word
	string   *string
	label    *string
}

func (o *Operand) isRegister() bool { return o.register != nil }
func (o *Operand) isNumber() bool   { return o.number != nil }
func (o *Operand) isString() bool   { return o.string != nil }
func (o *Operand) isLabel() bool    { return o.label != nil }
func (o *Operand) String() string {
	if o.isRegister() {
		return *o.register
	}
	if o.isString() {
		return strconv.Quote(*o.string)
	}
	if o.isLabel() {
		return *o.label
	}
	if o.isNumber() {
		return fmt.Sprintf("x%X", *o.number)
	}

	return ""
}

func isWhitespace(char byte) bool {
	return char == ' ' || char == '\t' || char == '\n'
}

func eatSpaces(line string, pos int) int {
	for pos < len(line) {
		if isWhitespace(line[pos]) {
			pos++
			continue
		}

		break
	}

	return pos
}

// parse identifier
// return identifier and position in original string after
func parseIdentifier(line string, pos int) (string, int) {
	start := pos
	for pos < len(line) && !isWhitespace(line[pos]) && line[pos] != ',' {
		pos++
	}
	return strings.ToUpper(line[start:pos]), pos
}

// if error is not nil, second return value should point to error position
func parseOperand(line string, pos int) (Operand, int, error) {
	// string
	if line[pos] == '"' {
		var buffer strings.Builder
		pos++
		unquoteBuffer := line[pos:]
		for len(unquoteBuffer) > 0 {
			if unquoteBuffer[0] == '"' {
				unquoteBuffer = unquoteBuffer[1:]
				break
			}

			// parse character
			char, _, tail, err := strconv.UnquoteChar(unquoteBuffer, '"')
			if err == strconv.ErrSyntax {
				// parse escape
				if unquoteBuffer[0] == '\\' && len(unquoteBuffer) >= 2 && unquoteBuffer[1] == 'e' {
					char = 0x1B
					err = nil
					tail = unquoteBuffer[2:]
				}
			}
			if err != nil {
				return Operand{}, pos, err
			}

			// append char to result and continue
			buffer.WriteRune(char)
			unquoteBuffer = tail
		}

		parsed := strings.TrimSuffix(line, unquoteBuffer)
		pos = len(parsed)

		ret := buffer.String()
		return Operand{string: &ret}, pos, nil
	}

	// labels, registers as numbers initially could be parsed as identifiers
	identifier, newPos := parseIdentifier(line, pos)
	if isRegister(identifier) {
		return Operand{register: &identifier}, newPos, nil
	}

	base := 0
	neg := 1
	// try to parse identifier as number
	if identifier[0] == '#' {
		// decimal
		base = 10
	} else if identifier[0] == 'X' {
		// hexadecimal
		base = 16
	} else {
		// label
		return Operand{label: &identifier}, newPos, nil
	}

	identifier = identifier[1:]

	if identifier[0] == '-' {
		neg = -1
		identifier = identifier[1:]
	}

	// continue parse as number
	number, err := strconv.ParseUint(identifier, base, 16)
	if err != nil {
		// todo: parse err.(*NumError) ?
		return Operand{}, pos, err
	}

	// strconv.ParseInt checks bit length
	word := Word(int(number) * neg)

	return Operand{number: &word}, newPos, nil
}

// check fif given identifier if opcode
func isOpcode(identifier string) bool {
	for i := range strOps {
		if strOps[i] == identifier {
			return true
		}
	}
	return false
}

func isRegister(identifier string) bool {
	for i := range strRegs {
		if strRegs[i] == identifier {
			return true
		}
	}
	return false
}

func parseInput(reader io.Reader) ([]Line, error) {
	var lines []Line
	var err error

	const (
		ParseLabelAndOpcode = iota
		ParseOpcode
		ParseOperands
	)

	r := bufio.NewReader(reader)
	lineno := 0
	done := false
	for done == false {
		var line string
		lineno++
		line, err = r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				done = true
			} else {
				return nil, errors.Wrapf(err, "line %d: cannot read line", lineno)
			}
		}

		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")

		currentLine := Line{}

		state := ParseLabelAndOpcode

	ParseLine:
		for i := 0; i < len(line); {
			// skip spaces
			i = eatSpaces(line, i)

			// empty line or no characters left
			// this check guarantee that we have an input
			if len(line[i:]) == 0 {
				break
			}

			// parse comment
			if line[i] == ';' {
				// save comment if any
				if len(line) > i {
					currentLine.Comment = line[i+1:]
				}
				// finish line processing
				i = len(line)
				continue
			}

			switch state {
			case ParseLabelAndOpcode:
				var identifier string
				identifier, i = parseIdentifier(line, i)

				if len(identifier) == 0 {
					return nil, errors.Errorf("label or opcode expected at %d:%d", lineno, i)
				}

				// no label on this line
				if isOpcode(identifier) {
					currentLine.Opcode = identifier
					state = ParseOperands
					continue ParseLine
				}

				currentLine.Label = identifier
				state = ParseOpcode
				continue ParseLine

			case ParseOpcode:
				identifier, tmpPos := parseIdentifier(line, i)
				if !isOpcode(identifier) {
					return nil, errors.Errorf("opcode expected at %d:%d", lineno, i)
				}
				i = tmpPos

				currentLine.Opcode = identifier
				state = ParseOperands
				continue ParseLine

			case ParseOperands:
				operand, tmpPos, err := parseOperand(line, i)
				if err != nil {
					return nil, errors.Errorf("%s at %d:%d", err.Error(), lineno, tmpPos)
				}
				i = tmpPos
				currentLine.Operands = append(currentLine.Operands, operand)

				// eat spaces and comma
				i = eatSpaces(line, i)
				if i < len(line) && line[i] == ',' {
					i++
				}

				// do not change state, parse comment or operand again
				continue ParseLine
			}
			panic("unreachable")
			//return nil, errors.Errorf("unknown input at %d: %s", lineno, line[i:])
		}

		lines = append(lines, currentLine)
		if currentLine.Opcode == stropEnd {
			break
		}
	}

	return lines, nil
}

func ParseAssembly(reader io.Reader) (*VM, error) {
	lines, err := parseInput(reader)
	if err != nil {
		return nil, err
	}

	m, err := assembleVM(lines)
	//if err == nil {
	//	m.Dump(true)
	//}
	return m, err
}
