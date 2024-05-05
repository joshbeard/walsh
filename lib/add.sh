#!/usr/bin/env bash
# Adds the image(s) to a list file (favorites).
# Images can be specified as arguments or by specifying a display to query.
source "$HOME/.local/share/wallpaper/etc/wallpaper.cfg"
source "$HOME/.local/share/wallpaper/lib/common.sh"

usage() {
  echo "add - Adds wallpaper images to list."
  echo
  echo "Usage: add [DISPLAY|<file1> <file2>...] LIST"
  echo "  <digit>               Specify a single digit for display."
  echo "  <list>                Specify a list name to add to."
  echo "  <file1> <file2> ...   Specify one or more file paths."
  echo
  echo "Example:"
  echo "  add 1 favs # Adds the current wallpaper on display 1 to the 'favs' list."
  exit 0
}

if [ "$#" -eq 0 ] || [ "$1" == "-h" ] || [ "$1" == "--help" ]; then
    usage
fi

images="$*"
last_arg="${*: -1}"
list_name="$last_arg"
display=""

# If a display is specified, get the wallpaper that's currently set on that
# display.
if [[ "$1" =~ ^[0-9]+$ ]]; then
    display="$1"
    wallpaper=$(get_current_wallpaper "$display")

    log_debug "Wallpaper on display $display: $wallpaper"

    if [ -z "$wallpaper" ]; then
        echo "No wallpaper set on display $display"
        exit
    fi

    images="$wallpaper"
fi

add_to_blacklist() {
    image="$1"
    echo "Adding $image to blacklist"

    if [ -n "$remote" ]; then
        wallpaper_dir="${var_dir}/remote"
    fi

    echo "$(md5sum "${wallpaper_dir}/${image}" | awk '{print $1}')::$image" >> "$blacklist_file"

    if [ -z "$remote" ]; then
        echo "Removing $image from wallpaper directory"
        rm -f "$image"
    fi
}

IFS=$'\n'
for image in $images; do
    if [ -z "$remote" ]; then
        if [ ! -f "${wallpaper_dir}/${image}" ]; then
            echo "File not found: ${wallpaper_dir}/${image}"
            continue
        fi
    fi

    if [ ! -d "$lists_dir" ]; then
        mkdir -p "$lists_dir"
    fi

    if [ "$list_name" == "blacklist" ]; then
        add_to_blacklist "$image"
        continue
    fi

    if [ -f "${lists_dir}/${list_name}.txt" ]; then
        if grep -q "$image" "${lists_dir}/${list_name}.txt"; then
            echo "Image ${image} is already in ${list_name} list"
            continue
        fi
    fi

    echo "Adding $image to list ${lists_dir}/${list_name}.txt"

    echo "$image" >> "${lists_dir}/${list_name}.txt"
done

if [ "$list_name" == "blacklist" ]; then
    # Replace {{DISPLAY}} with the display number and {{IMAGE}} with the image.
    post_run_cmd=$(echo "$blacklist_post_run_cmd" | sed "s|{{DISPLAY}}|$display|g" | sed "s|{{IMAGE}}|$image|g")
    if [ -n "$blacklist_post_run_cmd" ]; then
        log_debug "Running blacklist_post_run_cmd: $post_run_cmd"
        eval "$post_run_cmd"
    fi
fi
