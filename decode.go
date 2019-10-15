package lc3

func MakeNegative(value int) Word {
	return Word(value)
}

func getOpcode(instruction Word) Word {
	return instruction >> 12
}

func IsNegative(x Word) bool {
	return x&0x8000 > 0
}

func signExtend(x Word, bitsCount uint) Word {
	if x&(1<<(bitsCount-1)) > 0 {
		//fmt.Printf("x:    %016b\n", x)
		//fmt.Printf("max:  %016b\n", WordMax)
		//fmt.Printf("mask: %016b\n", (WordMax<<bitsCount) & WordMax)
		//fmt.Printf("res:  %016b\n", x|(WordMax<<bitsCount))
		// todo: find simpler function
		return x | ((WordMax << bitsCount) & WordMax)
	}
	return x
}

func getNBits(x Word, from uint, bits uint) Word {
	return (x >> from) & (1<<bits - 1)
}

func getNBitsExtended(x Word, from uint, bits uint) Word {
	return signExtend(getNBits(x, from, bits), bits)
}
