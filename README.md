# Josh's Wallpaper Scripts

Scripts for managing desktop wallpapers.

They're currently somewhat specific to my Linux desktop and depend on:

* [gosimac](https://github.com/1995parham/gosimac) for downloading wallpapers
  from Bing and Usplash (via cron with [`bin/wall-get.sh`](bin/wall-get.sh)).
* on xorg: xrandr and nitrogen
* on wayland: hyprland and [swww](https://github.com/Horus645/swww)

These are just hacked together over time to provide certain functionality.
I don't even see my wallpapers most of the time, since they're typically
covered with other windows.

The result is random wallpapers across all connected displays sourced from
thousands of random images at a regular interval that I can add to lists on
Xorg or Wayland.

## Install

* Clone to `$HOME/.local/share/wallpaper` (or somewhere)
* Modify the [config file](etc/wallpaper.cfg)
* Export the `bin` directory in `$PATH`

## Features

* Download wallpapers from Bing and Unsplash
* Randomly set a wallpaper per display
* Set wallpapers on demand
* Add wallpapers to a list and set from a list
* Track recent wallpapers and avoid setting them for a while
* Blacklist wallpapers
* Supports Xorg and Wayland

## Usage

### Help

See help for the root command and subcommands with the `--help` argument:

```shell
walls.sh --help
walls.sh add --help
walls.sh blacklist --help
walls.sh download --help
walls.sh list --help
walls.sh set --help
walls.sh start --help
walls.sh view --help
```

### Setting Wallpaper

Randomize wallpapers across all detected displays and exit:

```shell
walls.sh set
```
Set a random wallpaper on all displays and exit:

```shell
walls.sh set --once
```

Set a wallpaper on display 2 and exit:

```shell
walls.sh set -d 2 --once
```

Set wallpapers at an interval:

```shell
walls.sh set --interval 600
```

### Using Lists

Save the current wallpaper on display 1 to a list called "nature":

```shell
walls.sh add 1 nature
```

Set a wallpaper from a list:

```shell
walls.sh set -d 2 --once --list nature
```

View lists:

```shell
walls.sh list         # view a list of lists
walls.sh list nature  # view the files in the 'nature' list
```

### Blacklisting

Since this downloads a bunch of random images from the Internet, it may be
necessary to 'blacklist' an image to prevent it from being used as a wallpaper
in the future. It won't prevent its download, but it'll be deleted immediately
and if it is downloaded again.

```shell
walls.sh blacklist 0  # blacklist the wallpaper on display 0
```

### Downloading

```shell
walls.sh download
```

### Viewing

```shell
walls.sh view 0
```
