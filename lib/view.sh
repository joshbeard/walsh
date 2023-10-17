#!/usr/bin/env bash
source "$HOME/.local/share/wallpaper/etc/wallpaper.cfg"
source "$HOME/.local/share/wallpaper/lib/common.sh"

usage() {
    echo "view - View the wallpaper set."
    echo
    echo "Usage: view DISPLAY"
    echo "  <digit>  Specify a single digit for display."
    echo
    echo "Example:"
    echo "  view 1 # View the wallpaper set on display 1."
    exit 0
}

display="$1"
echo grep "^$display:" "$currently_set_file" | awk -F ':' '{print $2}'
current=$(grep "^$display:" "$currently_set_file" | awk -F ':' '{print $2}')
if [ -z "$current" ]; then
    echo "No wallpaper set on display $display"
    exit 1
fi

# Replace {{IMAGE}} with the image path in the viewer_cmd
viewer_cmd=$(echo "$viewer_cmd" | sed "s|{{IMAGE}}|${wallpaper_dir}/${current}|g")
echo "Current wallpaper: $current"
$viewer_cmd "${wallpaper_dir}/${current}"
