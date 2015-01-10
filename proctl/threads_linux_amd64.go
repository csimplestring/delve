package proctl

/*
#include "stdint.h"
#include <sys/ptrace.h>
#include <stddef.h>
#include <sys/user.h>
#include <sys/debugreg.h>
#include <errno.h>

int setHardwareBreakpoint(int reg, int tid, uint64_t addr) {
	if (reg < 0 || reg > 3) return -1;

	uint64_t dr7 = (0x1 | DR_RW_EXECUTE | DR_LEN_8);

	if (ptrace(PTRACE_POKEUSER, tid, offsetof(struct user, u_debugreg[0]), addr)) {
		return -1;
	}
	if (ptrace(PTRACE_POKEUSER, tid, offsetof(struct user, u_debugreg[DR_CONTROL]), dr7)) {
		return -2;
	}
	return 0;
}
*/
import "C"

import (
	"fmt"
	"syscall"
)

func (thread *ThreadContext) setBreakpointInProcess(tid int, addr uint64) (int, error) {
	for i, v := range thread.Process.HWBreakPoints {
		if v == nil {
			ret := C.setHardwareBreakpoint(C.int(i), C.int(thread.Id), C.uint64_t(addr))
			if ret < 0 {
				return -1, fmt.Errorf("could not set hardware breakpoint")
			}
			return i, nil
		}
	}
	int3 := []byte{0xCC}
	_, err := writeMemory(thread.Id, uintptr(addr), int3)
	return -1, err
}

type Regs struct {
	regs *syscall.PtraceRegs
}

func (r *Regs) PC() uint64 {
	return r.regs.PC()
}

func (r *Regs) SP() uint64 {
	return r.regs.Rsp
}

func (r *Regs) SetPC(tid int, pc uint64) error {
	r.regs.SetPC(pc)
	return syscall.PtraceSetRegs(tid, r.regs)
}

func registers(tid int) (Registers, error) {
	var regs syscall.PtraceRegs
	err := syscall.PtraceGetRegs(tid, &regs)
	if err != nil {
		return nil, err
	}
	return &Regs{&regs}, nil
}

func writeMemory(tid int, addr uintptr, data []byte) (int, error) {
	return syscall.PtracePokeData(tid, addr, data)
}

func readMemory(tid int, addr uintptr, data []byte) (int, error) {
	return syscall.PtracePeekData(tid, addr, data)
}
