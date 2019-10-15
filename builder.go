package lc3

// TODO: check register size? make "ret |= reg & 7"

func signCompress(x Word, bitsCount Word) Word {
	negative := IsNegative(x) // save negative flag
	x &= 1<<bitsCount - 1     // remove non significant bits
	if negative {
		x |= 1 << (bitsCount - 1) // restore negative flag
	}
	return x
}

func NewBR(flags Word, pcOffset9 Word) Word {
	if flags == 0 {
		panic("impossible flags: 000")
	}

	var ret Word = OpBr

	ret <<= 3
	ret |= signCompress(flags, 3)

	ret <<= 9
	ret |= signCompress(pcOffset9, 9)

	return ret
}

func NewAddRegister(dr, sr1, sr2 Word) Word {
	var ret Word = OpAdd

	ret <<= 3
	ret |= dr

	ret <<= 3
	ret |= sr1

	ret <<= 1 // type = 0

	ret <<= 2 // empty 2 bits

	ret <<= 3
	ret |= sr2

	return ret
}

func NewAddImmediate(dr, sr, imm5 Word) Word {
	var ret Word = OpAdd

	ret <<= 3
	ret |= dr

	ret <<= 3
	ret |= sr

	ret <<= 1 // type = 1
	ret |= 1

	ret <<= 5
	ret |= signCompress(imm5, 5)

	return ret
}

func NewLd(dr, pcOffset9 Word) Word {
	var ret Word = OpLd

	ret <<= 3
	ret |= dr

	ret <<= 9 // value
	ret |= signCompress(pcOffset9, 9)
	return ret
}

func NewSt(sr, pcOffset9 Word) Word {
	var ret Word = OpSt // opcode

	ret <<= 3
	ret |= sr

	ret <<= 9
	ret |= signCompress(pcOffset9, 9)

	return ret
}

func NewJsr(pcOffset11 Word) Word {
	var ret Word = OpJsr // opcode

	ret <<= 1 // type = 1
	ret |= 1

	ret <<= 11
	ret |= signCompress(pcOffset11, 11)

	return ret
}

func NewJsrr(baseR Word) Word {
	var ret Word = OpJsr // opcode

	ret <<= 1 // type = 0

	ret <<= 2 // 2 empty bits

	ret <<= 3
	ret |= baseR

	ret <<= 6 // empty bits
	return ret
}

func NewAndImmediate(dr, sr, imm5 Word) Word {
	var ret Word = OpAnd // opcode

	ret <<= 3
	ret |= dr

	ret <<= 3
	ret |= sr

	ret <<= 1 // type = 1
	ret |= 1

	ret <<= 5
	ret |= signCompress(imm5, 5)

	return ret
}

func NewAndRegister(dr, sr1, sr2 Word) Word {
	var ret Word = OpAnd

	ret <<= 3
	ret |= dr

	ret <<= 3
	ret |= sr1

	ret <<= 1 // type = 0

	ret <<= 2 // empty 2 bits

	ret <<= 3
	ret |= sr2

	return ret
}

func NewLdr(dr, baseR, pcOffset6 Word) Word {
	var ret Word = OpLdr

	ret <<= 3
	ret |= dr

	ret <<= 3
	ret |= baseR

	ret <<= 6
	ret |= signCompress(pcOffset6, 6)

	return ret
}

func NewStr(sr, baseR, offset6 Word) Word {
	var ret Word = OpStr

	ret <<= 3
	ret |= sr

	ret <<= 3
	ret |= baseR

	ret <<= 6
	ret |= signCompress(offset6, 6)

	return ret
}

func NewNot(dr, sr Word) Word {
	var ret Word = OpNot

	ret <<= 3
	ret |= dr

	ret <<= 3
	ret |= sr

	ret <<= 6 // 6 ones
	ret |= 1<<6 - 1

	return ret
}

func NewLdi(dr, pcOffset9 Word) Word {
	var ret Word = OpLdi

	ret <<= 3
	ret |= dr

	ret <<= 9
	ret |= signCompress(pcOffset9, 9)

	return ret
}

func NewSti(sr, pcOffset9 Word) Word {
	var ret Word = OpSti

	ret <<= 3
	ret |= sr

	ret <<= 9
	ret |= signCompress(pcOffset9, 9)

	return ret
}

func NewJmp(baseR Word) Word {
	var ret Word = OpJmp

	ret <<= 3 // empty bits

	ret <<= 3
	ret |= baseR

	ret <<= 6 // empty bits

	return ret
}

func NewRet() Word {
	return NewJmp(RegR7)
}

func NewLea(dr, pcOffset9 Word) Word {
	var ret Word = OpLea

	ret <<= 3
	ret |= dr

	ret <<= 9
	ret |= signCompress(pcOffset9, 9)

	return ret
}

func NewTrap(trapvect8 Word) Word {
	var ret Word = OpTrap

	ret <<= 4 // empty bits

	ret <<= 8

	ret |= signCompress(trapvect8, 8)

	return ret
}

func NewGetc() Word {
	return NewTrap(0x20)
}

func NewOut() Word {
	return NewTrap(0x21)
}

func NewPuts() Word {
	return NewTrap(0x22)
}

func NewIn() Word {
	return NewTrap(0x23)
}

func NewPutsp() Word {
	return NewTrap(0x24)
}

func NewHalt() Word {
	return NewTrap(0x25)
}
