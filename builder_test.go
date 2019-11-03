package lc3

import (
	"testing"
)

func decodeBinaryString(input string) Word {
	var ret Word

	for len(input) > 0 {
		char := input[0]
		if char == ' ' {
			input = input[1:]
			continue
		}

		if char == '1' {
			ret <<= 1
			ret |= 1
		} else if char == '0' {
			ret <<= 1
		} else {
			panic(char)
		}
		input = input[1:]
	}
	return ret
}

func Test_signCompress(t *testing.T) {
	type testCase struct {
		value   Word
		toBits  Word
		encoded Word
	}
	testData := []testCase{
		// 42(dec)  = 0000 0000 0010 1010(bin)
		{value: 42, toBits: 6, encoded: 42}, // normal encoding
		{value: 42, toBits: 5, encoded: 10}, // one bit overflow
		{value: 42, toBits: 4, encoded: 10}, // two bits overflow
		// -42(dec) = 1111 1111 1101 0110(bin)
		{value: MakeNegative(-42), toBits: 16, encoded: 65494}, // no overflow
		{value: MakeNegative(-42), toBits: 10, encoded: 982},   // no overflow
		{value: MakeNegative(-42), toBits: 5, encoded: 22},     // overflow
	}

	for i := range testData {
		encoded := signCompress(testData[i].value, testData[i].toBits)
		if encoded != testData[i].encoded {
			t.Errorf("%d != %d", encoded, testData[i].encoded)
		}
	}
}

func Test_NewBRInvalid(t *testing.T) {
	NewBR(0, 0)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("panic expected for NewBR(false, false, false, 0)")
		}
	}()
}

func Test_NewBR(t *testing.T) {
	type testCase struct {
		n         bool
		z         bool
		p         bool
		pcOffset9 Word
		encoded   string
	}

	testData := []testCase{
		{false, false, true, 1, "0000 001 000000001"},
		{false, true, false, 2, "0000 010 000000010"},
		{false, true, true, 3, "0000 011 000000011"},
		{true, false, false, 4, "0000 100 000000100"},
		{true, false, true, 5, "0000 101 000000101"},
		{true, true, false, 6, "0000 110 000000110"},
		{true, true, true, MakeNegative(-1), "0000 111 111111111"},
	}

	for i := range testData {
		word := decodeBinaryString(testData[i].encoded)
		var flags Word = 0
		if testData[i].n {
			flags |= FlN
		}
		if testData[i].z {
			flags |= FlZ
		}
		if testData[i].p {
			flags |= FlP
		}
		instruction := NewBR(flags, testData[i].pcOffset9)
		if instruction != word {
			t.Errorf("%d: %016b != %016b", i, instruction, word)
		}
	}
}

func Test_NewAddRegister(t *testing.T) {
	type testCase struct {
		dr      Word
		sr1     Word
		sr2     Word
		encoded string
	}
	testData := []testCase{
		{0, 0, 0, "0001 000 000 0 00 000"},
		{5, 6, 7, "0001 101 110 0 00 111"},
	}

	for i := range testData {
		word := decodeBinaryString(testData[i].encoded)
		instruction := NewAddRegister(testData[i].dr, testData[i].sr1, testData[i].sr2)
		if instruction != word {
			t.Errorf("%d: %016b != %016b", i, instruction, word)
		}
	}
}

func Test_NewAddImmediate(t *testing.T) {
	type testCase struct {
		dr      Word
		sr      Word
		imm5    Word
		encoded string
	}

	testData := []testCase{
		{0, 0, 0, "0001 000 000 1 00000"},
		{4, 5, MakeNegative(-2), "0001 100 101 1 11110"},
	}

	for i := range testData {
		word := decodeBinaryString(testData[i].encoded)
		instruction := NewAddImmediate(testData[i].dr, testData[i].sr, testData[i].imm5)
		if instruction != word {
			t.Errorf("%d: %016b != %016b", i, instruction, word)
		}
	}
}

