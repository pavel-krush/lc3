package lc3

import "testing"

func Test_Add(t *testing.T) {
	vmTestCases{
		newVMTestCase().setAssemblerCode(`
					add r0, r0, #0
					add r1, r1, #1
					add r2, r2, #-1
					add r3, r1, r2
					halt`).
			expectRegister(RegR0, 0).
			expectRegister(RegR1, 1).
			expectRegister(RegR2, MakeNegative(-1)).
			expectRegister(RegR3, 0),
	}.Run(t)
}

func Test_BR(t *testing.T) {
	vmTestCases{
		// BRN
		newVMTestCase().setAssemblerCode(`
					add r0, r0, #-1
					brz zcond ;not jump
					brp pcond ;not jump
					brn ncond ;jump
					halt ;unreachable
			zcond	halt ;unreachable
			pcond	halt ;unreachable
			ncond	halt ;must stop here
			`).expectRegister(RegPC, 8),
		// BRZ
		newVMTestCase().setAssemblerCode(`
					add r0, r0, #0
					brp pcond ;not jump
					brn ncond ;not jump
					brz zcond ;jump
					halt ;unreachable
			zcond	halt ;unreachable
			pcond	halt ;unreachable
			ncond	halt ;must stop here
			`).expectRegister(RegPC, 6),
		// BRO
		newVMTestCase().setAssemblerCode(`
					add r0, r0, #1
					brn ncond ;not jump
					brz zcond ;not jump
					brp pcond ;jump
					halt ;unreachable
			zcond	halt ;unreachable
			pcond	halt ;unreachable
			ncond	halt ;must stop here
			`).expectRegister(RegPC, 7),
	}.Run(t)
}

func Test_Ld(t *testing.T) {
	vmTestCases{
		newVMTestCase().
			setAssemblerCode(`
					ld r0, #-1 ;load this instruction into r0
					halt`).
			expectRegister(RegR0, NewLd(RegR0, MakeNegative(-1))),
	}.Run(t)
}

func Test_St(t *testing.T) {
	vmTestCases{
		newVMTestCase().setAssemblerCode(`
					add r0, r0, #13
					st r0, #1 ;store 13 in the word after the halt instruction
					halt
			`).expectMemory(3, 13),
	}.Run(t)
}

func Test_Jsr(t *testing.T) {
	vmTestCases{
		newVMTestCase().setAssemblerCode(`
					jsr #1 ; must jump and set r7 = 1
					halt ;unreachable
					halt ;must stop here`).
			expectRegister(RegR7, 1).
			expectRegister(RegPC, 3),
		newVMTestCase().setAssemblerCode(`
					add r0, r0, #3
					jsrr r0 ;save r0 to r7 and jump to second halt
					halt
					halt
					`).
			expectRegister(RegR7, 2).expectRegister(RegPC, 4),
	}.Run(t)
}

func Test_And(t *testing.T) {
	vmTestCases{
		newVMTestCase().
			setAssemblerCode(`
					add r0, r0, #13
					add r1, r1, #42
					and r2, r0, r1
					halt`).
			expectRegister(RegR2, 8).
			expectFlags(FlP),
	}.Run(t)
}

func Test_Ldr(t *testing.T) {
	vmTestCases{
		newVMTestCase().
			setAssemblerCode(`
					add r6, r6, #3 
					ldr r0, r6, #1
					halt
					.fill #1 ;r6 will point here
					.fill #42 ;this value must be loaded`).
			expectRegister(RegR0, 42),
	}.Run(t)
}

func Test_Str(t *testing.T) {
	vmTestCases{
		newVMTestCase().
			setAssemblerCode(`
					add r6, r6, #4 
					add r0, r0, #13
					str r0, r6, #1
					halt
					.fill #1 ;r6 will point here
					.fill #42 ;this value will be overwritten`).
			expectMemory(5, 13),
	}.Run(t)
}

func Test_Not(t *testing.T) {
	vmTestCases{
		newVMTestCase().
			setAssemblerCode(`
					add r0, r0, #13
					not r1, r0
					halt`).
			expectRegister(RegR1, MakeNegative(-14)),
	}.Run(t)
}

func Test_Ldi(t *testing.T) {
	vmTestCases{
		newVMTestCase().
			setAssemblerCode(`
					ldi r1, #1 ;address of first .fill
					halt
					.fill x3 ;pointer to next word
					.fill x42 ;this value must be loaded`).
			expectRegister(RegR1, 0x42),
	}.Run(t)
}

func Test_Sti(t *testing.T) {
	vmTestCases{
		newVMTestCase().
			setAssemblerCode(`
					add r1, r1, #13
					sti r1, #1 ;address of first .fill
					halt
					.fill x4 ;pointer to next word
					.fill x42 ;this value will be overwritten`).
			expectMemory(4, 13),
	}.Run(t)
}

func Test_Jmp(t *testing.T) {
	vmTestCases{
		newVMTestCase().
			setAssemblerCode(`
					add r0, r0, #3
					jmp r0
					halt
					halt ;must stop here`).
			expectRegister(RegPC, 4),
	}.Run(t)
}

func Test_Lea(t *testing.T) {
	vmTestCases{
		newVMTestCase().
			setAssemblerCode(`
					lea r1, stack
					stack halt`).
			expectRegister(RegR1, 1),
	}.Run(t)
}
