package main

import (
	"fmt"
	memory "main/memory"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	HANDLE   uintptr
	TWOMPID  uint32
	BaseAddr int64
)

func GetTWOM() {
	var entry syscall.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(syscall.ProcessEntry32{}))

	snhandle := memory.CreateToolhelp32Snapshot()

	ok, snapshot := memory.Process32FirstW(snhandle, entry)
	if ok {
		for {
			okproc, process := memory.Process32NextW(snhandle, snapshot)
			if !okproc {
				break
			}

			if strings.Contains(syscall.UTF16ToString(process.ExeFile[:]), "This War of Mine.exe") {
				TWOMPID = process.ProcessID
				break
			}
		}
	}

	if !(TWOMPID > 0) {
		fmt.Println("Failed to find the process id of This War of Mine. (is This War of Mine.exe running?) -> Retrying in 5 seconds...")
		return
	}

	HANDLE = memory.OpenProcess(TWOMPID)

	if !(HANDLE > 0) {
		fmt.Println("Failed to find the handle of This War of Mine. (Antivirus blocking OpenProcess?) -> Retrying in 5 seconds...")
		return
	}

	Modules := memory.EnumProcessModules(HANDLE)

	for i := 0; i < len(Modules); i++ {
		ModuleFileName := memory.GetModuleFileNameExW(HANDLE, Modules[i])
		if strings.Contains(ModuleFileName, "This War of Mine.exe") {
			ModuleInfo := memory.GetModuleInformation(HANDLE, Modules[i])
			BaseAddr = int64(ModuleInfo.LpBaseOfDll)
			break
		}
	}

	if !(BaseAddr > 0) {
		fmt.Println("Failed to find the base address of This War of Mine. (Antivirus blocking module operations?) -> Retrying in 5 seconds...")
		return
	}

	fmt.Println("This War of Mine PID, Handle, Base Adress was found without issues.")
}

func main() {
	for !(TWOMPID > 0 && HANDLE > 0 && BaseAddr > 0) {
		GetTWOM()
		time.Sleep(5 * time.Second)
	}
	fmt.Printf("TWOM PID: %v\r\nTWOM Handle: %v\r\nTWOM Base Address: %v\r\n", TWOMPID, HANDLE, BaseAddr)
	path, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get current directory... ->", err)
		fmt.Scanln()
		os.Exit(1)
	}
	fmt.Println("Working Directory is ->", path)

	dllPath := fmt.Sprintf(`%v\TWOMUHook\x64\TWOMUHook.dll`, path)

	if _, err := os.Stat(dllPath); err != nil {
		fmt.Println("Failed to get dllPath... ->", err)
		fmt.Scanln()
		os.Exit(1)
	}

	Valloc := memory.VirtualAllocEx(HANDLE, 0, uintptr(len(dllPath)+1), 0x00002000|0x00001000, 4)

	memory.WriteProcessMemory(HANDLE, Valloc, []byte(dllPath), uintptr(len(dllPath)+1))

	modKernel, err := syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		fmt.Println("Failed to load kernel32.dll ->", err)
		fmt.Scanln()
		os.Exit(1)
	}

	LoadLibrary, err := syscall.GetProcAddress(modKernel, "LoadLibraryA")
	if err != nil {
		fmt.Println("Failed to get LoadLibraryA ProcAddress ->", err)
		fmt.Scanln()
		os.Exit(1)
	}

	memory.CreateRemoteThread(HANDLE, 0, 0, LoadLibrary, Valloc, 0)

	fmt.Println("TWOMUHook was injected successfully... Exiting in 5 seconds")
	time.Sleep(5 * time.Second)
	os.Exit(0)
}
