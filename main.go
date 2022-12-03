package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	memory "main/memory"
	"math"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	Abandoned_Cottage = iota
	Shelled_Cottage
	Ruined_Block_of_Flats
	Looted_Gas_Station
	Ghost_House
	Garage
	Small_Apartment_Building
	Decrepit_Squad
	Brothel
	Shelled_School
	Warehouse
	Ruined_Villa
	Semi_Detached_House
	Military_Outpost
	Hotel
	Construction_Site
	Quiet_House
	Supermarket
	Sniper_Junction
	St_Marys_Church
	City_Hospital
	Old_Town
	Port
	Airport
	Pharmacy
	Ruined_Toy_Store
	Park
	Bakery
	Shelled_Brewery
	The_Samuel_Institute
	Railway_Station
	Music_Club
)

var (
	Step float32 = 0.7

	UseWASD           bool = false
	PReadMem          bool = false
	PFixCam           bool = false
	DisablePEffect    bool = false
	DisableRainEffect bool = false
	DisableOutlines   bool = false
	ModifyWndProc     bool = false

	WASDChan    = make(chan struct{})
	ReadMemChan = make(chan struct{})
	FixCamChan  = make(chan struct{})

	TWOMPID  uint32
	BaseAddr int64

	HANDLE, XPos, YPos, CMode, Rain, Pencil, WndProc, Outline uintptr
	XBuffer, YBuffer, CMBuffer                                []byte
	X, Y                                                      float32
	CM                                                        uint32

	// For Custom Scenarios
	LocationsPTR      uintptr
	LocationsAddr     []uintptr
	LocationsToChoose int
	LocationsAdded    = make(map[int]uintptr)
	LocationsData     = map[int]int{
		Abandoned_Cottage:        1,
		Shelled_Cottage:          1,
		Ruined_Block_of_Flats:    1,
		Looted_Gas_Station:       1,
		Ghost_House:              1,
		Garage:                   1,
		Small_Apartment_Building: 2,
		Decrepit_Squad:           1,
		Brothel:                  1,
		Shelled_School:           2,
		Warehouse:                1,
		Ruined_Villa:             2,
		Semi_Detached_House:      2,
		Military_Outpost:         1,
		Hotel:                    3,
		Construction_Site:        2,
		Quiet_House:              1,
		Supermarket:              2,
		Sniper_Junction:          2,
		St_Marys_Church:          2,
		City_Hospital:            1,
		Old_Town:                 1,
		Port:                     1,
		Airport:                  1,
		Pharmacy:                 1,
		Ruined_Toy_Store:         1,
		Park:                     1,
		Bakery:                   1,
		Shelled_Brewery:          1,
		The_Samuel_Institute:     1,
		Railway_Station:          1,
		Music_Club:               2,
	}

	Ceasefirep, Intensityp, WinterComesp, WinterHarshnessp, WinterLengthp uintptr
	Ceasefire, Intensity, WinterComes, WinterHarshness, WinterLength      int

	ExitRegex  = regexp.MustCompile(`(?mi)(exit|quit)`)
	ClearRegex = regexp.MustCompile(`(?mi)(clear|cls)`)

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

	// f3 44 0f 10 05 35 10 4a 00
	// MOVSS XMM8,[TWOM+6EE7C0]
	Pencil = uintptr(BaseAddr) + 0x24D782

	// 0f b7 05 b8 d7 72 00
	// MOVZX EAX,word ptr [TWOM+8E1ADE]
	Rain = uintptr(BaseAddr) + 0x1B431F

	//	83 e8 02
	//	SUB EAX, 0x2
	//	74 26
	//	JZ TWOM+4C2C5C
	WndProc = uintptr(BaseAddr) + 0x4C2C31

	//	f3 0f 10 0d d8 2f 60 00
	//	MOVSS XMM1, dword ptr [TWOM+819490]
	Outline = uintptr(BaseAddr) + 0x2164B0

	PrintMenu()
}

