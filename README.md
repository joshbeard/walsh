# walsh

Josh's wallpaper tool

This is a simple wallpaper manager for randomizing images on multiple displays
from different sources, saving images to lists, blacklisting images, and more.
It's a personal tool that fits my needs by wrapping other tools to do the real
work.

It supports a variety of desktop environments.

<img align="right" width="256" height="256" src=".doc/image.jpg">

It's evolved from a couple of simple shell scripts, to a single more intuitive
script, and now a Go program. It's really just a wrapper, though.

## Features

* Download wallpapers from Bing and Unsplash using
  [gosimac](https://github.com/1995parham/gosimac)
* Randomly set a wallpaper per display
* Set wallpapers on demand, per-display
* Add wallpapers to a list and set from a list
* Track recent wallpapers and avoid setting them for a while
* Blacklist wallpapers
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


## Usage

```shell
walsh [command] [flags]
```

See `walsh help` for more information.

If you run `walsh` without any arguments, it defaults to the `set` command and
will set a random wallpaper on each display.

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

## Configuration

Standard XDG configuration directories are used for configuration files.

The default configuration file is `${XDG_CONFIG_HOME}/walsh/config.yml`
(e.g. `~/.config/walsh/config.yaml`) and will be created if it does not exist.

The default configuration is as follows:

```yaml
sources: []
current: ${XDG_DATA_HOME}/walsh/current.json
blacklist: ${XDG_CONFIG_HOME}/walsh/blacklist.json
history: ${XDG_DATA_HOME}/walsh/history.json
lists_dir: ${XDG_DATA_HOME}/walsh/lists
tmp_dir: ${XDG_CACHE_HOME}/walsh
history_size: 50
cache_size: 50
interval: 0
```

For example:

```yaml
sources: []
current: /home/user/.local/share/walsh/current.json
blacklist: /home/user/.config/walsh/blacklist.json
history: /home/user/.local/share/walsh/history.json
lists_dir: /home/user/.local/share/walsh/lists
tmp_dir: /home/user/.cache/walsh
history_size: 50
cache_size: 50
interval: 0
```


### Sources

Sources is a list of directories or URIs to source images from. Directories
are absolute paths to directories on the local filesystem. Alternatively, an
SSH URI can be used to source images from a remote directory.

```yaml
sources:
  - /home/user/Pictures/wallpapers
  - ssh://user@host/path/to/wallpapers
```

To use a remote source in a crontab, make sure SSH is configured correct. E.g.

```plain
*/10 * * * * pgrep XORG && DISPLAY=:0 XDG_SESSION_TYPE=x11 SSH_AUTH_SOCK=/path/to/ssh_agent nice -n 19 walls.sh set --once
```

### Lists Directory

Lists directory is the directory where lists of wallpapers are stored. Lists
are plain text files with one image path per line.

### Blacklist File

Blacklist file is a plain text file with one image path per line. Images in
the blacklist will not be set as wallpapers.

### History File

History file is a plain text file with one image path per line. Images in the
history file will not be set as wallpapers until all other images have been
set.

### Current File

Current file is a JSON file that tracks the currently set wallpaper.
