# TWOMU

This War of Mine Utils

# What is This?

This app makes it possible to use the W/A/S/D keys to control the camera in This War of Mine, <br>as well as adding several features, like hiding character outlines, or the pencil effect from the game.

<figure>
    <img src="demo.gif">
    <figcaption>(old) Video Demo</figcaption>
</figure>

# How does this work?

It's pretty simple, when the app detects that the W/A/S/D key(s) are held down, <br>or when you enabled a function from the terminal, it writes memory to specific addresses.

# How can I use it?

Download the app from the [Releases page](https://github.com/cmd777/TWOMBC/releases/latest), then run it. simple as that.

Step -> determines how fast/much the camera should move when pressing W/A/S/D, The default value is 0.7

ReadMem -> will periodically (every 10ms) write the X, Y coordinates from the game's memory to a stored one<br>This fixes an issue where pressing tab or using the mouse to change camera position would rubberband the camera back.

FixCam -> will periodically (every 10ms) check, and set a value to an address that controls the camera mode.<br>This fixes a notorious issue, that when you loaded into a level, or moved the camera by other means, would disable the ability to use W/A/S/D controls.

DisablePencil -> will disable the in-game pencil effect.<br>Note: this doesn't have an effect on frame rate.

DisableRain -> will disable the in-game rain effect.<br>Note: this doesn't have an effect on frame rate.

DisableOutlines -> will disable the flashing character outlines.<br>Note: this doesn't have an effect on frame rate.

ModifyWndProc -> will change (NOP) TWOM's WndProc WM_SIZE.<br>Whenever TWOM is minimized, and then reopened, there is about a 2s black screen before the game can show anything (1s because of kernel32's 1000ms sleep, and another one rendering everything), ModifyWndProc will NOP the 'if' condition to WM_SIZE, and make the 2s process near instantaneous<br>This comes at a downside, as attempting to resize the game from anything lower than 100% resolution back to 100% makes everything low resolution (NOTE: It's possible to change resolution back to 100%, it needs to be done from the settings menu.) other than that, there are no other downsides found.

Randomize Settings, Randomize Locations -> will randomly set values in the `My Story` tab<br>
Randomize Settings requires the `My Story` tab to be open <br>
Randomize Locations requires the `My City Map` tab to be open.

# Nothing is happening | Weird characters in the terminal
If you haven't, when TWOM is started, make sure to run option 1 in TWOMU. This will find the game's PID and Handle.

If you have any option enabled, but is not working, run [DebugView](https://learn.microsoft.com/en-us/sysinternals/downloads/debugview), and check the logs.
<br>If you believe this is an issue with TWOMU, submit a bug report.

If you see any unusual characters like &larr;, that is because your terminal doesn't support ANSI escape sequences.
If you are using any windows version after 1511 (build number 10586), then you can easily fix this.
<br>Open regedit, and go to `HKEY_CURRENT_USER\Console`, then create a DWORD (32-bit value) with the name `VirtualTerminalLevel`, and set the value to 1

TBD: SetConsoleTextAttribute, that should work on most windows versions.

# Known Issues
Terminals that don't support ANSI escape sequences have &larr; & no color.<br>
PReadMem breaks or bugs out Stories' intros & animations.<br>
Attempting to set random values in the My Stories tab after leaving a previously saved game will break it.<br>

# My antivirus says it's a virus, is it?

<figure>
    <blockquote>
        <p>
            Commercial virus scanning programs are often confused by the structure of Go binaries, <br> which they don't see as often as those compiled from other languages.
        </p>
    </blockquote>
    <figcaption>Excerpt from <a href="https://go.dev/doc/faq#virus">https://go.dev/doc/faq#virus</a></figcaption>
</figure>

Aside from the antivirus being confused by the app's structure, it may also detect that the program imports user32.dll, kernel32.dll, psapi.dll, and that it calls function like GetAsyncKeyState, Read/Write ProcessMemory, it will mark the app as a virus.

There is an extremely high chance that the antivirus will detect TWOMUHook.dll as well.

**Why?**

It does the same things the Go version does, except, it hooks onto DirectX's EndScene using Detours, and that's a big no-no from antiviruses.

It's open source, you can build everything for yourself, and change things that don't suit your liking

# Can I run it on Linux?

Yes, you can! It requires some tinkering though.

1. Download the latest [TWOMU](https://github.com/cmd777/TWOMBC/releases/latest) Version
2. Install [Wine](https://wiki.winehq.org/Download)
3. Change Steam Compatability to Proton
    - Right click on This War of Mine
    - Go to Properties
    - Click on Compatability
    - Check `Force the use of a specific Steam Play compatability tool`
    - Select the latest proton version (right under Proton Experimental)
4. Paste the TWOMUHook directory to `/<STEAM_PATH>/SteamLibrary/steamapps/compatdata/282070/pfx` (If you wish to use the ImGui version, if not, skip this step.)
5. Open your terminal, and paste the following command
    - `STEAM_COMPAT_CLIENT_INSTALL_PATH=... STEAM_COMPAT_DATA_PATH=... WINEPREFIX=$PWD {PROTONPATH} run "/<TWOMU_INSTALL_PATH>/TWOMU-x64.exe"`

STEAM_COMPAT_CLIENT_INSTALL_PATH is where ".steam/steam" is. (if `apt install steam` was used, it should be `"$HOME/.steam/steam"`)

STEAM_COMPAT_DATA_PATH is where steam installed the game (for example, `"/<STEAM_PATH>/SteamLibrary/steamapps/compatdata/282070"`)

PROTONPATH is where proton is installed (should be in steam's root directory, `"/home/<USERNAME>/.steam/root/steamapps/common/Proton 7.0/proton"`)

## Limitations

Unforunately, it's not perfect yet, and some functions don't work well.

| Function                  | Compatability         |
| :---:                     | :---:                 |
| UseWASD                   | :heavy_check_mark:    |
| ReadMemory                | :heavy_check_mark:    |
| FixCamera                 | :heavy_check_mark:    |
| Disable Pencil Effect     | :interrobang:         |
| Disable Rain Effect       | :interrobang:         |
| Disable Outlines          | :interrobang:         |
| Modify WndProc            | :heavy_check_mark:    |

:heavy_check_mark: - The function works without issues

:interrobang: - The function works, but is prone to bugs.
