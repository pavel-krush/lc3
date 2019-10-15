package lc3

import (
	"fmt"
	"strings"
)

type NumericLiteralMode int

const (
	ModeHex NumericLiteralMode = iota
	ModeDec
)

var encodingLiteralMode = ModeHex

var instructionEncoders = map[Word]func(Word) string{
	OpBr:   encodeBR,
	OpAdd:  encodeAdd,
	OpLd:   encodeLd,
	OpSt:   encodeSt,
	OpJsr:  encodeJsr,
	OpAnd:  encodeAnd,
	OpLdr:  encodeLdr,
	OpStr:  encodeStr,
	OpRti:  encodeRti,
	OpNot:  encodeNot,
	OpLdi:  encodeLdi,
	OpSti:  encodeSti,
	OpJmp:  encodeJmp,
	OpRes:  encodeRes,
	OpLea:  encodeLea,
	OpTrap: encodeTrap,
}

func SetEncodeNumericLiteralMode(mode NumericLiteralMode) {
	encodingLiteralMode = mode
}

func EncodeInstruction(instruction Word) string {
	opcode := getOpcode(instruction)
	return instructionEncoders[opcode](instruction)
}

func encodeNumericLiteral(value Word) string {
	if encodingLiteralMode == ModeHex {
		return fmt.Sprintf("x%x", value)
	} else if encodingLiteralMode == ModeDec {
		return fmt.Sprintf("#%d", value)
	}
	panic(encodingLiteralMode)
}

func encodeRegister(register Word) string {
	return fmt.Sprintf("R%d", register)
}

func encodeBR(instruction Word) string {
	var flagsMapping = map[Word]string{
		/* nzp */
		/* 000 */ 0: "BR",
		/* 001 */ 1: "BRP",
		/* 010 */ 2: "BRZ",
		/* 011 */ 3: "BRZP",
		/* 100 */ 4: "BRN",
		/* 101 */ 5: "BRNP",
		/* 110 */ 6: "BRNZ",
		/* 111 */ 7: "BRNZP",
	}
	return fmt.Sprintf("%s %s", flagsMapping[getNBits(instruction, 9, 3)], encodeNumericLiteral(getNBitsExtended(instruction, 0, 9)))
}

func encodeAdd(instruction Word) string {
	var ret strings.Builder
	ret.WriteString("ADD ")
	if getNBits(instruction, 5, 1) == 0 {
		ret.WriteString(encodeRegister(getNBits(instruction, 9, 3)))
		ret.WriteString(", ")
		ret.WriteString(encodeRegister(getNBits(instruction, 6, 3)))
		ret.WriteString(", ")
		ret.WriteString(encodeRegister(getNBits(instruction, 0, 3)))
	} else {
		ret.WriteString(encodeRegister(getNBits(instruction, 9, 3)))
		ret.WriteString(", ")
		ret.WriteString(encodeRegister(getNBits(instruction, 6, 3)))
		ret.WriteString(", ")
		ret.WriteString(encodeNumericLiteral(getNBitsExtended(instruction, 0, 5)))
	}
	return ret.String()
}

func encodeLd(instruction Word) string {
	var ret strings.Builder
	ret.WriteString("LD ")
	ret.WriteString(encodeRegister(getNBits(instruction, 9, 3)))
	ret.WriteString(", ")
	ret.WriteString(encodeNumericLiteral(getNBitsExtended(instruction, 0, 9)))
	return ret.String()
}

func encodeSt(instruction Word) string {
	var ret strings.Builder
	ret.WriteString("ST ")
	ret.WriteString(encodeNumericLiteral(getNBitsExtended(instruction, 0, 9)))
	return ret.String()
}

func encodeJsr(instruction Word) string {
	var ret strings.Builder

	if getNBits(instruction, 11, 1) == 1 {
		ret.WriteString("JSR ")
		ret.WriteString(encodeNumericLiteral(getNBitsExtended(instruction, 0, 11)))
	} else {
		ret.WriteString("JSRR ")
		ret.WriteString(encodeRegister(getNBits(instruction, 6, 3)))
	}

	return ret.String()
}