func PrintPatches() {
	fmt.Println()
	fmt.Println("This War of Mine Utils -> Patches")
	fmt.Println("The following options will patch the game executable itself.")
	fmt.Println("Whilst it shouldn't happen, make sure you have a backup of your savefiles in case of a corruption caused by the patched exe.")
	fmt.Println()

	// pretty pointless for now.. maybe i'll do something with it later
	fmt.Println("allow-more-instances) Will patch the game to allow more than 1 running instance")
	fmt.Println("cls|clear) Clear Screen")
	fmt.Println("leave) Go Back to The Normal Menu")

	var Option string

	fmt.Scan(&Option)

	switch Option {
	case ClearRegex.FindString(Option):
		ClearScreen()
		PrintPatches()
	case "allow-more-instances":
		var ExePath string

		fmt.Print("Please specify where This War of Mine.exe is (or just drag the exe here) -> ")

		inreader := bufio.NewScanner(os.Stdin)

		for {
			inreader.Scan()
			if ExePath = inreader.Text(); ExePath != "" {
				break
			}
		}

		ExePath = strings.Replace(ExePath, "\"", "", -1)

		twombytes, err := os.ReadFile(ExePath)
		if err != nil {
			fmt.Println("Failed to Read File -> ", err)
		}
		fmt.Println("File Read OK")

		Addr := memory.ScanBytes(twombytes, []byte{0x48, 0x85, 0xC0, 0x74, 0x3A, 0x48, 0x8B, 0xC8, 0xFF, 0x15, 0x08, 0x33, 0x0B, 0x00})
		if Addr == 0 {
			fmt.Println("Failed to Find Bytes")
		} else {
			twombytes[Addr+3] = 0xEB
		}

		errw := os.WriteFile("twom-mi.exe", twombytes, 0666)
		if errw != nil {
			fmt.Println("Failed to Write File -> ", errw)
		}

		if err == nil && errw == nil {
			fmt.Println("twom-mi.exe was written successfully. leaving patches menu.")
		} else {
			fmt.Println("Something failed while attempting to patch the exe, leaving patches menu.")
		}
	case "leave":
		ClearScreen()
		PrintMenu()
	default:
		fmt.Println(Option, "is not a valid option.")
	}
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

	if PReadMem {
		fmt.Println("\033[92m[on]\033[0m\t3) Read & Write game memory to a stored one")
	} else {
		fmt.Println("\033[91m[off]\033[0m\t3) Read & Write game memory to a stored one")
	}

	if PFixCam {
		fmt.Println("\033[92m[on]\033[0m\t4) Fix Camera for WASD Controls")
	} else {
		fmt.Println("\033[91m[off]\033[0m\t4) Fix Camera for WASD Controls")
	}

	if DisablePEffect {
		fmt.Println("\033[92m[on]\033[0m\t5) Hide Pencil Effect")
	} else {
		fmt.Println("\033[91m[off]\033[0m\t5) Hide Pencil Effect")
	}

	if DisableRainEffect {
		fmt.Println("\033[92m[on]\033[0m\t6) Hide Rain")
	} else {
		fmt.Println("\033[91m[off]\033[0m\t6) Hide Rain")
	}

	if DisableOutlines {
		fmt.Println("\033[92m[on]\033[0m\t7) Hide Outlines")
	} else {
		fmt.Println("\033[91m[off]\033[0m\t7) Hide Outlines")
	}

	if ModifyWndProc {
		fmt.Println("\033[92m[on]\033[0m\t8) Modify WndProc")
	} else {
		fmt.Println("\033[91m[off]\033[0m\t8) Modify WndProc")
	}

	fmt.Println()
	fmt.Println("----------------------------------------")
	fmt.Println()

	fmt.Println("i) Randomize Settings\t(Custom Scenario)")
	fmt.Println("o) Randomize Locations\t(Custom Scenario)")

	fmt.Println("0) Change Step Value ( now", Step, ")")

	fmt.Println("a) Toggle All")
	fmt.Println("p) Patches")
	fmt.Println("q) Inject TWOMUHook")
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

		Option = strings.ToLower(Option)

		switch Option {
		case ExitRegex.FindString(Option):
			os.Exit(0)
		case ClearRegex.FindString(Option):
			ClearScreen()
			PrintMenu()
		case "1":
			GetTWOM()
			PrintMenu()
		case "2":
			if !UseWASD {
				go WASDControls()
			} else {
				WASDChan <- struct{}{}
			}
			UseWASD = !UseWASD
			ClearScreen()
			PrintMenu()
		case "3":
			if !PReadMem {
				go ReadMem()
			} else {
				ReadMemChan <- struct{}{}
			}
			PReadMem = !PReadMem
			ClearScreen()
			PrintMenu()
		case "4":
			if !PFixCam {
				go FixCam()
			} else {
				FixCamChan <- struct{}{}
			}
			PFixCam = !PFixCam
			ClearScreen()
			PrintMenu()
		case "5":
			if !DisablePEffect {
				memory.NOP(HANDLE, Pencil, 9)
			} else {
				memory.WriteProcessMemory(HANDLE, Pencil, []byte{0xF3, 0x44, 0x0F, 0x10, 0x05, 0x35, 0x10, 0x4A, 0x00}, 9)
			}
			DisablePEffect = !DisablePEffect
			ClearScreen()
			PrintMenu()
		case "6":
			if !DisableRainEffect {
				memory.NOP(HANDLE, Rain, 7)
			} else {
				memory.WriteProcessMemory(HANDLE, Rain, []byte{0x0F, 0xB7, 0x05, 0xB8, 0xD7, 0x72, 0x00}, 7)
			}
			DisableRainEffect = !DisableRainEffect
			ClearScreen()
			PrintMenu()
		case "7":
			if !DisableOutlines {
				memory.NOP(HANDLE, Outline, 8)
			} else {
				memory.WriteProcessMemory(HANDLE, Outline, []byte{0xF3, 0x0F, 0x10, 0x0D, 0xD8, 0x2F, 0x60, 0x00}, 8)
			}
			DisableOutlines = !DisableOutlines
			ClearScreen()
			PrintMenu()
		case "8":
			if !ModifyWndProc {
				memory.NOP(HANDLE, WndProc, 5)
			} else {
				memory.WriteProcessMemory(HANDLE, WndProc, []byte{0x83, 0xE8, 0x02, 0x74, 0x26}, 5)
			}
			ModifyWndProc = !ModifyWndProc
			ClearScreen()
			PrintMenu()
		case "a":
			if !UseWASD {
				go WASDControls()
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
		case "i":
			Ceasefirep = memory.Offsets(HANDLE, uintptr(BaseAddr), 0x008A7998, 0x4B0, 0x08, 0x10, 0x30, 0x00, 0x34)
			Intensityp = memory.Offsets(HANDLE, uintptr(BaseAddr), 0x008A7998, 0x4B0, 0x08, 0x10, 0x30, 0x08, 0x34)
			WinterComesp = memory.Offsets(HANDLE, uintptr(BaseAddr), 0x008A7998, 0x4B0, 0x08, 0x10, 0x30, 0x10, 0x34)
			WinterHarshnessp = memory.Offsets(HANDLE, uintptr(BaseAddr), 0x008A7998, 0x4B0, 0x08, 0x10, 0x30, 0x18, 0x34)
			WinterLengthp = memory.Offsets(HANDLE, uintptr(BaseAddr), 0x008A7998, 0x4B0, 0x08, 0x10, 0x30, 0x20, 0x34)

			rand.Seed(time.Now().UnixNano())
			Ceasefire = rand.Intn(13)
			Intensity = rand.Intn(3)
			WinterComes = 1 + rand.Intn(3)
			WinterHarshness = rand.Intn(3)
			WinterLength = rand.Intn(3)

			LocationsToChoose = 8 + Ceasefire

			memory.WriteProcessMemory(HANDLE, Ceasefirep, []byte{byte(Ceasefire)}, 1)
			memory.WriteProcessMemory(HANDLE, Intensityp, []byte{byte(Intensity)}, 1)
			memory.WriteProcessMemory(HANDLE, WinterComesp, []byte{byte(WinterComes)}, 1)
			memory.WriteProcessMemory(HANDLE, WinterHarshnessp, []byte{byte(WinterHarshness)}, 1)
			memory.WriteProcessMemory(HANDLE, WinterLengthp, []byte{byte(WinterLength)}, 1)
			PrintMenu()
		case "o":
			LocationsPTR = memory.Offsets(HANDLE, uintptr(BaseAddr), 0x008A7998, 0x800, 0x220, 0x00, 0x00, 0x120, 0x10)

			LocationsAddr = nil
			for k := range LocationsAdded {
				delete(LocationsAdded, k)
			}

			LocationsAddr = append(LocationsAddr, LocationsPTR)
			for i := 0; i < 31; i++ {
				LocationsAddr = append(LocationsAddr, LocationsAddr[len(LocationsAddr)-1]+56)
			}

			rand.Seed(time.Now().UnixNano())
			for i := 0; i < LocationsToChoose; i++ {
				id := rand.Intn(len(LocationsAddr))
				if LocationsAdded[id] != 0 {
					i--
				} else {
					LocationsAdded[id] = LocationsAddr[id]
				}
			}

			for id, pointer := range LocationsAdded {
				memory.WriteProcessMemory(HANDLE, pointer, []byte{byte(rand.Intn(LocationsData[id])), 0x00, 0x00, 0x00}, 4)
			}
			PrintMenu()
		case "p":
			ClearScreen()
			PrintPatches()
		case "q":
			fmt.Println("Attempting to Inject TWOMUHook")
			path, err := os.Getwd()
			if err != nil {
				fmt.Println("Failed to get Current Directory")
			}

			dllPath := path + "\\TWOMUHook\\x64\\TWOMUHook.dll"

			Valloc := memory.VirtualAllocEx(HANDLE, 0, uintptr(len(dllPath)+1), 0x00002000|0x00001000, 4)

			memory.WriteProcessMemory(HANDLE, Valloc, []byte(dllPath), uintptr(len(dllPath)+1))

			modKernel, err := syscall.LoadLibrary("kernel32.dll")
			if err != nil {
				memory.OutputDebugStringW(err.Error())
			}

			LoadLibrary, err := syscall.GetProcAddress(modKernel, "LoadLibraryA")
			if err != nil {
				memory.OutputDebugStringW(err.Error())
			}

			memory.CreateRemoteThread(HANDLE, 0, 0, LoadLibrary, Valloc, 0)
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
			if pid := memory.GetWindowThreadProcessId(memory.GetForegroundWindow()); pid == TWOMPID {
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
			CMBuffer = memory.ReadProcessMemory(HANDLE, CMode, 4)
			CM = binary.LittleEndian.Uint32(CMBuffer)

			if CM != 148602 {
				memory.WriteProcessMemory(HANDLE, CMode, []byte{0x7A, 0x44, 0x02}, 3)
			}
			if !S.Stop() {
				S.Reset(10 * time.Millisecond)
			}
		}
	}
}
