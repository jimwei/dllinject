package main

import (
	"log"
	"syscall"
	"unsafe"
)

type Inject struct{}

func (this *Inject) Inject(desHandle syscall.Handle) {
	k32 := syscall.MustLoadDLL("kernel32.dll")
	defer k32.Release()
	_virtualAlloc := k32.MustFindProc("VirtualAllocEx")
	_writeProcessMemory := k32.MustFindProc("WriteProcessMemory")
	_createRemoteThread := k32.MustFindProc("CreateRemoteThread")

	log.Println(_writeProcessMemory)
	log.Println(_createRemoteThread)

	r1, _, err := _virtualAlloc.Call(uintptr(desHandle),
		0,
		THREADSIZE,
		MEM_COMMIT|MEM_RESERVE,
		PAGE_EXECUTE_READWRITE)
	if int(r1) == 0 {
		log.Println("exec VirtualAllocEx fail. r1=0", err.Error())
		return
	}
	//inject the threradproc into destination process
	_threadProc := ThreadProc
	//	log.Println("the thread proc addr is", string(reflect.ValueOf(threadProc).UnsafeAddr()))
	r1, _, err = _writeProcessMemory.Call(uintptr(desHandle),
		r1,
		uintptr(unsafe.Pointer(&_threadProc)),
		THREADSIZE,
		0)
	log.Println("exec WriteProcessMemory.", err.Error())
	if int(r1) == 0 {
		log.Println("exec WriteProcessMemory fail. r1=0", err.Error())
		return
	}

}
