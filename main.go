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
	PCheckFG          bool
	PReadMem          bool
	PFixCam           bool
	DisablePEffect    bool
	DisableRainEffect bool
	ModifyWndProc     bool
	PrintErr          bool
	Step              float32

	TWOMPID  uint32
	BaseAddr int64

	XPos, YPos, CMode, Rain, Pencil, WndProc uintptr
	XBuffer, YBuffer, CMBuffer               []uint8
	X, Y                                     float32
	CM                                       uint32

	CwMem = make(chan bool)
	Mutex sync.Mutex

	S = time.NewTimer(10 * time.Millisecond)
)

func main() {
	fmt.Println("This War of Mine Better Camera")

	// because flag.Float32Var doesn't exist.
	var tmpFloat64 float64

	flag.Float64Var(&tmpFloat64, "Step", 0.7, "Step determines how fast/much the camera should move when pressing W/A/S/D")

	flag.BoolVar(&PCheckFG, "CheckFG", true, "CheckFG will periodically check if This War of Mine is the foreground application\r\nThis fixes an issue, where even if TWOM is not foreground, key inputs would register, and set X, Y positions from other applications.\r\nRecommended value is true")

	flag.BoolVar(&PReadMem, "ReadMem", true, "ReadMem will periodically (every 10ms) write the X, Y coordinates from the game's memory to a stored one\r\nThis fixes an issue where pressing tab or using the mouse to change camera position would rubberband the camera back.\r\nRecommended value is true")

	flag.BoolVar(&PFixCam, "FixCam", true, "FixCam will periodically (every 10ms) check, and set a value to an address that controls the camera mode.\r\nThis fixes a notorious issue, that when you loaded into a level, or moved the camera by other means, would disable the ability to use W/A/S/D controls\r\nHighly recommended to keep this value on true.")

	flag.BoolVar(&DisablePEffect, "DisablePencil", false, "If set to true, DisablePencil will disable the in-game pencil effect.\r\nNote: this doesn't have an effect on frame rate.")

	flag.BoolVar(&DisableRainEffect, "DisableRain", false, "If set to true, DisableRain will disable the in-game rain effect.\r\nNote: this doesn't have an effect on frame rate.")

	flag.BoolVar(&ModifyWndProc, "ModifyWndProc", false, "If set to true, ModifyWndProc will change (NOP) TWOM's WndProc WM_SIZE.\r\nWhenever TWOM is minimized, and then reopened, there is about a 2s black screen before the game can show anything (1s because of kernel32's 1000ms sleep, and another one rendering everything), ModifyWndProc will NOP the 'if' condition to WM_SIZE, and make the 2s process near instantaneous\r\nThis comes at a downside, as attempting to resize the game from anything lower than 100% resolution back to 100% makes everything low resolution (NOTE: It's possible to change resolution back to 100%, it needs to be done from the settings menu.)\r\n other than that, there are no other downsides found.")

	flag.BoolVar(&PrintErr, "PrintErr", false, "If set to true, PrintErr will print any error that comes up.\r\nHowever, it can be quite spammy.")

	flag.Parse()

	Step = float32(tmpFloat64)

	if Step <= 0 {
		Step = 0.7
	}

	fmt.Printf("Step: %v | CheckFG: %v | ReadMem: %v | FixCam: %v | PrintErr: %v\r\n", Step, PCheckFG, PReadMem, PFixCam, PrintErr)

	var entry syscall.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(syscall.ProcessEntry32{}))

	snhandle, err := memory.CreateToolhelp32Snapshot()
	if err.(syscall.Errno) != 0 && PrintErr {
		fmt.Println("[0] CreateToolhelp32Snapshot ->", err)
	}

	ok, snapshot, err := memory.Process32FirstW(snhandle, entry)
	if ok {
		for {
			okproc, process, err := memory.Process32NextW(snhandle, snapshot)
			if !okproc {
				break
			}
			if err.(syscall.Errno) != 0 && PrintErr {
				fmt.Println("[0] Process32NextW ->", err)
			}
			if strings.Contains(syscall.UTF16ToString(process.ExeFile[:]), "This War of Mine.exe") {
				TWOMPID = process.ProcessID
				break
			}
		}
	}
	if err.(syscall.Errno) != 0 && PrintErr {
		fmt.Println("[0] Process32FirstW ->", err)
	}

	fmt.Println("This War of Mine PID ->", TWOMPID)

	HANDLE, err := memory.OpenProcess(TWOMPID)
	if err.(syscall.Errno) != 0 && PrintErr {
		fmt.Println("[0] OpenProcess ->", err)
	}

	Modules, err := memory.EnumProcessModules(HANDLE)
	if err.(syscall.Errno) != 0 && PrintErr {
		fmt.Println("[0] EnumProcessModules ->", err)
	}

	for i := 0; i < len(Modules); i++ {
		ModuleFileName, err := memory.GetModuleFileNameExW(HANDLE, Modules[i])
		if err.(syscall.Errno) != 0 && PrintErr {
			fmt.Println("[0] GetModuleFileNameExW ->", err)
		}
		if strings.Contains(ModuleFileName, "This War of Mine.exe") {
			ModuleInfo, err := memory.GetModuleInformation(HANDLE, Modules[i])
			if err.(syscall.Errno) != 0 && PrintErr {
				fmt.Println("[0] GetModuleInformation ->", err)
			}
			BaseAddr = int64(ModuleInfo.LpBaseOfDll)
			break
		}
	}

	XPos = memory.Offsets(HANDLE, uintptr(BaseAddr), 0x009064D0, 0x70)
	XBuffer, err = memory.ReadProcessMemory(HANDLE, XPos, 8)
	if err.(syscall.Errno) != 0 && PrintErr {
		fmt.Println("[0] ReadProcessMemory ->", err)
	}
	X = math.Float32frombits(binary.LittleEndian.Uint32(XBuffer))

	YPos = memory.Offsets(HANDLE, uintptr(BaseAddr), 0x009064D0, 0x78)
	YBuffer, err = memory.ReadProcessMemory(HANDLE, YPos, 8)
	if err.(syscall.Errno) != 0 && PrintErr {
		fmt.Println("[1] ReadProcessMemory ->", err)
	}
	Y = math.Float32frombits(binary.LittleEndian.Uint32(YBuffer))

	CMode = memory.Offsets(HANDLE, uintptr(BaseAddr), 0x009064D0, 0xA6)

	Pencil = uintptr(BaseAddr) + 0x24D782

	Rain = uintptr(BaseAddr) + 0x1B431F

	WndProc = uintptr(BaseAddr) + 0x4C2C31

	if DisablePEffect {
		err := memory.NOP(HANDLE, Pencil, 9)
		if err.(syscall.Errno) != 0 && PrintErr {
			fmt.Println("[0] NOP ->", err)
		}
	}

	if DisableRainEffect {
		err := memory.NOP(HANDLE, Rain, 7)
		if err.(syscall.Errno) != 0 && PrintErr {
			fmt.Println("[1] NOP ->", err)
		}
	}

	if ModifyWndProc {
		err := memory.NOP(HANDLE, WndProc, 5)
		if err.(syscall.Errno) != 0 && PrintErr {
			fmt.Println("[2] NOP ->", err)
		}
	}

	if PCheckFG {
		go CheckFG()
	}

	if PReadMem {
		go ReadMem(HANDLE, &Mutex)
	}

	if PFixCam {
		go FixCam(HANDLE, &Mutex)
	}

	for {
		<-S.C
		Mutex.Lock()
		if <-CwMem {
			// W
			if ok, err := memory.GetAsyncKeyState(0x57); ok && err.(syscall.Errno) == 0 {
				Y += Step
				err := memory.WriteProcessMemoryFloat(HANDLE, YPos, Y)
				if err.(syscall.Errno) != 0 && PrintErr {
					fmt.Println("[0] WriteProcessMemoryFloat ->", err)
				}
			} else {
				if err.(syscall.Errno) != 0 && PrintErr {
					fmt.Println("[0] GetAsyncKeyState ->", err)
				}
			}

			// A
			if ok, err := memory.GetAsyncKeyState(0x41); ok && err.(syscall.Errno) == 0 {
				X -= Step
				err := memory.WriteProcessMemoryFloat(HANDLE, XPos, X)
				if err.(syscall.Errno) != 0 && PrintErr {
					fmt.Println("[1] WriteProcessMemoryFloat ->", err)
				}
			} else {
				if err.(syscall.Errno) != 0 && PrintErr {
					fmt.Println("[1] GetAsyncKeyState ->", err)
				}
			}

			// S
			if ok, err := memory.GetAsyncKeyState(0x53); ok && err.(syscall.Errno) == 0 {
				Y -= Step
				err := memory.WriteProcessMemoryFloat(HANDLE, YPos, Y)
				if err.(syscall.Errno) != 0 && PrintErr {
					fmt.Println("[2] WriteProcessMemoryFloat ->", err)
				}
			} else {
				if err.(syscall.Errno) != 0 && PrintErr {
					fmt.Println("[2] GetAsyncKeyState ->", err)
				}
			}

			// D
			if ok, err := memory.GetAsyncKeyState(0x44); ok && err.(syscall.Errno) == 0 {
				X += Step
				err := memory.WriteProcessMemoryFloat(HANDLE, XPos, X)
				if err.(syscall.Errno) != 0 && PrintErr {
					fmt.Println("[3] WriteProcessMemoryFloat ->", err)
				}
			} else {
				if err.(syscall.Errno) != 0 && PrintErr {
					fmt.Println("[3] GetAsyncKeyState ->", err)
				}
			}
		}
		Mutex.Unlock()
		if !S.Stop() {
			S.Reset(10 * time.Millisecond)
		}
	}
}