func Test_NewLd(t *testing.T) {
	type testCase struct {
		dr        Word
		pcOffset9 Word
		encoded   string
	}

	testData := []testCase{
		{0, 0, "0010 000 000000000"},
		{5, 42, "0010 101 000101010"},
		{0, MakeNegative(-1), "0010 000 111111111"},
	}
	for i := range testData {
		word := decodeBinaryString(testData[i].encoded)
		instruction := NewLd(testData[i].dr, testData[i].pcOffset9)
		if instruction != word {
			t.Errorf("%d: %016b != %016b", i, instruction, word)
		}
	}
}

func Test_NewSt(t *testing.T) {
	type testCase struct {
		sr        Word
		pcOffset9 Word
		encoded   string
	}
	testData := []testCase{
		{0, 0, "0011 000 000000000"},
		{5, 42, "0011 101 000101010"},
		{0, MakeNegative(-1), "0011 000 111111111"},
	}
	for i := range testData {
		word := decodeBinaryString(testData[i].encoded)
		instruction := NewSt(testData[i].sr, testData[i].pcOffset9)
		if instruction != word {
			t.Errorf("%d: %016b != %016b", i, instruction, word)
		}
	}
}

func Test_NewJsr(t *testing.T) {
	type testCase struct {
		pcOffset11 Word
		encoded    string
	}
	testData := []testCase{
		{0, "0100 1 00000000000"},
		{42, "0100 1 00000101010"},
		{MakeNegative(-1), "0100 1 11111111111"},
	}

	for i := range testData {
		word := decodeBinaryString(testData[i].encoded)
		instruction := NewJsr(testData[i].pcOffset11)
		if instruction != word {
			t.Errorf("%d: %016b != %016b", i, instruction, word)
		}
	}
}

func Test_NewJsrr(t *testing.T) {
	type testCase struct {
		baseR   Word
		encoded string
	}
	testData := []testCase{
		{0, "0100 0 00 000 000000"},
		{5, "0100 0 00 101 000000"},
	}

	for i := range testData {
		word := decodeBinaryString(testData[i].encoded)
		instruction := NewJsrr(testData[i].baseR)
		if instruction != word {
			t.Errorf("%d: %016b != %016b", i, instruction, word)
		}
	}
}

func Test_NewAndImmediate(t *testing.T) {
	type testCase struct {
		dr      Word
		sr      Word
		imm5    Word
		encoded string
	}

	testData := []testCase{
		{0, 0, 0, "0101 000 000 1 00000"},
		{5, 7, 13, "0101 101 111 1 01101"},
	}
	for i := range testData {
		word := decodeBinaryString(testData[i].encoded)
		instruction := NewAndImmediate(testData[i].dr, testData[i].sr, testData[i].imm5)
		if instruction != word {
			t.Errorf("%d: %016b != %016b", i, instruction, word)
		}
	}
}

func Test_NewAndRegister(t *testing.T) {
	type testCase struct {
		dr      Word
		sr1     Word
		sr2     Word
		encoded string
	}

	testData := []testCase{
		{0, 0, 0, "0101 000 000 0 00 000"},
		{5, 3, 7, "0101 101 011 0 00 111"},
	}

	for i := range testData {
		word := decodeBinaryString(testData[i].encoded)
		instruction := NewAndRegister(testData[i].dr, testData[i].sr1, testData[i].sr2)
		if instruction != word {
			t.Errorf("%d: %016b != %016b", i, instruction, word)
		}
	}
}

func Test_NewLdr(t *testing.T) {
	type testCase struct {
		dr        Word
		baseR     Word
		pcOffset6 Word
		encoded   string
	}

	testData := []testCase{
		{0, 0, 0, "0110 000 000 000000"},
		{5, 3, 7, "0110 101 011 000111"},
		{7, 7, MakeNegative(-1), "0110 111 111 111111"},
	}

	for i := range testData {
		word := decodeBinaryString(testData[i].encoded)
		instruction := NewLdr(testData[i].dr, testData[i].baseR, testData[i].pcOffset6)
		if instruction != word {
			t.Errorf("%d: %016b != %016b", i, instruction, word)
		}
	}
}

