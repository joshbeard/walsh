#!/usr/bin/env bash
source "$HOME/.local/share/wallpaper/etc/wallpaper.cfg"

export SWWW_TRANSITION_FPS=60
export SWWW_TRANSITION_STEP=2

check_required_command swww hyprctl

# Function to set wallpaper for a specific display
# Arguments:
#   display: The display to set the wallpaper for
#   img: The image to set as wallpaper
set_wallpaper() {
    set_wallpaper_cmd=$(echo "$wayland_set_wallpaper_cmd" | sed "s|{{DISPLAY}}|$display|g" | sed "s|{{IMAGE}}|$img|g")
    log_debug "Running command: $set_wallpaper_cmd"
    eval $set_wallpaper_cmd || return 1
}

# Function to get the list of monitors
get_monitors() {
    echo hyprctl monitors | grep "^Monitor"
}

# Function to get the current wallpaper
# Arguments:
#   display: The display to get the wallpaper for
get_current_wallpaper() {
    display="$1"
    query=$(swww query | head -n "$display" | tail -n 1)
    echo "$query" | awk -F 'image: ' '{print $2}'
}
