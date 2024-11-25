# Changelog

This project uses [semantic versioning](https://semver.org/) and
[conventional commits](https://www.conventionalcommits.org/en/v1.0.0/).

Throughout the `0.x` series, breaking changes may occur in minor releases.

## 0.6.0 (Unreleased)

There's a couple of changes in CLI arguments in this release.

- feat!: `walsh run` command
  - The interval flag was moved from the `set` command to its own command called
    `run`. This also provides an optional system tray icon and menu.
- feat: System Tray and Menu
  - Adds an optional system tray icon that allows for quick access to common
    commands. It uses [fyne.io/systray](https://github.com/fyne-io/systray)
    and works on Linux/BSD and macOS.
  - Existing CLI commands still behave the same way and can be used in
    conjunction with the system tray.
- feat: Fuzzy menu for rofi, wofi, dmenu, choose, etc. (`walsh menu`)
  - Run `walsh menu` to launch a fuzzy menu in your tool of choice for quick
    wallpaper actions.
  - Bind this to a keybinding for quick access.
- fix: (macOS): Display identification and consistency
- fix: Lower case JSON keys for image objects (e.g. current, history, lists)
  in config.
  - Existing configs will be updated automatically.
- refactor: Move the looped set command to its own package called "scheduler".
- refactor: Check current image each time "view" or "blacklist" is ran
- This previously checked the current state file for the image, which only
  tracks the image as it was last set on a specific display. This change
  improves accuracy, particularly with virtual desktops on macOS.

## 0.5.4 - 2024-08-16

- fix: Hyprland display indexing @joshbeard (#52)
- fix: update cache dir @joshbeard (#51)

## 0.5.3 - 2024-08-14

- fix: don't try to create download dest @joshbeard (#49)

## 0.5.2 - 2024-07-23

- fix: fault tolerance @joshbeard (#48)
- chore: Bump golang.org/x/vuln from 1.1.2 to 1.1.3 @dependabot (#47)
- chore: Bump github.com/adrg/xdg from 0.4.0 to 0.5.0 @dependabot (#46)
- chore: Bump actions/setup-go from 5.0.1 to 5.0.2 @dependabot (#45)

## 0.5.1 - 2024-06-30

- fix: discover hyprland instances @joshbeard (#44)

## 0.5.0 - 2024-06-26

- feat: Custom set and view commands @joshbeard (#43)
- ci: add deepsource config @joshbeard (#42)

## 0.4.7 - 2024-06-25

- fix: don't panic on unknown session @joshbeard (#39)
- ci: add on.push rule for codeql @joshbeard (#41)
- chore: Use go x.y.z version @joshbeard (#40)
- docs: add title image @joshbeard (#38)

## 0.4.6 - 2024-06-24

- feat: Add --no-move arg to download @joshbeard (#37)
- fix: Improve ordering of source determination @joshbeard (#36)
- fix: Improve source and target @joshbeard (#35)
- docs: improve banner @joshbeard (#34)
- fix: build date @joshbeard (#33)

## 0.4.5 - 2024-06-24

- fix: Log output when downloading @joshbeard (#32)
- feat: add colored banner on version info @joshbeard (#31)
- docs: another readme update @joshbeard (#30)

## 0.4.4 - 2024-06-24

- maint: minor cleanups @joshbeard (#29)
- docs: More readme improvements @joshbeard (#28)
- docs: More readme improvements @joshbeard (#27)
- docs: readme improvements @joshbeard (#26)
- docs: remove duplicate license file @joshbeard (#25)

## 0.4.3 - 2024-06-24

- fix: image repetition on displays @joshbeard (#24)

## 0.4.2 - 2024-06-24

- feat: Avoid using same image on multiple displays @joshbeard (#23)
- build: Improve version output @joshbeard (#23)

## 0.4.1 - 2024-06-24

- fix: download dir; lock files @joshbeard (#22)
- fix: install script on darwin/bsd @joshbeard (#21)
- docs: Add download example commands @joshbeard (#20)

## 0.4.0 - 2024-06-24

- feat: download @joshbeard (#19)
- fix: hyprland display index @joshbeard (#18)

## 0.3.0 - 2024-06-24

- feat: sway support @joshbeard (#17)

## 0.2.3 - 2024-06-24

- docs: example with SSH source @joshbeard (#16)
- fix: refresh session/displays on interval @joshbeard (#15)

## 0.2.2 - 2024-06-24

- fix: don't set concurrently on xorg @joshbeard (#14)
- docs: fix license formatting @joshbeard (#13)
- docs: minor readme updates @joshbeard (#12)

## 0.2.1 - 2024-06-24

- fix: Xorg fix and custom logfile @joshbeard (#11)

## 0.2.0 - 2024-06-24

- feat: xorg and macos support @joshbeard (#10)

## 0.1.0 - 2024-06-23

- refactor: Rewrite in Go @joshbeard (#2)

## 0.0.8 - 2024-06-15

- fix: setting image on specific display

## 0.0.7 - 2024-05-05

- fix: blacklist file updates
- feat: test ssh connections
- fix: display identification

## 0.0.6 - 2024-02-27

- fix: get current wallpaper on wayland

## 0.0.5 - 2024-01-15

- docs: codacy badge

## 0.0.4 - 2024-01-15

- fix: remote images
  - refactor: Make `get_remote_image()` its own function/generic
  - fix: path inconsistencies with local vs remote walls
  - feat: Support blacklisting and adding remote images to lists

## 0.0.3 - 2023-11-01

- chore: cleanup zsh completion

## 0.0.2 - 2023-10-19

- feat: verbose list output

## 0.0.1 - 2023-10-18

- feat: initial release (shell)
  - cli tool
  - derived from a single shell script and organized into separate files with
    subcommands
  - supports xorg and hyprland

## Pre-0.0.1

- original source in [joshbeard/dotfiles](https://github.com/joshbeard/dotfiles)
  - 2023-10-11 - [`357a6b1`](https://github.com/joshbeard/dotfiles/commit/357a6b1a695b4e40953ab9db50bed4fcce06f692):
  xorg support
  - 2023-09-16 - [`093924c`](https://github.com/joshbeard/dotfiles/commit/093924c01d83f6201b6ddde8665dc53db56b7ce4):
    break script into separate files with subcommands and config
  - 2023-06-16 - [`56801a0`](https://github.com/joshbeard/dotfiles/commit/56801a0b31f868188bc1121785199e91fb2006c8):
    original single shell script for randomizing files in a directory with `swww`
