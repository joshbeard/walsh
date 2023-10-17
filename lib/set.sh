#!/usr/bin/env bash
# Set a random wallpaper at regular intervals from a directory of
# images. This sets a different wallpaper for each monitor.

# This script will randomly go through the files of a directory, setting it
# up as the wallpaper at regular intervals

# Defaults
interval=600

source "$HOME/.local/share/wallpaper/etc/wallpaper.cfg"
source "$HOME/.local/share/wallpaper/lib/common.sh"

# Initialize flags and variables
once_flag=false
display_flag=false
display= # Default display value
list=""
no_track=false
ignore_track=false

usage() {
    echo "set - Set a random wallpaper at regular intervals from a directory of images."
    echo
    echo "Usage: set [--once] [--interval <seconds>] [-d, --display <digit>] <dir containing images>"
    echo "  --once        Set the wallpaper once and exit."
    echo "  --interval <seconds>  Set the interval between wallpaper changes."
    echo "  -l|--list <list>      Set wallpaper from a list of images."
    echo "  -d, --display <digit> Specify a single digit for display."
    echo "  -h, --help            Display this help message."
    echo "  --no-track            Do not track the last wallpaper set."
    echo "  --ignore-track        Ignore the last wallpaper set."
    echo "  <dir containing images>  Specify a directory containing images."
    exit 0
}

# Process command line arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --once)
            once_flag=true
            shift
            ;;
        --interval)
            interval_flag=true
            if [[ -n "$2" && "$2" =~ ^[0-9]+$ ]]; then
                interval="$2"
                shift 2
            else
                echo "Error: --interval requires a positive integer argument."
                exit 1
            fi
            ;;
        -d|--display)
            display_flag=true
            if [[ -n "$2" && "$2" =~ ^[0-9]$ ]]; then
                display="$2"
                shift 2
            else
                echo "Error: -d, --display requires a single digit argument."
                exit 1
            fi
            ;;
        -l|--list)
            if [[ -n "$2" ]]; then
                list="$2"
                shift 2
            else
                echo "Error: -l, --list requires an argument."
                exit 1
            fi
            ;;
        --no-track)
            no_track=true
            shift
            ;;
        --ignore-track)
            ignore_track=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            if [[ -d "$1" ]]; then
                wallpaper_dir="$1"
            else
                echo "Error: Invalid argument: $1"
                usage
                exit 1
            fi
            shift
            ;;
    esac
done

if [[ -z "$wallpaper_dir" ]]; then
    echo "Usage: $0 [--once] [--interval <seconds>] [-d, --display <digit>] <dir containing images>"
    exit 1
fi

get_display_name() {
    segment=$(echo "$1" | grep -o "Monitor \([^ ]*\)" | awk '{print $2}')
    echo "$segment"
}

get_images() {
    if [ -n "$list" ]; then
        cat "${lists_dir}/${list}.txt"
    else
        find "$wallpaper_dir" -maxdepth 1 -type f -iname '*.jpg' -o -iname '*.jpeg' -o -iname '*.png'
    fi
}

get_random_image() {
    if [ -n "$1" ]; then
        echo "$1" | shuf -n 1
    fi
}

update_current_info() {
    local display="$1"
    local new_img="$2"
    local img_basename=$(basename "$new_img")

    if [ ! -e "$currently_set_file" ]; then
        touch "$currently_set_file"
    fi

    if grep "^$display:" $currently_set_file; then
        sed -i "/^${display}:/s|.*|${display}:${img_basename}|" "$currently_set_file"
    else
        echo "$display:$img_basename" >> "$currently_set_file"
    fi
}

add_to_track_file() {
    img="$1"
    # Add the image to the list of last wallpapers. Keep 1000 lines in the file.
    if [ -n "$img" ]; then
        img_basename=$(basename "$img")
        echo "$img_basename" >> "$track_file"
        tail -n $last_set_count "$track_file" > "$track_file.tmp"
        mv "$track_file.tmp" "$track_file"
    fi
}

while true; do
    monitors=$(get_monitors)
    log_debug "found monitors: $monitors"

    images=$(get_images)
    log_debug "found $(echo "$images" | wc -l) images"

    IFS=$'\n'
    displaynum=0

    for monitor in $monitors; do
        current_display=$(get_display_name "$monitor")
        img=""
        img_basename=""

        if [ "$display_flag" = true ] && [ "$displaynum" != "$display" ]; then
            log_debug "Skipping display $displaynum ($display)"
            displaynum=$((displaynum+1))
            continue
        fi

        x=0
        # Try to set the wallpaper 10 times before giving up.
        while [ $x -lt 10 ]; do
            x=$((x+1))

            img=$(get_random_image "$images")
            img_basename=$(basename "$img")

            if [ -n "$list" ]; then
                img="${wallpaper_dir}/${img}"
            fi


            if [ "$ignore_track" = false ]; then
                # Check if the image is in the list of last wallpapers and skip it if it is.
                grep -q "$img_basename" "$track_file" && echo "Skipping $img_basename because it was recently set" && continue
            fi

            set_wallpaper "$displaynum" "$img" && break
            sleep 1
        done

        if [ "$no_track" == false ]; then
            add_to_track_file "$img"
        fi

        update_current_info "$displaynum" "$img"
        log_info "Display $displaynum set to: $img"
        displaynum=$((displaynum+1))
    done

    if [ "$once_flag" = true ]; then
        echo "Exiting because --once flag was set"
        break
    fi

    echo "Sleeping for $interval seconds"
    sleep "$interval"
done


