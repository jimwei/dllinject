package main

import (
	_ "errors"
	"log"
	"os/exec"
	"reflect"
	"runtime"
	"syscall"
	"time"
)

const (
	//createremotethread
	PROCESS_CREATE_THREAD  = 0x0002
	PROCESS_CREATE_PROCESS = 0x0080
	PROCESS_VM_OPERATION   = 0x0008
	PROCESS_VM_READ        = 0x0010
	PROCESS_VM_WRITE       = 0x0020

	//virtualallocEx
	MEM_COMMIT             = 0x1000
	MEM_RESERVE            = 0x2000
	PAGE_EXECUTE_READWRITE = 0x40

	THREADSIZE = 4096
)

type threadProc func(uintptr) uint32

func ThreadProc(lparam uintptr) uint32 {
	return 0
}
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	runtime.LockOSThread()
	cmd := exec.Command("notepad")
	cmd.Start()
	//wait the app run successfully
	time.Sleep(time.Microsecond * 100)

	//open process
	_desHandler, err := syscall.OpenProcess(PROCESS_CREATE_THREAD|
		PROCESS_VM_OPERATION|
		PROCESS_VM_WRITE,
		false,
		uint32(cmd.Process.Pid))
	if err != nil {
		log.Println("OpenProcess", err.Error())
	}
	log.Println("the destination handler is", _desHandler)

	inject := &Inject{}
	inject.Inject(_desHandler)
	return
	// load kernel32 dll
	k32, err := syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		log.Println("LoadLibrary('kernel32.dll')", err.Error())
		return
	} else {
		log.Println("load kernel successfully.")
	}
	defer syscall.FreeLibrary(k32)

	//get VirtualAllocEx
	_virtualAlloc, err := syscall.GetProcAddress(k32, "VirtualAllocEx")
	if err != nil {
		log.Println("VirtualAllocEx", err.Error())
		return
	} else {
		log.Println("get virtualalloex successfully.")
	}
	_, _, err = syscall.Syscall6(_virtualAlloc,
		uintptr(_desHandler),
		0,
		THREADSIZE,
		MEM_COMMIT|MEM_RESERVE,
		PAGE_EXECUTE_READWRITE,
		0,
		0)

	if err != nil {
		log.Println("Syscall6(_virtualAlloc", err.Error())
		return
	} else {
		log.Println("Syscall6 _virtualAlloc successfully.")
	}

	//get writeprocessmemory api
	_writeProcessMemory, err := syscall.GetProcAddress(k32, "WriteProcessMemory")
	if err != nil {
		log.Println("WriteProcessMemory", err.Error())
		return
	} else {
		log.Println("get writeprocessmemory successfully.")
	}
	//inject the threradproc into destination process
	_threadProc := ThreadProc
	//	log.Println("the thread proc addr is", string(reflect.ValueOf(threadProc).UnsafeAddr()))

	_, _, err = syscall.Syscall6(_writeProcessMemory,
		uintptr(_desHandler),
		_virtualAlloc,
		reflect.ValueOf(_threadProc).UnsafeAddr(),
		THREADSIZE,
		0,
		0,
		0)
	if err != nil {
		log.Println("Syscall6(_writeProcessMemory", err.Error())
		return
	}

	//create remote thead in destination process
	// get createremotethread proc
	rt, err := syscall.GetProcAddress(k32, "CreateRemoteThread")

	if err != nil {
		log.Println("CreateRemoteThread", err.Error())
		return
	}
	log.Println(rt)
	var dwWriteBytes uintptr
	_, _, err = syscall.Syscall6(rt,
		uintptr(_desHandler),
		0,
		0,
		_virtualAlloc,
		0,
		0,
		dwWriteBytes)
	if err != nil {
		log.Println("Syscall6(CreateRemoteThread)", err.Error())
		return
	}
}
