package gfx

import (
	"syscall"
	"unsafe"
	"errors"
)

var (
	qpc, qpf *syscall.Proc
)

func init() {
	dll, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return
	}
	qpf, _ = dll.FindProc("QueryPerformanceFrequency")
	qpc, _ = dll.FindProc("QueryPerformanceCounter")
}

func QueryPerformanceCounter() (int64, error) {
	if qpc == nil {
		return 0, errors.New("unavailable")
	}
	var ctr int64
	ret, _, err := qpc.Call(uintptr(unsafe.Pointer(&ctr)))
	if ret == 0 {
		return 0, err
	}

	return ctr, nil
}

func QueryPerformanceFreq() (int64, error) {
	if qpf == nil {
		return 0, errors.New("unavailable")
	}
	var ctr int64
	ret, _, err := qpf.Call(uintptr(unsafe.Pointer(&ctr)))
	if ret == 0 {
		return 0, err
	}

	return ctr, nil
}
