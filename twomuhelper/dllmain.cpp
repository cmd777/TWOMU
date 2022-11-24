#include <stdio.h>
#include <Windows.h>

void MainConsole()
{
    AllocConsole();
    FILE* scon = new FILE();
    freopen_s(&scon, "CONOUT$", "w", stdout);
    printf("Testing, no functionality yet.");
}

BOOL APIENTRY DllMain( HMODULE hModule,
                       DWORD  ul_reason_for_call,
                       LPVOID lpReserved
                     )
{
    switch (ul_reason_for_call)
    {
        case DLL_PROCESS_ATTACH:
        {
            HANDLE hThread = CreateThread(NULL, 0, (LPTHREAD_START_ROUTINE)MainConsole, hModule, 0, 0);
            if (hThread != NULL)
            {
                CloseHandle(hThread);
            }
            break;
        }
        case DLL_THREAD_ATTACH:
        {

        }
        case DLL_THREAD_DETACH:
        {

        }
        case DLL_PROCESS_DETACH:
        {
            break;
        }
    }
    return TRUE;
}

