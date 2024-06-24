# walsh

This is a simple wallpaper manager for randomizing images on multiple displays
from different sources, saving images to lists, blacklisting images, and more.
It's a personal tool that fits my needs by wrapping other tools to do the real
work.

It supports a variety of desktop environments.

<img align="right" width="256" height="256" src=".doc/image.jpg">

It's evolved from a couple of simple shell scripts, to a single more intuitive
script, and now a Go program. It's really just a wrapper, though.

## Features

These are the main reasons for creating this tool:

* Download wallpapers from Bing and Unsplash using
  [gosimac](https://github.com/1995parham/gosimac)
* Randomly or specifically set a wallpaper per display
* Set wallpapers on demand or at an interval
* Add wallpapers to lists and set from lists
* Track recent wallpapers and avoid setting them for a while
* Blacklist wallpapers (useful when downloading from Bing or Unsplash)
* Source images from remote server over SSH
* Supports Xorg, Wayland, macOS

## Dependencies

* [gosimac](https://github.com/1995parham/gosimac) for downloading wallpapers
  from Bing and Usplash (e.g. via cron). Only needed if `download` is used.
* on xorg: `xrandr` and [nitrogen](https://wiki.archlinux.org/title/Nitrogen)
* on wayland: hyprland and [swww](https://github.com/Horus645/swww) (hyprland's
  `hyprctl` command is used, but this will be changed to something more generic
  soon).

The result is random wallpapers across all connected displays, sourced from
thousands of random images at a regular interval that I can add to lists
whether running on Xorg or Wayland.

## Installation

The latest release can be found on the [releases](https://github.com/joshbeard/walsh/releases)
page and can be downloaded and installed manually.

To download and install the latest version of walsh using `curl` and piping it
to the shell, run the following command:

```sh
curl -sfL https://raw.githubusercontent.com/joshbeard/walsh/master/install.sh | sh -
```

If you want to specify a custom installation directory, you can set the
`INSTALL_DIR` environment variable or pass the `-d` (or `--dir`) argument. For
example:

```sh
# Using INSTALL_DIR environment variable
INSTALL_DIR=/usr/local/bin curl -sfL https://raw.githubusercontent.com/joshbeard/walsh/master/install.sh | sh -

# Using -d (or --dir) argument
curl -sfL https://raw.githubusercontent.com/joshbeard/walsh/master/install.sh | sh -s -- -d /usr/local/bin
```

The script will:
- Detect your OS and architecture.
- Download the latest release of walsh from GitHub.
- Verify the checksum of the downloaded package.
- Extract the binary and move it to the specified directory (default is `$HOME/bin`).

Make sure the installation directory is in your `PATH` so you can easily run
`walsh` from anywhere.

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

Set a random wallpaper on each display using the configured sources:

```shell
walsh set
```

_`s` is an alias for `set`._

Set a random wallpaper on a specific display:

```shell
walsh set -d 1
```

_You can also omit the `-d` flag and specify a number without it._

```shell
walsh s 1
```

Set a random wallpaper from a specific list:

```shell
walsh set -l my-list
```

Set a random wallpaper from a directory:

```shell
walsh set ~/Pictures/wallpapers
```

Set a specific wallpaper on each display:

```shell
walsh set ~/Pictures/wallpapers/wallpaper.jpg
```

Set a specific wallpaper on a specific display:

```shell
walsh set -d 1 ~/Pictures/wallpapers/wallpaper.jpg
```

Set a random wallpaper from an SSH source:

```shell
walsh set ssh://user@host/path/to/wallpapers
```

### View Wallpaper

View the current wallpaper on each display:

```shell
walsh view
```

View the current wallpaper on a specific display:

```shell
walsh view -d 1
```

### Blacklist

Blacklist the current wallpaper on a specific display:

```shell
walsh bl -d 1
```


```shell
walsh bl 1
```

### Desktop Environments

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
cache_dir: ${XDG_CACHE_HOME}/walsh/cache

# The number of images to keep in the history file.
history_size: 50

# The number of images to keep in the cache.
cache_size: 50

# The interval in seconds to set a new wallpaper. Set to 0 to disable.
interval: 0

# A destination path to download images to. This is used by the 'download'
# command.
download_dest: ${XDG_HOME}/Pictures/Wallpapers
```

* On Linux and BSD, `${XDG_CONFIG_HOME}` defaults to `~/.config`,
  `${XDG_DATA_HOME}` defaults to `~/.local/share`, and `${XDG_CACHE_HOME}`
  defaults to `~/.cache`.
* On macOS, `${XDG_CONFIG_HOME}` defaults to `~/Library/Application Support`,
  `${XDG_DATA_HOME}` defaults to `~/Library/Application Support`, and
  `${XDG_CACHE_HOME}` defaults to `~/Library/Caches`.

### Sources

Sources is a list of directories or URIs to source images from. Directories
are absolute paths to directories on the local filesystem. Alternatively, an
SSH URI can be used to source images from a remote directory.

```yaml
sources:
  - /home/user/Pictures/wallpapers
  - ssh://user@host/path/to/wallpapers
```
