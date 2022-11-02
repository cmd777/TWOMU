# TWOMBC

This War of Mine Better Camera

# What is This?

This app makes it possible to use the W/A/S/D keys to control the camera in This War of Mine

<figure>
    <img src="demo.gif">
    <figcaption>Video Demo</figcaption>
</figure>

# How does this work?

It's pretty simple, when the app detects that the W/A/S/D key(s) are held down, it writes memory to specific addresses.

# How can I use it?

Download the app from the [Releases page](https://github.com/cmd777/TWOMBC/releases/latest), if you want to adjust the settings, open cmd and type

```bash
C:\Users\YourUserName\Downloads>twombc-arch.exe -Step float -CheckFG=bool -ReadMem=bool -FixCam=bool -PrintErr=bool
```

Step -> determines how fast/much the camera should move when pressing W/A/S/D, The default value is 0.7

CheckFG -> will periodically check if This War of Mine is the foreground application<br>This fixes an issue, where even if TWOM is not foreground, key inputs would register, and set X, Y positions from other applications.<br>Recommended value is true

ReadMem -> will periodically (every 10ms) write the X, Y coordinates from the game's memory to a stored one<br>This fixes an issue where pressing tab or using the mouse to change camera position would rubberband the camera back.<br>Recommended value is true

FixCam -> will periodically (every 10ms) check, and set a value to an address that controls the camera mode.<br>This fixes a notorious issue, that when you loaded into a level, or moved the camera by other means, would disable the ability to use W/A/S/D controls<br>Highly recommended to keep this value on true.

PrintErr -> If set to true, PrintErr will print any error that comes up.<br>However, it can be quite spammy.

Not setting anything, and just running the exe as is will set the default values (0.7, true, true, true, false)

**Make sure This War of Mine is running before launching the program**

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

It's open source, if you don't want to use the pre-built exe, you can also compile the application yourself