func Test_NewStr(t *testing.T) {
	type testCase struct {
		sr      Word
		baseR   Word
		offset6 Word
		encoded string
	}

	testData := []testCase{
		{0, 0, 0, "0111 000 000 000000"},
		{5, 3, 7, "0111 101 011 000111"},
		{7, 7, MakeNegative(-1), "0111 111 111 111111"},
	}

	for i := range testData {
		word := decodeBinaryString(testData[i].encoded)
		instruction := NewStr(testData[i].sr, testData[i].baseR, testData[i].offset6)
		if instruction != word {
			t.Errorf("%d: %016b != %016b", i, instruction, word)
		}
	}
}

func Test_NewNot(t *testing.T) {
	type testCase struct {
		dr      Word
		sr      Word
		encoded string
	}

	testData := []testCase{
		{0, 0, "1001 000 000 111111"},
		{5, 3, "1001 101 011 111111"},
	}

	for i := range testData {
		word := decodeBinaryString(testData[i].encoded)
		instruction := NewNot(testData[i].dr, testData[i].sr)
		if instruction != word {
			t.Errorf("%d: %016b != %016b", i, instruction, word)
		}
	}
}

func Test_NewLdi(t *testing.T) {
	type testCase struct {
		dr        Word
		pcOffset9 Word
		encoded   string
	}
	testData := []testCase{
		{0, 0, "1010 000 000000000"},
		{5, 42, "1010 101 000101010"},
		{0, MakeNegative(-1), "1010 000 111111111"},
	}
	for i := range testData {
		word := decodeBinaryString(testData[i].encoded)
		instruction := NewLdi(testData[i].dr, testData[i].pcOffset9)
		if instruction != word {
			t.Errorf("%d: %016b != %016b", i, instruction, word)
		}
	}
}

func Test_NewSti(t *testing.T) {
	type testCase struct {
		sr        Word
		pcOffset9 Word
		encoded   string
	}
	testData := []testCase{
		{0, 0, "1011 000 000000000"},
		{5, 42, "1011 101 000101010"},
		{0, MakeNegative(-1), "1011 000 111111111"},
	}
	for i := range testData {
		word := decodeBinaryString(testData[i].encoded)
		instruction := NewSti(testData[i].sr, testData[i].pcOffset9)
		if instruction != word {
			t.Errorf("%d: %016b != %016b", i, instruction, word)
		}
	}
}

func Test_NewJmp(t *testing.T) {
	type testCase struct {
		baseR   Word
		encoded string
	}
	testData := []testCase{
		{5, "1100 000 101 000000"},
	}
	for i := range testData {
		word := decodeBinaryString(testData[i].encoded)
		instruction := NewJmp(testData[i].baseR)
		if instruction != word {
			t.Errorf("%d: %016b != %016b", i, instruction, word)
		}
	}
}

func Test_NewLea(t *testing.T) {
	type testCase struct {
		dr        Word
		pcOffset9 Word
		encoded   string
	}
	testData := []testCase{
		{0, 0, "1110 000 000000000"},
		{5, 42, "1110 101 000101010"},
		{0, MakeNegative(-1), "1110 000 111111111"},
	}
	for i := range testData {
		word := decodeBinaryString(testData[i].encoded)
		instruction := NewLea(testData[i].dr, testData[i].pcOffset9)
		if instruction != word {
			t.Errorf("%d: %016b != %016b", i, instruction, word)
		}
	}
}

func Test_NewTrap(t *testing.T) {
	type testCase struct {
		trapVect8 Word
		encoded   string
	}
	testData := []testCase{
		{0, "1111 0000 00000000"},
		{0x27, "1111 0000 00100111"},
	}
	for i := range testData {
		word := decodeBinaryString(testData[i].encoded)
		instruction := NewTrap(testData[i].trapVect8)
		if instruction != word {
			t.Errorf("%d: %016b != %016b", i, instruction, word)
		}
	}
}
