# PiPs

[Русский](README.md) | [English](README.en.md) | [Українська](README.uk.md)

My first project on GitHub. I made it for myself and decided to share it.

A small Windows utility that lets you watch videos in Picture-in-Picture (PiP) mode and click on the windows underneath the video.

The app runs in the system tray. You can control it with hotkeys or its tray menu. You don't need AutoHotkey or a browser extension to use PiPs.

## Features

- Turn clicks through the PiP window on or off.
- Pause and resume playback.
- Automatically round the PiP window's corners, with a toggle in the menu.
- Restore mouse control with an emergency reset.
- Optionally start the app with Windows.
- Russian, English, and Ukrainian interface languages.
- Restore mouse control and the original corner settings when you exit through the app's menu.

## System and browsers

The app is built for Windows 10/11. macOS and Linux are not supported. The main build is for Windows x64; the source code also passes tests on x86.

Supported browsers:

- **Opera**
- **Microsoft Edge**
- **Google Chrome**

## Getting started

1. Use a ready-made build from Releases, or build `pip-toggle.exe` using the instructions below.
2. Put the file in a permanent folder where you have write access and run it. You don't need Go installed to run the ready-made `.exe`.
3. Look for the PiPs icon in the system tray near the clock. It may be hidden in the tray's overflow menu.
4. Open a video in Opera, Edge, or Chrome and use the browser's Picture-in-Picture feature.
5. Press `Ctrl+Alt+P`: you can now click on the windows underneath the video. Press it again to bring back normal PiP controls.

When updating, close the old version using **Exit** in the tray menu before starting the new one.

## Hotkeys

- `Ctrl+Alt+P` — toggle clicks through the PiP window.
- `Ctrl+Alt+Space` — pause / resume playback.
- `Ctrl+Alt+Shift+P` — emergency reset: restore mouse control for the selected PiP window.

While clicks pass through the window, you can't use its buttons or move it with the mouse. Turn this mode off with the hotkey or tray menu first.

## Tray menu

The menu shows the current mode and lets you toggle clicks through the window, control playback, turn rounded corners on or off, manage startup, run an emergency reset, and exit.

Rounded corners are enabled when the app starts. Startup with Windows is enabled only through the menu and is saved for the current Windows user. If you move the `.exe` to another folder, turn startup off first, then enable it again from the app in its new location.

## Language

Choose your language in the tray menu: **Language → Русский / English / Українська** (the menu is called **«Язык»** in Russian). The menu and notifications switch right away, and your choice is saved between launches.

## Known issues


- **Window not found:** make sure you have opened a separate video PiP window using the browser's own feature, rather than just a tab with a video.
- **A hotkey doesn't work:** check whether another app is using it; you can also use the matching option in the tray menu.
- **You can't click on or move the video:** turn clicks through the window off, or press `Ctrl+Alt+Shift+P`.
- **Pause controls another app:** if the PiP window doesn't handle the command directly, PiPs falls back to the Windows media key. Which app receives it depends on the active media session.

A `pip-toggle.log` file is created next to the `.exe`. You can use it for troubleshooting. Before sharing a log, check it for personal file paths and window titles. Logs are excluded from Git by `.gitignore`.

## Building from source

You'll need Windows and Go 1.26.3 or newer. Run these PowerShell commands from the project folder.

```powershell
go mod download
$env:GOARCH = 'amd64'
go build -mod=readonly -trimpath -ldflags '-H=windowsgui' -o pip-toggle.exe .
```

The `-H=windowsgui` flag builds the app without a console window. To build for x86, replace `amd64` with `386`.

Check the source code:

```powershell
go test -mod=readonly ./...
go vet -mod=readonly ./...
```

The diagnostic tool shows information about visible windows belonging to supported browsers:

```powershell
go build -mod=readonly -trimpath -o diag.exe ./cmd/diag
.\diag.exe
```

## Publishing

The repository includes the source code, tests, `go.mod`, `go.sum`, and documentation.

You can share the ready-made `pip-toggle.exe` as a separate archive in GitHub Releases, together with `LICENSE` and `THIRD_PARTY_NOTICES.md`.

## License

The PiPs source code is available under the [MIT License](LICENSE): you can use, change, and share it. Keep the license text and credit me as the author :)

Licenses for third-party dependencies are listed in [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
