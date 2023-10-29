#!/usr/bin/env bash
# View lists
source "$HOME/.local/share/wallpaper/etc/wallpaper.cfg"
source "$HOME/.local/share/wallpaper/lib/common.sh"

verbose=false

usage() {
  echo "list - View lists"
  echo
  echo "Usage: list [FLAGS] [LIST]"
  echo "  <list>  Specify a list name to view."
  echo
  echo "Flags:"
  echo "  -v      Verbose output."
  echo "  -h      Display this help message."
  echo
  echo "Example:"
  echo "  list         # List all lists."
  echo "  list nature  # View the 'nature' list."
  exit 0
}

# Check for flags.
while getopts ":vh" opt; do
    case "${opt}" in
        v)
            verbose=true
            shift
            ;;
        h)
            usage
            ;;
        \?)
            echo "Invalid option: -${OPTARG}" >&2
            usage
            ;;
    esac
done

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
if [ ! -f "${lists_dir}/${list_name}.txt" ]; then
    echo "List '${list_name}' does not exist." >&2
    exit 1
fi

if [ "$verbose" = true ]; then
    count="$(wc -l "${lists_dir}/${list_name}.txt" | awk '{print $1}')"
    echo "List '${list_name}' (${count}):"
    echo "----------------------------------------"
fi
cat "${lists_dir}/${list_name}.txt"
