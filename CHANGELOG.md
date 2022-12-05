## [v3.0.0](https://github.com/cmd777/TWOMU/releases/tag/v3.0.0) (2022-12-04)

### Added
- MSBuild to build TWOMUHook
- Solution files, they can be found in the `TWOMUHook` directory (compiling only works with the x64 version)
- Regex to Go, this will find both cls and clear without case sensitivity
- New option to Inject TWOMUHook (q)
- Disable Pencil Effect, Disable Rain Effect, Disable Character Outlines, Modify WndProc to the ImGui version
- NOP, Offsets Function to the ImGui version
- WASD Controls to ImGui
- Keyboard navigation option to ImGui

### Changed
- (Workflows) Runner to Windows
- Files are now zipped
- All inputs in Go are now lowercase, that means that Q will be registered as q

<hr>

## [v3.0.0-pre.4](https://github.com/cmd777/TWOMU/releases/tag/v3.0.0-pre.4) (2022-12-02)
## Added
- NOP, Offsets Function to the ImGui version
- WASD Controls to ImGui
- Keyboard navigation option to ImGui
## Removed
- Nop Arrays from the ImGui version

<hr>

## [v3.0.0-pre.3](https://github.com/cmd777/TWOMU/releases/tag/v3.0.0-pre.3) (2022-11-29)
## Added
- Regex to Go, this will find both cls and clear without case sensitivity
- New option to Inject TWOMUHook (q)
- Disable Pencil Effect, Disable Rain Effect, Disable Character Outlines, Modify WndProc to the ImGui version
## Changed
- All inputs in Go are now lowercase, that means that Q will be registered as q

<hr>

## [v3.0.0-pre.2](https://github.com/cmd777/TWOMU/releases/tag/v3.0.0-pre.2) (2022-11-28)
## Added
- MSBuild to build TWOMUHook
- The built dll file to the zip file (the option to inject it is the same as the option used to find This War of Mine and calculate offsets - This will be changed.)
- Partially implemented ImGui (only the demo window is shown)
- Solution files, they can be found in the `TWOMUHook` directory (compiling only works with the x64 version)

<hr>

## [v3.0.0-pre.1](https://github.com/cmd777/TWOMU/releases/tag/v3.0.0-pre.1) (2022-11-25)

### Added
- (Workflows) G++ to build DllMain
- Custom "hasher" function to write CHANGELOG to Releases

### Changed
- (Workflows) Runner to Windows
- Files are now zipped

<hr>

## [v2.0.1](https://github.com/cmd777/TWOMU/releases/tag/v2.0.1) (2022-11-25)

### Added
- A patch option to allow multiple running instances of This War of Mine


### Removed
- Redundant forever loop in PrintPatches


<hr>

## [v2.0.1-pre.2](https://github.com/cmd777/TWOMU/releases/tag/v2.0.1-pre.2) (2022-11-23)
### Removed
- Redundant forever loop in PrintPatches

## [v2.0.1-pre.1](https://github.com/cmd777/TWOMU/releases/tag/v2.0.1-pre.1) (2022-11-23)
### Added
- A patch option to allow multiple running instances of This War of Mine

<hr>

## [v2.0.0](https://github.com/cmd777/TWOMU/releases/tag/v2.0.0) (2022-11-17)

### Added
- An option to randomize the My Stories & My City Map
- DisableOutlines, which disable the constantly flashing character outlines
- Terminal now has color, this requires a windows version after 1511 (build number 10586) and VirtualTerminalLevel to be set
### Changed
- Name from TWOMBC to TWOMU
- Errors will now be printed to Debug Output, and not the console
- Values can now be changed back from within the console
- Foreground application checking was merged into the WASD controls

### Removed
- Flags, and added a semi-interactive "interface"

<hr>

## [v1.0.4](https://github.com/cmd777/TWOMU/releases/tag/v1.0.4) (2022-11-05)

### Added
- ModifyWndProc, which if set to true, will change (NOP) TWOM's WndProc WM_SIZE

<hr>

## [v1.0.3](https://github.com/cmd777/TWOMU/releases/tag/v1.0.3) (2022-11-04)

### Added
- DisablePencil, which if set to true, will disable the in-game pencil effect (Note: this doesn't have an effect on frame rate)
- DisableRain, which if set to true, will disable the in-game rain effect (Note: this doesn't have an effect on frame rate)

<hr>

## [v1.0.2](https://github.com/cmd777/TWOMU/releases/tag/v1.0.2) (2022-11-02)

### Added
- PrintErr, which if set to true, will print any errors that show up (default value is false, as it can get quite spammy)
- Functions for CheckFG, ReadMem, FixCam
### Changed
- Step's default value from 0.3 -> 0.7

### Fixed
- A lot of typos, and updated the README
- Potential deadlock to time.Sleep with time.Timer

<hr>

## [v1.0.1](https://github.com/cmd777/TWOMU/releases/tag/v1.0.1) (2022-11-01)

### Added
- PReadMem, which will periodically (every 10ms) write the X, Y coordinates from the game's memory to a stored one
- Mutexes, they are locked at the beginning of a goroutine, and unlocked at the end (This fixes a data race)

### Changed
- Step's default value from 0.2 -> 0.3
- Most 100ms sleeps to 10ms

### Fixed
- CanWriteMem (bool) was deleted and replaced with CwMem (chan bool) this fixes a data race

<hr>

## [v1.0.0](https://github.com/cmd777/TWOMU/releases/tag/v1.0.0) (2022-10-31]

Initial Release
