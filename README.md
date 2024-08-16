# walsh

Walsh is a niche CLI wallpaper manager for randomizing images on multiple
displays from various configured sources. It supports saving images to lists,
blacklisting unwanted images, downloading images, and more, by acting as a
wrapper around other tools to query displays and set wallpapers.

<img align="right" width="256" height="256" src=".doc/image.jpg">

## Features

* Download wallpapers from Bing and Unsplash using [gosimac](https://github.com/1995parham/gosimac)
* Set wallpapers randomly or specifically for each display
* Change wallpapers on demand or at regular intervals
* Manage wallpaper lists and set wallpapers from these lists
* Track recent wallpapers to avoid repetition
* Blacklist unwanted wallpapers
* Source images from a remote server over SSH
* Supports Xorg, Wayland, and macOS

## Getting Started

1. Ensure [dependencies](#dependencies) are installed.
2. [Install](#installation) the latest version of walsh.
3. [Run](#usage) walsh to set a random wallpaper on each display.
    * You can run `walsh download bing` to get a few wallpapers to start.
    * Then just run `walsh set` to set a random wallpaper.
    * `~/Pictures/Wallpapers` is the default source directory.
4. [Configure](#configuration) walsh to customize it, specifically the sources
   you want to use.

## Dependencies

### Wayland

Only [swww](https://github.com/Horus645/swww) is supported for setting the
wallpaper on Wayland. swww works across Wayland compositors.

Hyprland and Sway have been tested and are known to work, using `hyprctl` and
`swaymsg` respectively.

### Xorg

* `xrandr`
* One of the following tools, in order of preference:
  * [nitrogen](https://wiki.archlinux.org/title/Nitrogen)
  * [feh](https://wiki.archlinux.org/title/Feh)
  * [xwallpaper](https://github.com/stoeckmann/xwallpaper)
  * [xsetbg](https://linux.die.net/man/1/xsetbg)

### macOS

No specific dependencies are required for macOS.

### Download from Bing and Unsplash

If the `download` command is used, the [gosimac](https://github.com/1995parham/gosimac)
tool should be installed and in the `PATH`.

## Installation

The latest release can be found on the [releases](https://github.com/joshbeard/walsh/releases)
page and can be downloaded and installed manually.

To download and install the latest version of walsh using `curl` and piping it
to the shell, run the following command:

```sh
curl -sfL https://raw.githubusercontent.com/joshbeard/walsh/master/install.sh | sh -
```

The script will:
- Detect your OS and architecture.
- Download the latest release of walsh from GitHub.
- Verify the checksum of the downloaded package.
- Extract the binary and move it to the specified directory (default is `$HOME/bin`).

Make sure the installation directory is in your `PATH` so you can easily run
`walsh` from anywhere.

### Custom Installation Directory

If you want to specify a custom installation directory, you can set the
`INSTALL_DIR` environment variable or pass the `-d` (or `--dir`) argument. For
example:

```sh
# Using INSTALL_DIR environment variable
INSTALL_DIR=/usr/local/bin curl -sfL https://raw.githubusercontent.com/joshbeard/walsh/master/install.sh | sh -

# Using -d (or --dir) argument
curl -sfL https://raw.githubusercontent.com/joshbeard/walsh/master/install.sh | sh -s -- -d /usr/local/bin
```

## Usage

```shell
walsh [command] [flags]
```

See `walsh help` for more information.

If you run `walsh` without any arguments, it defaults to the `set` command and
will set a random wallpaper on each display.

Ensure a configuration file exists at the default location and has at least one
source configured. See [Configuration](#configuration) for more information.

### Set Wallpaper


```shell
# Set a random wallpaper on each display using the configured sources:
walsh set

# Set a random wallpaper on a specific display:
walsh set -d 1

# _`s` is an alias for `set`._
# _You can also omit the `-d` flag and specify a number without it._
walsh s 1

# Set a random wallpaper from a specific list:
walsh set -l my-list

# Set a random wallpaper from a directory:
walsh set ~/Pictures/wallpapers

# Set a specific wallpaper on each display:
walsh set ~/Pictures/wallpapers/wallpaper.jpg

# Set a specific wallpaper on a specific display:
walsh set -d 1 ~/Pictures/wallpapers/wallpaper.jpg

# Set a random wallpaper from an SSH source:
walsh set ssh://user@host/path/to/wallpapers
```

### View Wallpaper


```shell
# View the current wallpaper on each display:
walsh view

# View the current wallpaper on a specific display:
walsh view -d 1
```

### Blacklist


```shell
# Blacklist the current wallpaper on a specific display:
walsh bl -d 1

walsh bl 1
```

### Download

Download wallpapers from Bing and Unsplash using
[gosimac](https://github.com/1995parham/gosimac).



```shell
walsh download bing
walsh download unsplash

# Use short aliases:
walsh dl b
walsh dl u

# Use a query with Unsplash:
walsh dl u -- --query "nature"
```

## Configuration

Standard XDG configuration directories are used for configuration files.

The default configuration file is `${XDG_CONFIG_HOME}/walsh/config.yml`
(e.g. `~/.config/walsh/config.yaml`) and will be created if it does not exist.

The default configuration is as follows:

```yaml
# A list of directories or URIs to source images from.
sources:
  - ${XDG_HOME}/Pictures/Wallpapers

# The file to track the currently set wallpaper.
current: ${XDG_DATA_HOME}/walsh/current.json

# The file to track blacklisted wallpapers.
blacklist: ${XDG_CONFIG_HOME}/walsh/blacklist.json

# The file to track wallpaper history.
history: ${XDG_DATA_HOME}/walsh/history.json

# The directory where lists of wallpapers are stored.
lists_dir: ${XDG_DATA_HOME}/walsh/lists

# The directory to store temporary files, such as images downloaded from remote
# sources.
cache_dir: ${XDG_CACHE_HOME}/walsh

# The number of images to keep in the history file.
history_size: 50

# The number of images to keep in the cache.
cache_size: 50

# The interval in seconds to set a new wallpaper. Set to 0 to disable.
interval: 0

# A destination path or URI to download images to. This is used by the
# 'download' # command.
# Specify a path with optional environment variables or an SSH URI.
download_dest: ${XDG_HOME}/Pictures/Wallpapers

# set_command is the command used to set the specified wallpaper.
# Use {{path}} to specify the path to the wallpaper and {{display}} to specify
# the display number.
# e.g. swww img '{{path}}' --outputs '{{display}}'
# By default, this uses 'swww' on Wayland, 'nitrogen' on Xorg, and falls back
# to looking for 'feh', 'xwallpaper', 'xsetbg'. Native system calls via
# AppleScript are used on macOS.
set_command: ""

# view_command is the command used to view the specified wallpaper.
# Use {{path}} to specify the path to the wallpaper.
# e.g. feh --bg-fill '{{path}}'
# By default, this uses 'xdg-open' on Linux/BSD and falls back to looking for
# feh, eog, eom. Preview is used by default on macOS.
view_command: ""
```

* On Linux and BSD, `${XDG_CONFIG_HOME}` defaults to `~/.config`,
  `${XDG_DATA_HOME}` defaults to `~/.local/share`, and `${XDG_CACHE_HOME}`
  defaults to `~/.cache`.
* On macOS, `${XDG_CONFIG_HOME}` defaults to `~/Library/Application Support`,
  `${XDG_DATA_HOME}` defaults to `~/Library/Application Support`, and
  `${XDG_CACHE_HOME}` defaults to `~/Library/Caches`.

### Sources

Sources is a list of directories or URIs to source images from. Directories
are absolute or relative paths to directories on the local filesystem and can
use environment variables (e.g. `${HOME}/Pictures/Wallpapers`).
Alternatively, an SSH URI can be used to source images from a remote directory.

```yaml
sources:
  - /home/user/Pictures/wallpapers
  - ssh://user@host:/path/to/wallpapers
  - ssh://myhost:/path/to/wallpapers
```

### Desktop Environment Integration

Run `walsh` however you like to set wallpapers. On Linux/BSD desktops, it's
preferred to use the startup configuration of your desktop environment to run
`walsh` at login. You can also use something like cron, systemd, or a launchd
agent to run `walsh` at regular intervals. Or just run it on demand.

If using an SSH source, you will need to ensure your SSH agent is running and
the `SSH_AUTH_SOCK` environment variable is set. This is necessary for the
remote source to work.

#### Hyprland

In `~/.config/hypr/hyprland.conf`:

```plain
exec-once = $HOME/bin/walsh set --interval 600
```

When an SSH source is used, you may need to set the `SSH_AUTH_SOCK` environment
variable:

```plain
exec-once = SSH_AUTH_SOCK=/run/user/1000/gcr/ssh $HOME/bin/walsh set --interval 600
```

#### i3

In `~/.config/i3/config`:

```plain
exec --no-startup-id $HOME/bin/walsh set --interval 600
```

When an SSH source is used, you may need to set the `SSH_AUTH_SOCK` environment
variable:

```plain
exec --no-startup-id export SSH_AUTH_SOCK=/run/user/1000/gcr/ssh && $HOME/bin/walsh set --interval 600
```

#### macOS

Create a `launchd` plist file at
`~/Library/LaunchAgents/com.github.joshbeard.walsh.plist` with the following
contents:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.yourusername.walsh</string>
    <key>ProgramArguments</key>
    <array>
        <string>/Users/yourusername/bin/walsh</string>
        <string>set</string>
        <string>--interval</string>
        <string>600</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/walsh.stdout</string>
    <key>StandardErrorPath</key>
    <string>/tmp/walsh.stderr</string>
</dict>
</plist>
```

Then load the plist:

```shell
launchctl load ~/Library/LaunchAgents/com.github.joshbeard.walsh.plist
```

To unload the plist:

```shell
launchctl unload ~/Library/LaunchAgents/com.github.joshbeard.walsh.plist
```

## License

walsh is released under the [0BSD license](LICENSE.md)

* swww is released under the [GPL-3.0 license](https://github.com/LGFae/swww/blob/main/LICENSE)
* gosimac is released under the [GPL-2.0 license](https://github.com/1995parham/gosimac/blob/main/LICENSE)
* nitrogen is released under the [GPL-2.0 license](https://github.com/l3ib/nitrogen/?tab=GPL-2.0-1-ov-file#readme)
