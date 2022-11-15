package pmemory

import (
	"encoding/binary"
	"fmt"
	"syscall"
	"unsafe"
)

var (
	// this should be wine compatible, but we must be able to find all DLLs & processes
	// otherwise we panic. (there is no reason to continue if we can't find one of them.)
	user32                    = syscall.MustLoadDLL("user32.dll")
	pGetAsyncKeyState         = user32.MustFindProc("GetAsyncKeyState")
	pGetForegroundWindow      = user32.MustFindProc("GetForegroundWindow")
	PGetWindowThreadProcessId = user32.MustFindProc("GetWindowThreadProcessId")

	kernel32            = syscall.MustLoadDLL("kernel32.dll")
	pReadProcessMemory  = kernel32.MustFindProc("ReadProcessMemory")
	pOpenProcess        = kernel32.MustFindProc("OpenProcess")
	pWriteProcessMemory = kernel32.MustFindProc("WriteProcessMemory")
	pOutputDebugStringW = kernel32.MustFindProc("OutputDebugStringW")

	pCreateToolhelp32Snapshot = kernel32.MustFindProc("CreateToolhelp32Snapshot")
	pProcess32FirstW          = kernel32.MustFindProc("Process32FirstW")
	pProcess32NextW           = kernel32.MustFindProc("Process32NextW")

	PSAPI                 = syscall.MustLoadDLL("psapi.dll")
	pEnumProcessModules   = PSAPI.MustFindProc("EnumProcessModules")
	pGetModuleFileNameExW = PSAPI.MustFindProc("GetModuleFileNameExW")
	pGetModuleInformation = PSAPI.MustFindProc("GetModuleInformation")

	// https://learn.microsoft.com/en-us/windows/win32/procthread/process-security-and-access-rights
	PROCESS_ALL_ACCESS = 0xFFFF
)

// https://learn.microsoft.com/en-us/windows/win32/api/memoryapi/nf-memoryapi-readprocessmemory
func ReadProcessMemory(hProcess, lpBaseAddress, nSize uintptr) (Buffer []uint8) {
	var nRead int
	buffer := make([]uint8, nSize)
	_, _, err := pReadProcessMemory.Call(
		// [in]  HANDLE  hProcess
		hProcess,
		// [in]  LPCVOID lpBaseAddress
		lpBaseAddress,
		// [out] LPVOID  lpBuffer
		uintptr(unsafe.Pointer(&buffer[0])),
		// [in]  SIZE_T  nSize
		nSize,
		// [out] SIZE_T  *lpNumberOfBytesRead
		uintptr(unsafe.Pointer(&nRead)),
	)
	if err.(syscall.Errno) != 0 {
		OutputDebugStringW(fmt.Sprintf("ReadProcessMemory -> %v", err.Error()))
	}
	return buffer

}

// https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-openprocess
func OpenProcess(pid uint32) (HANDLE uintptr) {
	v, _, err := pOpenProcess.Call(
		//[in] DWORD dwDesiredAccess
		uintptr(PROCESS_ALL_ACCESS),
		//[in] BOOL  bInheritHandle
		uintptr(0),
		//[in] DWORD dwProcessId
		uintptr(pid),
	)
	if err.(syscall.Errno) != 0 {
		OutputDebugStringW(fmt.Sprintf("OpenProcess -> %v", err.Error()))
	}
	return v
}

// https://learn.microsoft.com/en-us/windows/win32/api/psapi/nf-psapi-enumprocessmodules
func EnumProcessModules(HANDLE uintptr) (Modules []uintptr) {

	// This is only ran once, so it doesn't matter if we get a lot of repeat modules
	// "It is a good idea to specify a large array of HMODULE values, because it is hard to predict how many modules there will be in the process at the time you call EnumProcessModules."
	var lphModule [256]uintptr
	var lpcbNeeded uint32
	_, _, err := pEnumProcessModules.Call(
		//[in]  HANDLE  hProcess
		HANDLE,
		//[out] HMODULE *lphModule
		uintptr(unsafe.Pointer(&lphModule)),
		//[in]  DWORD   cb
		uintptr(len(lphModule)),
		//[out] LPDWORD lpcbNeeded
		uintptr(unsafe.Pointer(&lpcbNeeded)),
	)
	if err.(syscall.Errno) != 0 {
		OutputDebugStringW(fmt.Sprintf("EnumProcessModules -> %v", err.Error()))
	}
	return lphModule[:]
}

