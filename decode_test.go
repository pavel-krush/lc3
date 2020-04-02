package lc3

import "testing"

func Test_signExtend(t *testing.T) {
	type testCase struct {
		value     Word
		bitsCount uint
		extended  Word
	}

	testData := []testCase{
		{42, 8, 42},
		{42, 7, 42},
		{42, 6, 65514},
	}

	for i := range testData {
		decoded := signExtend(testData[i].value, testData[i].bitsCount)
		if decoded != testData[i].extended {
			t.Errorf("%d != %d", decoded, testData[i].extended)
		}
	}
}
