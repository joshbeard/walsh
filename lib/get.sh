#!/usr/bin/env bash
# Wrapper for the abandonware 'gosimac' program.
# Even with the right arguments, gosimac downloads images to its own directory.
# This script verifies that the downloaded images aren't in the blacklist, then
# moves them to the wallpaper directory specified in the config file if they
# aren't blacklisted.
source "$HOME/.local/share/wallpaper/etc/wallpaper.cfg"
source "$HOME/.local/share/wallpaper/lib/common.sh"

dl_bing=true
dl_unsplash=true
move=true

usage() {
    echo "Usage: $0 [bing|unsplash] [ARGS]"
    exit 0
}

# Parse arguments.
while [ $# -gt 0 ]; do
    case "$1" in
        -h|--help)
            usage
            ;;
        -n|--no-move)
            move=false
            ;;
        bing)
            dl_unsplash=false
            break
            ;;
        unsplash)
            dl_bing=false
            break
            ;;
        *)
            usage
            ;;
    esac
    shift
done

check_required_command md5sum gosimac

if [ "$dl_bing" = true ]; then
    echo "Downloading from Bing..."
    gosimac bing $@
fi

if [ "$dl_unsplash" = true ]; then
    echo "Downloading from Unsplash..."
    gosimac unsplash $@
fi

# Compare each file in src_path to the entries in blacklist.txt, formatted as
# md5sum::path. If the md5sum of the file matches the md5sum in blacklist.txt,
# then the file is removed from src_path.
for file in $src_path/*; do
    md5=$(md5sum "$file" | awk '{print $1}')
    if grep -q "$md5" "$blacklist_file"; then
        echo "File $file is blacklisted. Removing..."
        rm "$file"
    fi

    # Check if it already exits in wallpaper_dir.
    if [ -f "${wallpaper_dir}/$(basename "$file")" ]; then
        echo "File $file already exists in ${wallpaper_dir}. Removing..."
        rm "$file"
    fi
done

# Check if the directory has any files in it.
if [ -z "$(ls -A "$src_path")" ]; then
    echo "No new files to move."
    exit 0
fi

# Move all files in src_path to dst_path.
if [ "$move" = true ]; then
    mv -v "${src_path}"/* "${wallpaper_dir}"
fi
