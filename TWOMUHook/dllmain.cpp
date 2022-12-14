#define WIN32_LEAN_AND_MEAN
#include <windows.h>
#include <Psapi.h>
#include <string>
#include <mutex>

#include "dep/dxsdk/d3d9.h"
#include "dep/dxsdk/d3dx9.h"

#include "dep/detours/detours.h"

#include "dep/imgui/imgui.h"
#include "dep/imgui/imgui_impl_dx9.h"
#include "dep/imgui/imgui_impl_win32.h"

#pragma comment (lib, "dep/dxsdk/d3d9.lib")
#pragma comment (lib, "dep/dxsdk/d3dx9.lib")
#pragma comment (lib, "dep/detours/detours.lib")

#pragma region Typedef, vars

WNDPROC V_WNDPROC = NULL;
HWND V_HWND = NULL;
HMODULE V_HMODULE = NULL;

typedef HRESULT(__stdcall* EndScene)(IDirect3DDevice9* d3ddev9);
EndScene V_EndScene;

typedef HRESULT(__stdcall* Reset)(IDirect3DDevice9* d3ddev9, D3DPRESENT_PARAMETERS* d3dpp);
Reset V_Reset;

bool SHOW = true;
bool INIT = false;

bool WINIT = false;
bool MINIT = false;
bool CINIT = false;

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

std::mutex Mutex;

extern IMGUI_IMPL_API LRESULT ImGui_ImplWin32_WndProcHandler(HWND hWnd, UINT msg, WPARAM wParam, LPARAM lParam);

#pragma endregion

LRESULT __stdcall WndProc(HWND hWnd, UINT msg, WPARAM wParam, LPARAM lParam)
{
    if (SHOW && ImGui_ImplWin32_WndProcHandler(hWnd, msg, wParam, lParam))
    {
        return true;
    }
    return CallWindowProc(V_WNDPROC, hWnd, msg, wParam, lParam);
}

void NOP(HANDLE hProcess, LPVOID lpAddress, SIZE_T nSize)
{
    byte* nopArray = new byte[nSize];

    for (size_t i = 0; i < nSize; i++)
    {
        nopArray[i] = 0x90;
    }

    WriteProcessMemory(hProcess, lpAddress, nopArray, nSize, 0);

    delete[] nopArray;
}

void WASDControls()
{
    for (;;)
    {
        DWORD fgWPID = 0;
        GetWindowThreadProcessId(GetForegroundWindow(), &fgWPID);
        if (fgWPID == PID && UseWASD)
        {
            Mutex.lock();
            if (GetAsyncKeyState(0x57))
            {
                Y += Step;
                WriteProcessMemory(TWOMHandle, (LPVOID)YPos, &Y, sizeof(Y), 0);
            }
            if (GetAsyncKeyState(0x41))
            {
                X -= Step;
                WriteProcessMemory(TWOMHandle, (LPVOID)XPos, &X, sizeof(X), 0);
            }
            if (GetAsyncKeyState(0x53))
            {
                Y -= Step;
                WriteProcessMemory(TWOMHandle, (LPVOID)YPos, &Y, sizeof(Y), 0);
            }
            if (GetAsyncKeyState(0x44))
            {
                X += Step;
                WriteProcessMemory(TWOMHandle, (LPVOID)XPos, &X, sizeof(X), 0);
            }
            Mutex.unlock();
            Sleep(10);
        }
    }
}

void ReadMem()
{
    for (;;)
    {
        if (ReadMemory)
        {
            Mutex.lock();
            ReadProcessMemory(TWOMHandle, (LPVOID)XPos, &X, sizeof(X), 0);
            ReadProcessMemory(TWOMHandle, (LPVOID)YPos, &Y, sizeof(Y), 0);
            Mutex.unlock();
            Sleep(10);
        }
    }
}

void FixCam()
{
    for (;;)
    {
        if (FixCamera)
        {
            ReadProcessMemory(TWOMHandle, (LPVOID)CMode, &CM, sizeof(CM), 0);

            if (CM != 148602)
            {
                CM = 148602;
                WriteProcessMemory(TWOMHandle, (LPVOID)CMode, &CM, sizeof(CM), 0);
            }
            Sleep(10);
        }
    }
}

INT64 GetTWOM(HANDLE H)
{
    HMODULE Modules[256];

    DWORD cbNeeded;

    if (EnumProcessModules(H, Modules, sizeof(Modules), &cbNeeded))
    {
        for (size_t i = 0; i < sizeof(Modules); i++)
        {
            char ModuleFileName[260];

            if (GetModuleFileNameEx(H, Modules[i], ModuleFileName, sizeof(ModuleFileName)))
            {
                std::string ModuleName = ModuleFileName;

                if (ModuleName.find("This War of Mine") != std::string::npos)
                {
                    MODULEINFO MInfo = {};

                    GetModuleInformation(H, Modules[i], &MInfo, sizeof(MInfo));

                    return (INT64)MInfo.lpBaseOfDll;
                }
            }
        }
    }
    return NULL;
}

