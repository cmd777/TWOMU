# TWOMBC

This War of Mine Better Camera

# What is This?

This app makes it possible to use WASD to control the camera in This War of Mine

<figure>
    <img src="demo.gif">
    <figcaption>Video Demo</figcaption>
</figure>

# How does this work?

It's pretty simple, when the app detects that the 'a' key is held down, it writes memory to an address.

# How can I use it?

Download the app from the Releases page, if you want to adjust settings, open cmd and type

```bash
C:\Users\YourUserName\Downloads>twombc-arch.exe -Step float -CheckFG=bool -FixCam=bool
```

Step determines how fast/much the camera should move when pressing W/A/S/D, The default value is 0.2

CheckFG will periodically (every 100ms) check if This War of Mine is the foreground application<br>This fixes an issue, where even if TWOM is not foreground, key inputs would register, and set X, Y positions from other applications.<br>Recommended & Default value is true

FixCam will periodically (every 100ms) check, and set a value to an address that controls the camera mode.<br>This fixes a notorious issue, that when you loaded into a level, or moved the camera by other means, would disable the ability to use W/A/S/D controls<br>Highly recommended to keep this value on true.

Not setting anything, and just running the exe as is will set the default values (0.2, true, true)

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