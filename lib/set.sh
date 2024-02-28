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
        if [ -n "$remote" ]; then
            remote_url=$(echo "$remote" | sed -e 's|^[^:]*://||')
            remote_host=$(echo "$remote_url" | cut -d: -f1)
            remote_path=$(echo "$remote_url" | cut -d: -f2-)
            # Get the images from the remote directory
            log_info "Getting images from $remote_host:$remote_path"
            if ! ssh "$remote_host" "find \"$remote_path\" -maxdepth 1 -type f -iname '*.jpg' -o -iname '*.jpeg' -o -iname '*.png'"; then
                log_error "Error getting images from $remote_host:$remote_path"
                exit 1
            fi
        else
          find "$wallpaper_dir" -maxdepth 1 -type f -iname '*.jpg' -o -iname '*.jpeg' -o -iname '*.png'
        fi
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

    if grep "^$display:" "$currently_set_file"; then
        log_debug "Updating $display in $currently_set_file to ${img_basename}"
        sed -i "/^${display}:/s|.*|${display}:${img_basename}|" "$currently_set_file"
    else
        log_debug "Adding $display to $currently_set_file with ${img_basename}"
        echo "$display:$img_basename" >> "$currently_set_file"
    fi
}

add_to_track_file() {
    img="$1"
    # Add the image to the list of last wallpapers. Keep 1000 lines in the file.
    if [ -n "$img" ]; then
        img_basename=$(basename "$img")
        echo "$img_basename" >> "$track_file"
        tail -n "$last_set_count" "$track_file" > "$track_file.tmp"
        mv "$track_file.tmp" "$track_file"
    fi
}

get_remote_wallpaper() {
    # Remove the protocol:// from the remote URL
    remote_url=$(echo "$remote" | sed -e 's|^[^:]*://||')
    # Split the URL into host and path on the colon
    remote_host=$(echo "$remote_url" | cut -d: -f1)
    remote_path=$(echo "$remote_url" | cut -d: -f2-)

    img_name=$(basename "$img")
    log_info "Getting image ${img_name} from $remote_host:$img"
    # Get the images from the remote directory
    # log_info "Getting images from $remote_host:$remote_path"
    if ! scp "$remote_host:$img" "$var_dir/remote/$img_name"; then
        log_error "Error getting image from $remote_host:$remote_path/$img"
        exit 1
    fi
    img="$var_dir/remote/$img_name"
}

while true; do
    monitors=$(get_monitors)
    echo "monitors: $monitors"
    log_debug "found monitors: $monitors"

    images=$(get_images)
    log_debug "found $(echo "$images" | wc -l) images"

    IFS=$'\n'
    displayid=0

    for m in $monitors; do
        img=""
        img_basename=""

        if [ "$XDG_SESSION_TYPE" != "x11" ]; then
            displayid=$(get_display_name "$m")
            if [ -z "$displayid" ]; then
                log_error "No display name found for $m"
                exit 1
            fi
            # displayid="$displayid"
        fi

        if [ "$display_flag" = true ] && [ "$displayid" != "$display" ]; then
            log_debug "Skipping display $displayid ($display)"
            displayid=$((displaynum+1))
            continue
        fi

        x=0
        # Try to set the wallpaper 10 times before giving up.
        while [ $x -lt 10 ]; do
            x=$((x+1))

            img=$(get_random_image "$images")
            img_basename=$(basename "$img")

            if [ -n "$list" ]; then
                log_debug "Setting image from list: $img"
                ignore_track=true
                no_track=true
            fi


            if [ "$ignore_track" = false ]; then
                # Check if the image is in the list of last wallpapers and skip it if it is.
                if [ -f "$track_file" ] && grep -q "$img_basename" "$track_file"; then
                    log_debug "Skipping $img_basename because it was recently set"
                    continue
                fi
            fi

            if [ -n "$remote" ]; then
                get_remote_wallpaper "$img"
                img_path="$var_dir/remote/$img_basename"
            else
                img_path="${wallpaper_dir}/${img_basename}"
            fi

            echo "Setting display $displayid to: $img_path"
            set_wallpaper "$displayid" "$img_path" && break
            sleep 1
        done

        if [ "$no_track" == false ]; then
            add_to_track_file "$img_basename"
        fi

        update_current_info "$displayid" "$img_basename"
        log_info "Display $displayid set to: $img_basename"

        if [ "$XDG_SESSION_TYPE" == "x11" ]; then
            displaynum=$((displaynum+1))
        fi
    done

    # Remote stale images from $var_dir/remote - images that aren't currently set.
    if [ -n "$remote" ]; then
        log_debug "Removing stale images from $var_dir/remote"
        for img in $var_dir/remote/*; do
            if ! grep -q $(basename "$img") "$currently_set_file"; then
                log_debug "Removing stale image $img"
                rm "$img"
            fi
        done
    fi

    if [ "$once_flag" = true ]; then
        log_debug "Exiting because --once flag was set"
        break
    fi

    echo "Sleeping for $interval seconds"
    sleep "$interval"
done


