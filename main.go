package main

import (
	"encoding/binary"
	"fmt"
	memory "main/memory"
	"math"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var (
	Step float32 = 0.7

	UseWASD           bool = false
	PCheckFG          bool = false
	PReadMem          bool = false
	PFixCam           bool = false
	DisablePEffect    bool = false
	DisableRainEffect bool = false
	DisableOutlines   bool = false
	ModifyWndProc     bool = false

	WASDChan    = make(chan struct{})
	FGChan      = make(chan struct{})
	ReadMemChan = make(chan struct{})
	FixCamChan  = make(chan struct{})

	TWOMPID  uint32
	BaseAddr int64

	HANDLE, XPos, YPos, CMode, Rain, Pencil, WndProc, Outline uintptr
	XBuffer, YBuffer, CMBuffer                                []uint8
	X, Y                                                      float32
	CM                                                        uint32

	CwMem = make(chan bool)
	Mutex sync.Mutex

	S = time.NewTimer(10 * time.Millisecond)
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

	HANDLE = memory.OpenProcess(TWOMPID)

	Modules := memory.EnumProcessModules(HANDLE)

	for i := 0; i < len(Modules); i++ {
		ModuleFileName := memory.GetModuleFileNameExW(HANDLE, Modules[i])
		if strings.Contains(ModuleFileName, "This War of Mine.exe") {
			ModuleInfo := memory.GetModuleInformation(HANDLE, Modules[i])
			BaseAddr = int64(ModuleInfo.LpBaseOfDll)
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

	Pencil = uintptr(BaseAddr) + 0x24D782

	Rain = uintptr(BaseAddr) + 0x1B431F

	WndProc = uintptr(BaseAddr) + 0x4C2C31

	Outline = uintptr(BaseAddr) + 0x2164B0

	fmt.Println("Checking if everything works. . . . . . .")
	if TWOMPID != 0 {
		fmt.Println("TWOM's PID was found!")
	} else {
		fmt.Println("TWOM's PID could NOT be found! (Is the game running? Is an antivirus blocking CreateToolhelp32Snapshot?)")
	}

	if HANDLE != 0 {
		fmt.Println("TWOM's Handle was found!")
	} else {
		fmt.Println("TWOM's Handle could NOT be found! (Is the game running? Is an antivirus blocking OpenProcess?)")
	}

	if BaseAddr != 0 {
		fmt.Println("TWOM's Base Address was found!")
	} else {
		fmt.Println("TWOM's Base Address could NOT be found! (Is the game running? Is an antivirus blocking Module operations?)")
	}

	testBuffer := memory.ReadProcessMemory(HANDLE, WndProc, 8)
	if math.Float32frombits(binary.LittleEndian.Uint32(testBuffer)) != 0 {
		fmt.Println("Reading Process Memory Works!")
	} else {
		fmt.Println("Failed to Read Process Memory (Is the game running? Is an antivirus blocking ReadProcessMemory?)")
	}
	fmt.Println("Finished.")
}

func PrintMenu() {
	fmt.Println()
	fmt.Println("This War of Mine Utils")
	fmt.Println("1) Find Handle of TWOM and calculate offsets")

	if UseWASD {
		fmt.Println("\033[92m[on]\033[0m\t2) Use WASD Controls")
	} else {
		fmt.Println("\033[91m[off]\033[0m\t2) Use WASD Controls")
	}

	if PCheckFG {
		fmt.Println("\033[92m[on]\033[0m\t3) Check if TWOM is foreground")
	} else {
		fmt.Println("\033[91m[off]\033[0m\t3) Check if TWOM is foreground")
	}

	if PReadMem {
		fmt.Println("\033[92m[on]\033[0m\t4) Read & Write game memory to a stored one")
	} else {
		fmt.Println("\033[91m[off]\033[0m\t4) Read & Write game memory to a stored one")
	}

	if PFixCam {
		fmt.Println("\033[92m[on]\033[0m\t5) Fix Camera for WASD Controls")
	} else {
		fmt.Println("\033[91m[off]\033[0m\t5) Fix Camera for WASD Controls")
	}

	if DisablePEffect {
		fmt.Println("\033[92m[on]\033[0m\t6) Show Pencil Effect")
	} else {
		fmt.Println("\033[91m[off]\033[0m\t6) Show Pencil Effect")
	}

	if DisableRainEffect {
		fmt.Println("\033[92m[on]\033[0m\t7) Show Rain")
	} else {
		fmt.Println("\033[91m[off]\033[0m\t7) Show Rain")
	}

	if DisableOutlines {
		fmt.Println("\033[92m[on]\033[0m\t8) Show Outlines")
	} else {
		fmt.Println("\033[91m[off]\033[0m\t8) Show Outlines")
	}

	if ModifyWndProc {
		fmt.Println("\033[92m[on]\033[0m\t9) Modify WndProc")
	} else {
		fmt.Println("\033[91m[off]\033[0m\t9) Modify WndProc")
	}

	fmt.Println("0) Change Step Value ( now", Step, ")")

	fmt.Println("a) Toggle All")
	fmt.Println("cls|clear) Clear Screen")
	fmt.Println("exit|quit Exit TWOMU")
}

func ClearScreen() {
	fmt.Print("\033[2J")
	fmt.Print("\033[H")
}

func main() {
	var Option string

	PrintMenu()

	for {
		fmt.Scan(&Option)

		switch Option {
		case "exit":
			os.Exit(0)
		case "quit":
			os.Exit(0)
		case "clear":
			ClearScreen()
			PrintMenu()
		case "cls":
			ClearScreen()
			PrintMenu()
		case "1":
			GetTWOM()
			PrintMenu()
		case "2":
			if !PCheckFG {
				go CheckFG()
			}
			PCheckFG = true

			if !UseWASD {
				go WASDControls()
			} else {
				WASDChan <- struct{}{}
			}
			UseWASD = !UseWASD
			ClearScreen()
			PrintMenu()
		case "3":
			if !PCheckFG {
				go CheckFG()
			} else {
				FGChan <- struct{}{}
			}
			PCheckFG = !PCheckFG
			ClearScreen()
			PrintMenu()
		case "4":
			if !PReadMem {
				go ReadMem()
			} else {
				ReadMemChan <- struct{}{}
			}
			PReadMem = !PReadMem
			ClearScreen()
			PrintMenu()
		case "5":
			if !PFixCam {
				go FixCam()
			} else {
				FixCamChan <- struct{}{}
			}
			PFixCam = !PFixCam
			ClearScreen()
			PrintMenu()
		case "6":
			if !DisablePEffect {
				memory.NOP(HANDLE, Pencil, 9)
			} else {
				memory.WriteProcessMemory(HANDLE, Pencil, []byte{0xF3, 0x44, 0x0F, 0x10, 0x05, 0x35, 0x10, 0x4A, 0x00})
			}
			DisablePEffect = !DisablePEffect
			ClearScreen()
			PrintMenu()
		case "7":
			if !DisableRainEffect {
				memory.NOP(HANDLE, Rain, 7)
			} else {
				memory.WriteProcessMemory(HANDLE, Rain, []byte{0x0F, 0xB7, 0x05, 0xB8, 0xD7, 0x72, 0x00})
			}
			DisableRainEffect = !DisableRainEffect
			ClearScreen()
			PrintMenu()
		case "8":
			if !DisableOutlines {
				memory.NOP(HANDLE, Outline, 8)
			} else {
				memory.WriteProcessMemory(HANDLE, Outline, []byte{0xF3, 0x0F, 0x10, 0x0D, 0xD8, 0x2F, 0x60, 0x00})
			}
			DisableOutlines = !DisableOutlines
			ClearScreen()
			PrintMenu()
		case "9":
			if !ModifyWndProc {
				memory.NOP(HANDLE, WndProc, 5)
			} else {
				memory.WriteProcessMemory(HANDLE, WndProc, []byte{0x83, 0xE8, 0x02, 0x74, 0x26})
			}
			ModifyWndProc = !ModifyWndProc
			ClearScreen()
			PrintMenu()
		case "a":
			if !UseWASD {
				go WASDControls()
			}
			if !PCheckFG {
				go CheckFG()
			}
			if !PReadMem {
				go ReadMem()
			}
			if !PFixCam {
				go FixCam()
			}
			if !DisablePEffect {
				memory.NOP(HANDLE, Pencil, 9)
			}
			if !DisableRainEffect {
				memory.NOP(HANDLE, Rain, 7)
			}
			if !DisableOutlines {
				memory.NOP(HANDLE, Outline, 8)
			}
			if !ModifyWndProc {
				memory.NOP(HANDLE, WndProc, 5)
			}

			UseWASD = true
			PCheckFG = true
			PReadMem = true
			PFixCam = true
			DisablePEffect = true
			DisableRainEffect = true
			DisableOutlines = true
			ModifyWndProc = true
			ClearScreen()
			PrintMenu()
		case "0":
			fmt.Print("Enter new Step Value: ")
			fmt.Scan(&Step)
			if Step <= 0 {
				Step = 0.7
			}
			ClearScreen()
			PrintMenu()
		default:
			fmt.Println(Option, "is not a valid option.")
		}
	}
}

func WASDControls() {
	for {
		select {
		case <-WASDChan:
			return
		default:
			<-S.C
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
			if !S.Stop() {
				S.Reset(10 * time.Millisecond)
			}
		}
	}
}

func CheckFG() {
	for {
		select {
		case <-FGChan:
			return
		default:
			HWND := memory.GetForegroundWindow()
			if pid := memory.GetWindowThreadProcessId(HWND); pid == TWOMPID {
				CwMem <- true
			} else {
				CwMem <- false
			}
		}
	}
}

func ReadMem() {
	for {
		select {
		case <-ReadMemChan:
			return
		default:
			<-S.C
			Mutex.Lock()

			XBuffer = memory.ReadProcessMemory(HANDLE, XPos, 8)
			X = math.Float32frombits(binary.LittleEndian.Uint32(XBuffer))

			YBuffer = memory.ReadProcessMemory(HANDLE, YPos, 8)
			Y = math.Float32frombits(binary.LittleEndian.Uint32(YBuffer))

			Mutex.Unlock()
			if !S.Stop() {
				S.Reset(10 * time.Millisecond)
			}
		}
	}
}

func FixCam() {
	for {
		select {
		case <-FixCamChan:
			return
		default:
			<-S.C
			Mutex.Lock()
			CMBuffer = memory.ReadProcessMemory(HANDLE, CMode, 4)
			CM = binary.LittleEndian.Uint32(CMBuffer)

			if CM != 148602 {
				memory.WriteProcessMemory(HANDLE, CMode, []byte{0x7A, 0x44, 0x02})
			}
			Mutex.Unlock()
			if !S.Stop() {
				S.Reset(10 * time.Millisecond)
			}
		}
	}
}