func encodeAnd(instruction Word) string {
	var ret strings.Builder

	ret.WriteString("AND ")

	ret.WriteString(encodeRegister(getNBits(instruction, 9, 3)))
	ret.WriteString(", ")
	ret.WriteString(encodeRegister(getNBits(instruction, 6, 3)))
	if getNBits(instruction, 5, 1) == 0 {
		ret.WriteString(", ")
		ret.WriteString(encodeRegister(getNBits(instruction, 0, 3)))
	} else {
		ret.WriteString(", ")
		ret.WriteString(encodeNumericLiteral(getNBits(instruction, 0, 5)))
	}
	return ret.String()
}

func encodeLdr(instruction Word) string {
	var ret strings.Builder
	ret.WriteString("LDR ")
	ret.WriteString(encodeRegister(getNBits(instruction, 9, 3)))
	ret.WriteString(", ")
	ret.WriteString(encodeRegister(getNBits(instruction, 6, 3)))
	ret.WriteString(", ")
	ret.WriteString(encodeNumericLiteral(getNBitsExtended(instruction, 0, 6)))
	return ret.String()
}

func encodeStr(instruction Word) string {
	var ret strings.Builder
	ret.WriteString("STR ")
	ret.WriteString(encodeRegister(getNBits(instruction, 9, 3)))
	ret.WriteString(", ")
	ret.WriteString(encodeRegister(getNBits(instruction, 6, 3)))
	ret.WriteString(", ")
	ret.WriteString(encodeNumericLiteral(getNBitsExtended(instruction, 0, 6)))
	return ret.String()
}

func encodeRti(instruction Word) string {
	return ""
}

func encodeNot(instruction Word) string {
	var ret strings.Builder

	ret.WriteString("NOT ")
	ret.WriteString(encodeRegister(getNBits(instruction, 9, 3)))
	ret.WriteString(", ")
	ret.WriteString(encodeRegister(getNBits(instruction, 6, 3)))

	return ret.String()
}

func encodeLdi(instruction Word) string {
	var ret strings.Builder

	ret.WriteString("LDI ")
	ret.WriteString(encodeRegister(getNBits(instruction, 9, 3)))
	ret.WriteString(", ")
	ret.WriteString(encodeNumericLiteral(getNBitsExtended(instruction, 0, 9)))

	return ret.String()
}

func encodeSti(instruction Word) string {
	var ret strings.Builder

	ret.WriteString("STI ")
	ret.WriteString(encodeRegister(getNBits(instruction, 9, 3)))
	ret.WriteString(", ")
	ret.WriteString(encodeNumericLiteral(getNBitsExtended(instruction, 0, 9)))

	return ret.String()
}

func encodeJmp(instruction Word) string {
	var ret strings.Builder

	register := getNBits(instruction, 6, 3)

	if register == RegR7 {
		ret.WriteString("RET")
	} else {
		ret.WriteString("JMP ")
		ret.WriteString(encodeRegister(register))
	}

	return ret.String()
}

func encodeRes(instruction Word) string {
	return "RES"
}

func encodeLea(instruction Word) string {
	var ret strings.Builder

	ret.WriteString("LEA ")
	ret.WriteString(encodeRegister(getNBits(instruction, 9, 3)))
	ret.WriteString(", ")
	ret.WriteString(encodeNumericLiteral(getNBitsExtended(instruction, 0, 9)))

	return ret.String()
}

func encodeTrap(instruction Word) string {
	var ret strings.Builder

	vector := getNBitsExtended(instruction, 0, 8)
	switch vector {
	case TrapVectGetc:
		return "GETC"
	case TrapVectOut:
		return "OUT"
	case TrapVectPuts:
		return "PUTS"
	case TrapVectIn:
		return "IN"
	case TrapVectHalt:
		return "HALT"
	}

	ret.WriteString("TRAP ")
	ret.WriteString(encodeNumericLiteral(vector))

	return ret.String()
}
