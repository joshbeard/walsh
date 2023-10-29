#!/usr/bin/env bash
# View lists
source "$HOME/.local/share/wallpaper/etc/wallpaper.cfg"
source "$HOME/.local/share/wallpaper/lib/common.sh"

usage() {
  echo "list - View lists"
  echo
  echo "Usage: list [LIST]"
  echo "  <list>  Specify a list name to view."
  echo
  echo "Example:"
  echo "  list         # List all lists."
  echo "  list nature  # View the 'nature' list."
  exit 0
}

if [ "$1" == "-h" ] || [ "$1" == "--help" ]; then
    usage
fi

# No arguments provided, list all lists.
if [ -z "$1" ]; then
    for list in "${lists_dir}"/*.txt; do
        count="$(wc -l "${list}" | awk '{print $1}')"
        list_name="$(basename "${list}" .txt) (${count})"
        echo "${list_name}"
    done

    exit 0
fi

# If a list name is provided, view that list.
list_name="$1"
echo "Viewing ${lists_dir}/${list_name}.txt"
echo "============================================================"
cat "${lists_dir}/${list_name}.txt"
