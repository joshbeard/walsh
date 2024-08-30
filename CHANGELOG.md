# Changelog

## 0.6.0 (Unreleased)

- Feature: Add System Tray/Menu Bar
  - Adds an optional system tray icon that allows for quick access to common
    commands. It uses [fyne.io/systray](https://github.com/fyne-io/systray)
    and works across multiple platforms and desktop environments.
  - Existing CLI commands still behave the same way and can be used in
    conjunction with the system tray.
- Refactors:
  - Move the looped set command to its own package called "scheduler".
  - Check current image each time "view" or "blacklist" is ran
    - This previously checked the current state file for the image, which only
      tracks the image as it was last set on a specific display. This change
      improves accuracy, particularly with virtual desktops on macOS.
- Fix (macOS): Display identification and consistency
- Change (config): Lower case JSON keys for image objects (e.g. current,
  history, lists).
  - Existing configs will be updated automatically.


## 0.5.4 - 2024-08-16

- Fix Hyprland display indexing @joshbeard (#52)
- update cache dir @joshbeard (#51)

## 0.5.3 - 2024-08-14

- fix: don't try to create download dest @joshbeard (#49)

## 0.5.2 - 2024-07-23

- Bump golang.org/x/vuln from 1.1.2 to 1.1.3 @dependabot (#47)
- Bump github.com/adrg/xdg from 0.4.0 to 0.5.0 @dependabot (#46)
- Bump actions/setup-go from 5.0.1 to 5.0.2 @dependabot (#45)
- fix: fault tolerance @joshbeard (#48)

## 0.5.1 - 2024-06-30

- fix: discover hyprland instances @joshbeard (#44)

## 0.5.0 - 2024-06-26

- Custom set and view commands @joshbeard (#43)
- ci: add deepsource config @joshbeard (#42)

## 0.4.7 - 2024-06-25

- ci: add on.push rule for codeql @joshbeard (#41)
- Use go x.y.z version @joshbeard (#40)
- fix: don't panic on unknown session @joshbeard (#39)
- [docs] add title image @joshbeard (#38)

## 0.4.6 - 2024-06-24

- Add --no-move arg to download @joshbeard (#37)
- Improve ordering of source determination @joshbeard (#36)
- Improve source and target @joshbeard (#35)
- [docs] improve banner @joshbeard (#34)
- Fix date @joshbeard (#33)

## 0.4.5 - 2024-06-24

- Log output when downloading @joshbeard (#32)
- add colored banner on version info @joshbeard (#31)
- docs: another readme update @joshbeard (#30)

## 0.4.4 - 2024-06-24

- maint: minor cleanups @joshbeard (#29)
- [docs] More readme improvements @joshbeard (#28)
- [docs] More readme improvements @joshbeard (#27)
- [docs] readme improvements @joshbeard (#26)
- [docs] remove duplicate license file @joshbeard (#25)

## 0.4.3 - 2024-06-24

- fix: image repetition on displays @joshbeard (#24)

## 0.4.2 - 2024-06-24

- Improve version output @joshbeard (#23)
- Avoid using same image on multiple displays @joshbeard (#23)

## 0.4.1 - 2024-06-24

- fix: download dir; lock files @joshbeard (#22)
- fix: install script on darwin/bsd @joshbeard (#21)
- Add download example commands @joshbeard (#20)

## 0.4.0 - 2024-06-24

- feat: download @joshbeard (#19)
- fix hyprland display index @joshbeard (#18)

## 0.3.0 - 2024-06-24

- feat: sway support @joshbeard (#17)

## 0.2.3 - 2024-06-24

- [docs] example with SSH source @joshbeard (#16)
- fix: refresh session/displays on interval @joshbeard (#15)

## 0.2.2 - 2024-06-24

- fix: don't set concurrently on xorg @joshbeard (#14)
- [docs] fix license formatting @joshbeard (#13)
- [docs] minor readme updates @joshbeard (#12)

## 0.2.1 - 2024-06-24

- Xorg fix and custom logfile @joshbeard (#11)

## 0.2.0 - 2024-06-24

- xorg and macos support @joshbeard (#10)

## 0.1.0 - 2024-06-23

- Rewrite in Go @joshbeard (#2)
