#!/usr/bin/env bash
# Adds the image(s) to a favorites file.
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

images="$@"
first_arg="$1"
last_arg="${@: -1}"
list_name="$last_arg"

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

    blacklist_post_run_cmd="${blacklist_post_run_cmd} -d $display"

    images="$wallpaper"
fi

if [ ! -f "$favorites_file" ]; then
    touch "$favorites_file"
fi

add_to_blacklist() {
    image="$1"
    echo "Adding $image to blacklist"
    echo "$(md5sum ${wallpaper_dir}/${image} | awk '{print $1}')::$image" >> "$blacklist_file"
    rm -f "$image"
}

IFS=$'\n'
for image in $images; do
    if [ ! -f "${wallpaper_dir}/${image}" ]; then
        echo "File not found: ${wallpaper_dir}/${image}"
        continue
    fi

    if [ ! -d "$lists_dir" ]; then
        mkdir -p "$lists_dir"
    fi

    if [ "$list_name" == "blacklist" ]; then
        add_to_blacklist "$image"
        continue
    fi

    if grep -q "$image" "${lists_dir}/${list_name}.txt"; then
        echo "Image ${image} is already in ${list_name} list"
        continue
    fi

    echo "Adding $image to list ${lists_dir}/${list_name}.txt"

    echo "$image" >> "${lists_dir}/${list_name}.txt"
done

if [ "$list_name" == "blacklist" ]; then
    if [ -n "$blacklist_post_run_cmd" ]; then
        log_debug "Running blacklist_post_run_cmd: $blacklist_post_run_cmd"
        eval "$blacklist_post_run_cmd"
    fi
fi
