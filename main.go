package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	memory "main/memory"
	"math"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var (
	PCheckFG bool    = true
	PFixCam  bool    = true
	PReadMem bool    = true
	Step     float32 = 0.3
	TWOMPID  uint32
	BaseAddr int64

	XPos, YPos, CMode          uintptr
	XBuffer, YBuffer, CMBuffer []uint8
	X, Y                       float32
	CM                         uint32

	CwMem = make(chan bool)
	Mutex sync.Mutex
)

func main() {
	fmt.Println("This War of Mine Better Camera")

	// because flag.Float32Var doesn't exist.
	var tmpFloat64 float64

	flag.Float64Var(&tmpFloat64, "Step", 0.3, "Step determines how fast/much the camera should move when pressing W/A/S/D")

	flag.BoolVar(&PCheckFG, "CheckFG", true, "CheckFG will periodically (every 10ms) check if This War of Mine is the foreground application\r\nThis fixes an issue, where even if TWOM is not foreground, key inputs would register, and set X, Y positions from other applications.\r\nRecommended value is true")

	flag.BoolVar(&PReadMem, "ReadMem", true, "ReadMem will periodically (every 10ms) write the X, Y coordinates from the game's memory to a stored one\r\nThis fixes an issue where pressing tab or using the mouse to change camera position would rubberband the camera back.\r\nRecommended value is true")

	flag.BoolVar(&PFixCam, "FixCam", true, "FixCam will periodically (every 10ms) check, and set a value to an address that controls the camera mode.\r\nThis fixes a notorious issue, that when you loaded into a level, or moved the camera by other means, would disable the ability to use W/A/S/D controls\r\nHighly recommended to keep this value on true.")

	flag.Parse()

	Step = float32(tmpFloat64)

	if Step <= 0 {
		Step = 0.3
	}

	fmt.Printf("Step: %v | PCheckFG: %v | FixCam: %v\r\n", Step, PCheckFG, PFixCam)

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
	fmt.Println("This War of Mine PID ->", TWOMPID)

	HANDLE := memory.OpenProcess(TWOMPID)

	Modules := memory.EnumProcessModules(HANDLE)

	for i := 0; i < len(Modules); i++ {
		if strings.Contains(memory.GetModuleFileNameExW(HANDLE, Modules[i]), "This War of Mine.exe") {
			BaseAddr = int64(memory.GetModuleInformation(HANDLE, Modules[i]).LpBaseOfDll)
			break
		}
	}

	XPos = memory.Offsets(HANDLE, uintptr(BaseAddr), 0x009064D0, 0x70)
	XBuffer = memory.ReadProcessMemory(HANDLE, XPos, 8)
	X = math.Float32frombits(binary.LittleEndian.Uint32(XBuffer))

	YPos = memory.Offsets(HANDLE, uintptr(BaseAddr), 0x009064D0, 0x78)
	YBuffer = memory.ReadProcessMemory(HANDLE, YPos, 8)
	Y = math.Float32frombits(binary.LittleEndian.Uint32(YBuffer))

	CMode = memory.Offsets(HANDLE, uintptr(BaseAddr), 0x009064D0, 0xA6)

	if PCheckFG {
		go func() {
			for {
				if memory.GetWindowThreadProcessId(memory.GetForegroundWindow()) == TWOMPID {
					CwMem <- true
				} else {
					CwMem <- false
				}
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}

	if PReadMem {
		go func() {
			for {
				Mutex.Lock()

				XBuffer = memory.ReadProcessMemory(HANDLE, XPos, 8)
				X = math.Float32frombits(binary.LittleEndian.Uint32(XBuffer))

				YBuffer = memory.ReadProcessMemory(HANDLE, YPos, 8)
				Y = math.Float32frombits(binary.LittleEndian.Uint32(YBuffer))

				Mutex.Unlock()
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}

	if PFixCam {
		go func() {
			for {
				Mutex.Lock()
				CMBuffer = memory.ReadProcessMemory(HANDLE, CMode, 4)
				CM = binary.LittleEndian.Uint32(CMBuffer)

				if CM != 148602 {
					memory.WriteProcessMemoryInt(HANDLE, CMode, 148602)
				}
				Mutex.Unlock()
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}

	for {
		Mutex.Lock()
		if <-CwMem {
			// W
			if memory.GetAsyncKeyState(0x57) {
				Y += Step
				memory.WriteProcessMemoryFloat(HANDLE, YPos, Y)
			}

			// A
			if memory.GetAsyncKeyState(0x41) {
				X -= Step
				memory.WriteProcessMemoryFloat(HANDLE, XPos, X)
			}

			// S
			if memory.GetAsyncKeyState(0x53) {
				Y -= Step
				memory.WriteProcessMemoryFloat(HANDLE, YPos, Y)
			}

			// D
			if memory.GetAsyncKeyState(0x44) {
				X += Step
				memory.WriteProcessMemoryFloat(HANDLE, XPos, X)
			}
		}
		Mutex.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
}
