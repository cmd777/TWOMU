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
func ReadProcessMemory(hProcess, lpBaseAddress, nSize uintptr) (Buffer []uint8, c error) {
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
	return buffer, err

}

// https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-openprocess
func OpenProcess(pid uint32) (HANDLE uintptr, c error) {
	v, _, err := pOpenProcess.Call(
		//[in] DWORD dwDesiredAccess
		uintptr(PROCESS_ALL_ACCESS),
		//[in] BOOL  bInheritHandle
		uintptr(0),
		//[in] DWORD dwProcessId
		uintptr(pid),
	)
	return v, err
}

// https://learn.microsoft.com/en-us/windows/win32/api/psapi/nf-psapi-enumprocessmodules
func EnumProcessModules(HANDLE uintptr) (Modules []uintptr, c error) {

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
	return lphModule[:], err
}

// https://learn.microsoft.com/en-us/windows/win32/api/psapi/nf-psapi-getmodulefilenameexw
func GetModuleFileNameExW(HANDLE, Module uintptr) (ModuleFileName string, c error) {
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
	return syscall.UTF16ToString(buffer[:]), err
}

// https://learn.microsoft.com/en-us/windows/win32/api/psapi/nf-psapi-getmoduleinformation
func GetModuleInformation(HANDLE, MODULE uintptr) (MODULEINF MODULEINFO, c error) {
	var MODINFO MODULEINFO

	_, _, err := pGetModuleInformation.Call(
		HANDLE,
		MODULE,
		uintptr(unsafe.Pointer(&MODINFO)),
		uintptr(unsafe.Sizeof(MODINFO)),
	)

	return MODINFO, err
}

// https://learn.microsoft.com/en-us/windows/win32/api/memoryapi/nf-memoryapi-writeprocessmemory
func WriteProcessMemoryFloat(HANDLE, BaseAddr uintptr, ToWrite float32) (c error) {
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
	return err
}
func WriteProcessMemoryInt(HANDLE, BaseAddr uintptr, ToWrite int) (c error) {
	written := 0

	_, _, err := pWriteProcessMemory.Call(
		// [in]  HANDLE  hProcess,
		HANDLE,
		// [in]  LPVOID  lpBaseAddress,
		BaseAddr,
		// [in]  LPCVOID lpBuffer,
		// Note: any and/or interface{} will not work, it will give back a broken result.
		// So int has to be specified.
		uintptr(unsafe.Pointer(&ToWrite)),
		// [in]  SIZE_T  nSize,
		uintptr(unsafe.Sizeof(ToWrite)),
		// [out] SIZE_T  *lpNumberOfBytesWritten
		uintptr(unsafe.Pointer(&written)),
	)
	return err
}

// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getasynckeystate
//
// https://learn.microsoft.com/en-us/windows/win32/inputdev/virtual-key-codes
func GetAsyncKeyState(vKey int) (ok bool, c error) {
	//[in] int vKey
	if v, _, err := pGetAsyncKeyState.Call(uintptr(vKey)); v != 0 && err.(syscall.Errno) == 0 {
		return true, err
	} else {
		return false, err
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
			Buffer, _ = ReadProcessMemory(HANDLE, uintptr(BaseThing), 8)
			Pointer = binary.LittleEndian.Uint64(Buffer)
			Pointer += uint64(Offsets[i])
			fmt.Printf("%x\n", Pointer)
		} else {
			Buffer, _ = ReadProcessMemory(HANDLE, uintptr(Pointer), 8)
			Pointer = binary.LittleEndian.Uint64(Buffer)
			Pointer += uint64(Offsets[i])
			fmt.Printf("%x\n", Pointer)
		}
	}
	return uintptr(Pointer)
}

// https://learn.microsoft.com/en-us/windows/win32/api/tlhelp32/nf-tlhelp32-createtoolhelp32snapshot
func CreateToolhelp32Snapshot() (snapshot uintptr, c error) {
	v, _, err := pCreateToolhelp32Snapshot.Call(
		//[in] DWORD dwFlags,
		uintptr(0x00000002),
		//[in] DWORD th32ProcessID
		uintptr(0),
	)
	return v, err
}

// https://learn.microsoft.com/en-us/windows/win32/api/tlhelp32/nf-tlhelp32-process32firstw
func Process32FirstW(HANDLE uintptr, PE32 syscall.ProcessEntry32) (ok bool, procentry syscall.ProcessEntry32, c error) {
	v, _, err := pProcess32FirstW.Call(
		// [in]      HANDLE           hSnapshot,
		HANDLE,
		// [in, out] LPPROCESSENTRY32 lppe
		uintptr(unsafe.Pointer(&PE32)),
	)
	if v == 1 {
		return true, PE32, err
	}
	return false, PE32, err
}

// https://learn.microsoft.com/en-us/windows/win32/api/tlhelp32/nf-tlhelp32-process32nextw
func Process32NextW(HANDLE uintptr, PE32 syscall.ProcessEntry32) (ok bool, procentry syscall.ProcessEntry32, c error) {
	v, _, err := pProcess32NextW.Call(
		// [in]  HANDLE           hSnapshot,
		HANDLE,
		// [out] LPPROCESSENTRY32 lppe
		uintptr(unsafe.Pointer(&PE32)),
	)
	if v == 1 {
		return true, PE32, err
	}
	return false, PE32, err
}

// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getforegroundwindow
func GetForegroundWindow() (HWND uintptr, c error) {
	v, _, err := pGetForegroundWindow.Call()
	return v, err
}

// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getwindowthreadprocessid
func GetWindowThreadProcessId(HWND uintptr) (pid uint32, c error) {
	var PID uint32
	_, _, err := PGetWindowThreadProcessId.Call(
		// [in]            HWND    hWnd,
		HWND,
		// [out, optional] LPDWORD lpdwProcessId
		uintptr(unsafe.Pointer(&PID)),
	)
	return PID, err
}
