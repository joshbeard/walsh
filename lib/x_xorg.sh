#!/usr/bin/env bash
source "$HOME/.local/share/wallpaper/etc/wallpaper.cfg"

required_commands=("nitrogen" "xrandr")

# Function to set wallpaper for a specific display
# Arguments:
#   display: The display to set the wallpaper for
#   img: The image to set as wallpaper
set_wallpaper() {
    display="$1"
    img="$2"
    if [ -z "$display" ]; then
        log_error "No display specified"
        exit
    fi

    # If remote is set, retrieve the image from the remote host and store it
    # in ${var_dir}/remote.
    if [ -n "$remote" ]; then
        # Remove the protocol:// from the remote URL
        remote_url=$(echo "$remote" | sed -e 's|^[^:]*://||')
        # Split the URL into host and path on the colon
        remote_host=$(echo "$remote_url" | cut -d: -f1)
        remote_path=$(echo "$remote_url" | cut -d: -f2-)

        img_name=$(basename "$img")
        log_info "Getting image ${img_name} from $remote_host:$img"
        # Get the images from the remote directory
        # log_info "Getting images from $remote_host:$remote_path"
        scp "$remote_host:$img" "$var_dir/remote/$img_name"
        img="$var_dir/remote/$img_name"
    fi


    # Replace {{DISPLAY}} with the display number and {{IMAGE}} with the image.
    set_wallpaper_cmd=$(echo "$xorg_set_wallpaper_cmd" | sed "s|{{DISPLAY}}|$display|g" | sed "s|{{IMAGE}}|$img|g")
    log_info "Running command: $set_wallpaper_cmd"
    eval $set_wallpaper_cmd || return 1
}

# Function to get the list of monitors
get_monitors() {
    mons=$(xrandr --listactivemonitors | grep "^ " | awk '{print $1}')
    echo "$mons"
}

# Function to get the current wallpaper
# Arguments:
#   display: The display to get the wallpaper for
get_current_wallpaper() {
    display="$1"
    log_info "Getting current wallpaper for display $display from $currently_set_file"
    if [ -f "$currently_set_file" ]; then
        for line in $(cat "$currently_set_file"); do
            log_debug "Checking line: $line"
            if [[ "$line" =~ ^$display: ]]; then
                image=$(echo "$line" | awk -F ':' '{print $2}')
                log_debug "Found image: $image"
                echo "$image"
                return
            fi
        done
    else
        log_error "No wallpaper set on display $display"
        log_error "$currently_set_file does not exist"
        exit
    fi
}