INT64 Offsets(HANDLE hProcess, INT64 BaseAddr, INT64 Off[], SIZE_T OffSize)
{
    INT64 Buffer;

    ReadProcessMemory(hProcess, (LPVOID)BaseAddr, &Buffer, 8, 0);
    printf("BASE %I64x\n", Buffer);

    for (size_t i = 0; i < OffSize - 1; i++)
    {
        ReadProcessMemory(hProcess, (LPVOID)(Buffer + Off[i]), &Buffer, 8, 0);
        printf("%zu-> %I64x\n", i, Buffer);
    }

    Buffer += Off[OffSize - 1];
    printf("%zu-> %I64x", OffSize - 1, Buffer);

    return Buffer;
}

HRESULT __stdcall HookEndScene(IDirect3DDevice9* d3ddev9)
{
    if (d3ddev9 == NULL)
    {
        return V_EndScene(d3ddev9);
    }

    if (!INIT)
    {
        ImGui::CreateContext();
        ImGui::StyleColorsClassic();

        D3DDEVICE_CREATION_PARAMETERS d3ddevcp = {};
        d3ddev9->GetCreationParameters(&d3ddevcp);
        V_HWND = d3ddevcp.hFocusWindow;

        if (V_HWND != NULL)
        {
            V_WNDPROC = (WNDPROC)SetWindowLongPtr(V_HWND, GWLP_WNDPROC, (LONG_PTR)WndProc);

            GetWindowThreadProcessId(V_HWND, &PID);

            TWOMHandle = OpenProcess(PROCESS_ALL_ACCESS, false, PID);

            BaseAddr = GetTWOM(TWOMHandle);

            Pencil = BaseAddr + 0x24D782;
            Rain = BaseAddr + 0x1B431F;
            Outlines = BaseAddr + 0x2164B0;
            MWndProc = BaseAddr + 0x4C2C31;

            XPos = Offsets(TWOMHandle, (BaseAddr + 0x009064D0), XArray, sizeof(XArray) / sizeof(XArray[0]));
            YPos = Offsets(TWOMHandle, (BaseAddr + 0x009064D0), YArray, sizeof(YArray) / sizeof(YArray[0]));
            CMode = Offsets(TWOMHandle, (BaseAddr + 0x009064D0), CArray, sizeof(CArray) / sizeof(CArray[0]));

            ImGui_ImplWin32_Init(V_HWND);
            ImGui_ImplDX9_Init(d3ddev9);
            INIT = true;
        }
    }

    ImGui_ImplDX9_NewFrame();
    ImGui_ImplWin32_NewFrame();
    ImGui::NewFrame();

    if (GetAsyncKeyState(VK_DELETE) & 1) { SHOW = !SHOW; }

    if (SHOW)
    {
        ImGui::Begin("This War of Mine Utils", &SHOW);

        ImGui::GetStyle().WindowTitleAlign = ImVec2(0.5f, 0.5f);

        ImGui::Checkbox("Use WASD", &UseWASD);
        if (UseWASD && !WINIT)
        {
            CreateThread(NULL, NULL, (LPTHREAD_START_ROUTINE)WASDControls, NULL, NULL, NULL);
            WINIT = true;
        }

        ImGui::Checkbox("Store Game Memory", &ReadMemory);
        if (ReadMemory && !MINIT)
        {
            CreateThread(NULL, NULL, (LPTHREAD_START_ROUTINE)ReadMem, NULL, NULL, NULL);
            MINIT = true;
        }

        ImGui::Checkbox("Fix Camera", &FixCamera);
        if (FixCamera && !CINIT)
        {
            CreateThread(NULL, NULL, (LPTHREAD_START_ROUTINE)FixCam, NULL, NULL, NULL);
            CINIT = true;
        }

        ImGui::Checkbox("Disable Pencil Effect", &DisablePEffect);
        if (DisablePEffect) {
            NOP(TWOMHandle, (PVOID)Pencil, sizeof(V_Pencil));
        }
        else {
            WriteProcessMemory(TWOMHandle, (PVOID)Pencil, V_Pencil, sizeof(V_Pencil), 0);
        }

        ImGui::Checkbox("Disable Rain Effect", &DisableRainEffect);
        if (DisableRainEffect) {
            NOP(TWOMHandle, (PVOID)Rain, sizeof(V_Rain));
        }
        else {
            WriteProcessMemory(TWOMHandle, (PVOID)Rain, V_Rain, sizeof(V_Rain), 0);
        }

        ImGui::Checkbox("Disable Character Outlines", &DisableOutlines);
        if (DisableOutlines) {
            NOP(TWOMHandle, (PVOID)Outlines, sizeof(V_Outlines));
        }
        else {
            WriteProcessMemory(TWOMHandle, (PVOID)Outlines, V_Outlines, sizeof(V_Outlines), 0);
        }

        ImGui::Checkbox("Modify WndProc", &ModifyWndProc);
        if (ModifyWndProc) {
            NOP(TWOMHandle, (PVOID)MWndProc, sizeof(V_MWndProc));
        }
        else {
            WriteProcessMemory(TWOMHandle, (PVOID)MWndProc, V_MWndProc, sizeof(V_MWndProc), 0);
        }

        ImGui::Separator();

        ImGui::CheckboxFlags("[ImGui] Enable Keyboard Controls/Navigation", &ImGui::GetIO().ConfigFlags, ImGuiConfigFlags_NavEnableKeyboard);

        ImGui::End();
    }

    ImGui::EndFrame();
    ImGui::Render();
    ImGui_ImplDX9_RenderDrawData(ImGui::GetDrawData());
    return V_EndScene(d3ddev9);
}

