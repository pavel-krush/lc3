package lc3

import (
	"github.com/pkg/errors"
	"strings"
	"testing"
)

type vmTestCase struct {
	vm            *VM
	assemblerCode string

	expectedRegister map[int]Word
	isExpectedFlags  bool
	expectedFlags    Word
	expectedMemory   map[Word]Word
}

type vmTestCases []*vmTestCase

func (vmts vmTestCases) Run(t *testing.T) {
	for i := range vmts {
		err := vmts[i].Run()
		if err != nil {
			t.Error(err)
		}
	}
}

func newVMTestCase() *vmTestCase {
	return &vmTestCase{
		expectedRegister: make(map[int]Word),
		expectedMemory:   make(map[Word]Word),
	}
}

func (vmt *vmTestCase) setAssemblerCode(code string) *vmTestCase {
	code = strings.TrimSuffix(code, "\n") // cut first \n for convenient lines count in tests
	vmt.assemblerCode = code
	return vmt
}

func (vmt *vmTestCase) expectRegister(register int, value Word) *vmTestCase {
	vmt.expectedRegister[register] = value
	return vmt
}

func (vmt *vmTestCase) expectFlags(flags Word) *vmTestCase {
	vmt.isExpectedFlags = true
	vmt.expectedFlags = flags
	return vmt
}

func (vmt *vmTestCase) expectMemory(address Word, value Word) *vmTestCase {
	vmt.expectedMemory[address] = value
	return vmt
}

func (vmt *vmTestCase) checkExpectations() error {
	m := vmt.vm

	for register, value := range vmt.expectedRegister {
		if m.registers[register] != value {
			return errors.Errorf("expected register %d = %s, got %s", register, value.AsString(), m.registers[register].AsString())
		}
	}

	if vmt.isExpectedFlags && vmt.expectedFlags != m.registers[RegCond] {
		return errors.Errorf("expected flags = %s, got %s", vmt.expectedFlags.FlagsAsString(), m.registers[RegCond].FlagsAsString())
	}

	for address, value := range vmt.expectedMemory {
		vmValue := m.ReadMem(address)
		if vmValue != value {
			return errors.Errorf("expected memory at %s = %s, got %s", address.AsString(), value.AsString(), vmValue.AsString())
		}
	}

	return nil
}

func (vmt *vmTestCase) Run() error {
	m, err := ParseAssembly(strings.NewReader(vmt.assemblerCode))
	if err != nil {
		return err
	}

	m.Start()
	for {
		//m.Dump(true)
		err := m.Step()
		if err == ErrNotRunning {
			break
		}
	}

	vmt.vm = m

	return vmt.checkExpectations()
}
