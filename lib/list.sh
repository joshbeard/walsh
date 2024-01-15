#!/usr/bin/env bash
# View lists
source "$HOME/.local/share/wallpaper/etc/wallpaper.cfg"
source "$HOME/.local/share/wallpaper/lib/common.sh"

verbose=false

usage() {
  echo "list - View lists"
  echo
  echo "Usage: list [SUBCOMMAND] [FLAGS] [LIST]"
  echo
  echo "Subcommands:"
  echo "  cat     View the contents of a list."
  echo "  view    View the contents of a list in a viewer."
  echo
  echo "Arguments:"
  echo "  <list>  Specify a list name to view."
  echo
  echo "Flags:"
  echo "  -v      Verbose output."
  echo "  -h      Display this help message."
  echo
  echo "Example:"
  echo "  list              # List all lists."
  echo "  list cat nature   # View the 'nature' list files."
  echo "  list view nature  # View the 'nature' list images in a viewer."
  exit 0
}

# Cat the contents of a list.
cat_list() {
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
}

view_list() {
    filename="${lists_dir}/${1}.txt"
    if [ ! -f "$filename" ]; then
        echo "List '${1}' does not exist." >&2
        exit 1
    fi

    cd "$wallpaper_dir" || exit 1
    #feh -d -g 800x600 --scale-down --auto-zoom -f "$filename" &
    view_cmd="feh -d -g 800x600 --scale-down --auto-zoom -f ${filename} &"
    if [ "$verbose" = true ]; then
        echo "Viewing list '${1}' with command '${view_cmd}'"
    fi
    eval "$view_cmd"
}

# No arguments provided, list all lists.
if [ -z "$1" ]; then
    for list in "${lists_dir}"/*.txt; do
        count="$(wc -l "${list}" | awk '{print $1}')"
        list_name="$(basename "${list}" .txt) (${count})"
        echo "${list_name}"
    done

    exit 0
fi

# Check for subcommands.
subcommand="$1"
shift

# Parse options
while getopts ":vh-:" opt; do
    case "$1" in
        -v|--verbose)
            verbose=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            break
            ;;
    esac
done

# Parse subcommands
case "${subcommand}" in
    cat)
        cat_list "$@"
        exit 0
        ;;
    view)
        view_list "$@"
        exit 0
        ;;
    help|-h|--help)
        usage
        ;;
    *)
        ;;
esac