HRESULT __stdcall HookReset(IDirect3DDevice9* d3ddev9, D3DPRESENT_PARAMETERS* d3dpp)
{
    if (d3ddev9 == NULL)
    {
        return V_Reset(d3ddev9, d3dpp);
    }

    ImGui_ImplDX9_InvalidateDeviceObjects();
    auto Reset = V_Reset(d3ddev9, d3dpp);
    ImGui_ImplDX9_CreateDeviceObjects();

    return Reset;
}

DWORD __stdcall MainHook(LPVOID lpParameter)
{
    HWND wHwnd = CreateWindow("BUTTON", "tmpwindow", WS_SYSMENU | WS_MINIMIZEBOX, CW_USEDEFAULT, CW_USEDEFAULT, 100, 100, NULL, NULL, V_HMODULE, NULL);
    if (wHwnd == NULL)
    {
        return false;
    }

    IDirect3D9* d3d9 = Direct3DCreate9(D3D_SDK_VERSION);
    if (d3d9 == NULL)
    {
        DestroyWindow(wHwnd);
        return false;
    }

    D3DPRESENT_PARAMETERS d3dpp = {};

    ZeroMemory(&d3dpp, sizeof(D3DPRESENT_PARAMETERS));

    d3dpp.SwapEffect = D3DSWAPEFFECT_DISCARD;
    d3dpp.hDeviceWindow = wHwnd;
    d3dpp.BackBufferFormat = D3DFMT_UNKNOWN;
    d3dpp.Windowed = true;

    IDirect3DDevice9* d3ddev9;

    if (d3d9->CreateDevice(D3DADAPTER_DEFAULT, D3DDEVTYPE_HAL, wHwnd, D3DCREATE_SOFTWARE_VERTEXPROCESSING, &d3dpp, &d3ddev9) != D3D_OK)
    {
        d3d9->Release();
        DestroyWindow(wHwnd);
        return false;
    }

    if (d3ddev9 == NULL)
    {
        d3d9->Release();
        DestroyWindow(wHwnd);
        return false;
    }

    DWORD64* vTable = (DWORD64*)d3ddev9;
    vTable = (DWORD64*)vTable[0];

    V_EndScene = (EndScene)vTable[42];
    V_Reset = (Reset)vTable[16];

    DetourTransactionBegin();
    DetourUpdateThread(GetCurrentThread());

    DetourAttach(&(LPVOID&)V_EndScene, (PBYTE)HookEndScene);
    DetourAttach(&(LPVOID&)V_Reset, (PBYTE)HookReset);

    DetourTransactionCommit();

    d3ddev9->Release();
    d3d9->Release();
    DestroyWindow(wHwnd);
    return true;
}

BOOL __stdcall DllMain(HMODULE hModule, DWORD ul_reason_for_call, LPVOID lpReserved)
{
    if (ul_reason_for_call == DLL_PROCESS_ATTACH)
    {
        V_HMODULE = hModule;
        DisableThreadLibraryCalls(hModule);
        CreateThread(NULL, NULL, MainHook, NULL, NULL, NULL);
    }
    if (ul_reason_for_call == DLL_PROCESS_DETACH)
    {
        ImGui_ImplDX9_Shutdown();
        ImGui_ImplWin32_Shutdown();
        ImGui::DestroyContext();
        FreeLibraryAndExitThread(hModule, 0);
    }

    return TRUE;
}