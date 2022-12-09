#pragma once

#define WIN32_LEAN_AND_MEAN
#include <windows.h>
#include <Psapi.h>
#include <string>
#include <mutex>
#include <dwmapi.h>

#include "dep/dxsdk/d3d9.h"
#include "dep/dxsdk/d3dx9.h"

#include "dep/detours/detours.h"

#include "dep/imgui/imgui.h"
#include "dep/imgui/imgui_impl_dx9.h"
#include "dep/imgui/imgui_impl_win32.h"

#pragma comment (lib, "dep/dxsdk/d3d9.lib")
#pragma comment (lib, "dep/dxsdk/d3dx9.lib")
#pragma comment (lib, "dep/detours/detours.lib")

#define TWOMU_VERSION "v3.0.1-EXPERIMENTAL"

WNDPROC V_WNDPROC = NULL;
HWND V_HWND = NULL;
HMODULE V_HMODULE = NULL;

typedef HRESULT(__stdcall* EndScene)(IDirect3DDevice9* d3ddev9);
EndScene V_EndScene;

typedef HRESULT(__stdcall* Reset)(IDirect3DDevice9* d3ddev9, D3DPRESENT_PARAMETERS* d3dpp);
Reset V_Reset;

bool SHOW = true;
bool INIT = false;

bool UseWASD = false;
bool ReadMemory = false;
bool FixCamera = false;
bool DisablePEffect = false;
bool DisableRainEffect = false;
bool DisableOutlines = false;
bool ModifyWndProc = false;

DWORD PID;
HANDLE TWOMHandle;
INT64 BaseAddr;

byte V_Pencil[] = { 0xF3, 0x44, 0x0F, 0x10, 0x05, 0x35, 0x10, 0x4A, 0x00 };
byte V_Rain[] = { 0x0F, 0xB7, 0x05, 0xB8, 0xD7, 0x72, 0x00 };
byte V_Outlines[] = { 0xF3, 0x0F, 0x10, 0x0D, 0xD8, 0x2F, 0x60, 0x00 };
byte V_MWndProc[] = { 0x83, 0xE8, 0x02, 0x74, 0x26 };

INT64 Pencil;
INT64 Rain;
INT64 Outlines;
INT64 MWndProc;

INT64 XArray[] = { 0x70 };
INT64 YArray[] = { 0x78 };
INT64 CArray[] = { 0xA6 };
INT64 XPos;
INT64 YPos;
INT64 CMode;

float Step = 0.2f;
float X = 0.0f;
float Y = 0.0f;

int CM;
int IdealCameraMode = 148602;
int DWMDarkThemeValue = 1;

std::mutex Mutex;

extern IMGUI_IMPL_API LRESULT ImGui_ImplWin32_WndProcHandler(HWND hWnd, UINT msg, WPARAM wParam, LPARAM lParam);