// https://learn.microsoft.com/en-us/windows/win32/api/psapi/nf-psapi-getmodulefilenameexw
func GetModuleFileNameExW(HANDLE, Module uintptr) (ModuleFileName string) {
	// https://learn.microsoft.com/en-us/windows/win32/fileio/maximum-file-path-limitation
	// tldr; 260 is the max length of a windows path
	var buffer [260]uint16
	_, _, err := pGetModuleFileNameExW.Call(
		//	[in]  HANDLE  hProcess,
		HANDLE,
		//  [in, optional] HMODULE hModule,
		Module,
		//  [out] LPWSTR  lpFilename,
		uintptr(unsafe.Pointer(&buffer)),
		//  [in]  DWORD   nSize
		uintptr(uint32(len(buffer))),
	)
	if err.(syscall.Errno) != 0 {
		OutputDebugStringW(fmt.Sprintf("GetModuleFileNameExW -> %v", err.Error()))
	}
	return syscall.UTF16ToString(buffer[:])
}

// https://learn.microsoft.com/en-us/windows/win32/api/psapi/nf-psapi-getmoduleinformation
func GetModuleInformation(HANDLE, MODULE uintptr) (MODULEINF MODULEINFO) {
	var MODINFO MODULEINFO

	_, _, err := pGetModuleInformation.Call(
		HANDLE,
		MODULE,
		uintptr(unsafe.Pointer(&MODINFO)),
		uintptr(unsafe.Sizeof(MODINFO)),
	)
	if err.(syscall.Errno) != 0 {
		OutputDebugStringW(fmt.Sprintf("GetModuleInformation -> %v", err.Error()))
	}

	return MODINFO
}

// https://learn.microsoft.com/en-us/windows/win32/api/memoryapi/nf-memoryapi-writeprocessmemory
func WriteProcessMemoryFloat(HANDLE, BaseAddr uintptr, ToWrite float32) {
	written := 0

	_, _, err := pWriteProcessMemory.Call(
		// [in]  HANDLE  hProcess,
		HANDLE,
		// [in]  LPVOID  lpBaseAddress,
		BaseAddr,
		// [in]  LPCVOID lpBuffer,
		// Note: any and/or interface{} will not work, it will give back a broken result.
		// So float32 has to be specified.
		uintptr(unsafe.Pointer(&ToWrite)),
		// [in]  SIZE_T  nSize,
		uintptr(unsafe.Sizeof(ToWrite)),
		// [out] SIZE_T  *lpNumberOfBytesWritten
		uintptr(unsafe.Pointer(&written)),
	)
	if err.(syscall.Errno) != 0 {
		OutputDebugStringW(fmt.Sprintf("WriteProcessMemoryFloat -> %v", err.Error()))
	}
}
func WriteProcessMemory(HANDLE, BaseAddr uintptr, ToWrite []byte) {
	written := 0

	_, _, err := pWriteProcessMemory.Call(
		// [in]  HANDLE  hProcess,
		HANDLE,
		// [in]  LPVOID  lpBaseAddress,
		BaseAddr,
		// [in]  LPCVOID lpBuffer,
		// Note: any and/or interface{} will not work, it will give back a broken result.
		// We use a byte array here.
		uintptr(unsafe.Pointer(&ToWrite[0])),
		// [in]  SIZE_T  nSize,
		uintptr(len(ToWrite)),
		// [out] SIZE_T  *lpNumberOfBytesWritten
		uintptr(unsafe.Pointer(&written)),
	)
	if err.(syscall.Errno) != 0 {
		OutputDebugStringW(fmt.Sprintf("WriteProcessMemory -> %v", err.Error()))
	}
}

func NOP(HANDLE, BaseAddr uintptr, HowMany int) {
	written := 0

	var NopArray []byte

	for i := 0; i < HowMany; i++ {
		NopArray = append(NopArray, 0x90)
	}

	_, _, err := pWriteProcessMemory.Call(
		// [in]  HANDLE  hProcess,
		HANDLE,
		// [in]  LPVOID  lpBaseAddress,
		BaseAddr,
		// [in]  LPCVOID lpBuffer,
		uintptr(unsafe.Pointer(&NopArray[0])),
		// [in]  SIZE_T  nSize,
		uintptr(HowMany),
		// [out] SIZE_T  *lpNumberOfBytesWritten
		uintptr(unsafe.Pointer(&written)),
	)
	if err.(syscall.Errno) != 0 {
		OutputDebugStringW(fmt.Sprintf("NOP -> %v", err.Error()))
	}
}

// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getasynckeystate
//
// https://learn.microsoft.com/en-us/windows/win32/inputdev/virtual-key-codes
func GetAsyncKeyState(vKey int) (ok bool) {
	//[in] int vKey
	if v, _, err := pGetAsyncKeyState.Call(uintptr(vKey)); v != 0 && err.(syscall.Errno) == 0 {
		return true
	} else {
		return false
	}
}

type MODULEINFO struct {
	LpBaseOfDll uintptr
	SizeOfImage uint32
	EntryPoint  uintptr
}

func Offsets(HANDLE uintptr, BaseAddr uintptr, BaseOffset uintptr, Offsets ...uintptr) uintptr {
	BaseThing := BaseAddr + BaseOffset
	var Buffer []uint8
	var Pointer uint64
	for i := 0; i < len(Offsets); i++ {
		if i == 0 {
			Buffer = ReadProcessMemory(HANDLE, uintptr(BaseThing), 8)
			Pointer = binary.LittleEndian.Uint64(Buffer)
			Pointer += uint64(Offsets[i])
			fmt.Printf("%x\n", Pointer)
		} else {
			Buffer = ReadProcessMemory(HANDLE, uintptr(Pointer), 8)
			Pointer = binary.LittleEndian.Uint64(Buffer)
			Pointer += uint64(Offsets[i])
			fmt.Printf("%x\n", Pointer)
		}
	}
	return uintptr(Pointer)
}

// https://learn.microsoft.com/en-us/windows/win32/api/tlhelp32/nf-tlhelp32-createtoolhelp32snapshot
func CreateToolhelp32Snapshot() (snapshot uintptr) {
	v, _, err := pCreateToolhelp32Snapshot.Call(
		//[in] DWORD dwFlags,
		uintptr(0x00000002),
		//[in] DWORD th32ProcessID
		uintptr(0),
	)
	if err.(syscall.Errno) != 0 {
		OutputDebugStringW(fmt.Sprintf("CreateToolhelp32Snapshot -> %v", err.Error()))
	}
	return v
}

// https://learn.microsoft.com/en-us/windows/win32/api/tlhelp32/nf-tlhelp32-process32firstw
func Process32FirstW(HANDLE uintptr, PE32 syscall.ProcessEntry32) (ok bool, procentry syscall.ProcessEntry32) {
	v, _, err := pProcess32FirstW.Call(
		// [in]      HANDLE           hSnapshot,
		HANDLE,
		// [in, out] LPPROCESSENTRY32 lppe
		uintptr(unsafe.Pointer(&PE32)),
	)
	if err.(syscall.Errno) != 0 {
		OutputDebugStringW(fmt.Sprintf("Process32FirstW -> %v", err.Error()))
	}
	if v == 1 {
		return true, PE32
	}
	return false, PE32
}

// https://learn.microsoft.com/en-us/windows/win32/api/tlhelp32/nf-tlhelp32-process32nextw
func Process32NextW(HANDLE uintptr, PE32 syscall.ProcessEntry32) (ok bool, procentry syscall.ProcessEntry32) {
	v, _, err := pProcess32NextW.Call(
		// [in]  HANDLE           hSnapshot,
		HANDLE,
		// [out] LPPROCESSENTRY32 lppe
		uintptr(unsafe.Pointer(&PE32)),
	)
	if err.(syscall.Errno) != 0 {
		OutputDebugStringW(fmt.Sprintf("Process32NextW -> %v", err.Error()))
	}
	if v == 1 {
		return true, PE32
	}
	return false, PE32
}

// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getforegroundwindow
func GetForegroundWindow() (HWND uintptr) {
	v, _, _ := pGetForegroundWindow.Call()
	return v
}

// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getwindowthreadprocessid
func GetWindowThreadProcessId(HWND uintptr) (pid uint32) {
	var PID uint32
	_, _, err := PGetWindowThreadProcessId.Call(
		// [in]            HWND    hWnd,
		HWND,
		// [out, optional] LPDWORD lpdwProcessId
		uintptr(unsafe.Pointer(&PID)),
	)
	if err.(syscall.Errno) != 0 {
		OutputDebugStringW(fmt.Sprintf("GetWindowThreadProcessId -> %v", err.Error()))
	}
	return PID
}

// https://learn.microsoft.com/en-us/windows/win32/api/debugapi/nf-debugapi-outputdebugstringw
func OutputDebugStringW(lpOutputString string) {
	ptr, _ := syscall.UTF16PtrFromString(lpOutputString)
	pOutputDebugStringW.Call(uintptr(unsafe.Pointer(&ptr)))
}