func CheckFG() {
	for {
		HWND, err := memory.GetForegroundWindow()
		if err.(syscall.Errno) != 0 && PrintErr {
			fmt.Println("[0] GetForegroundWindow ->", err)
		}
		if pid, err := memory.GetWindowThreadProcessId(HWND); pid == TWOMPID && err.(syscall.Errno) == 0 {
			CwMem <- true
		} else {
			CwMem <- false
			if err.(syscall.Errno) != 0 && PrintErr {
				fmt.Println("[0] GetWindowThreadProcessId ->", err)
			}
		}
	}
}

func ReadMem(HANDLE uintptr, Mutex *sync.Mutex) {
	for {
		<-S.C
		Mutex.Lock()

		XBuffer, err := memory.ReadProcessMemory(HANDLE, XPos, 8)
		if err.(syscall.Errno) != 0 && PrintErr {
			fmt.Println("[2] ReadProcessMemory ->", err)
		}

		X = math.Float32frombits(binary.LittleEndian.Uint32(XBuffer))

		YBuffer, err = memory.ReadProcessMemory(HANDLE, YPos, 8)
		if err.(syscall.Errno) != 0 && PrintErr {
			fmt.Println("[3] ReadProcessMemory ->", err)
		}
		Y = math.Float32frombits(binary.LittleEndian.Uint32(YBuffer))

		Mutex.Unlock()
		if !S.Stop() {
			S.Reset(10 * time.Millisecond)
		}
	}
}

func FixCam(HANDLE uintptr, Mutex *sync.Mutex) {
	for {
		<-S.C
		Mutex.Lock()
		CMBuffer, err := memory.ReadProcessMemory(HANDLE, CMode, 4)
		if err.(syscall.Errno) != 0 && PrintErr {
			fmt.Println("[4] ReadProcessMemory ->", err)
		}
		CM = binary.LittleEndian.Uint32(CMBuffer)

		if CM != 148602 {
			err := memory.WriteProcessMemoryInt(HANDLE, CMode, 148602)
			if err.(syscall.Errno) != 0 && PrintErr {
				fmt.Println("[0] WriteProcessMemoryInt ->", err)
			}
		}
		Mutex.Unlock()
		if !S.Stop() {
			S.Reset(10 * time.Millisecond)
		}
	}
}
