# Define the function
_walls_completion() {
    local state
    local -a subcmds

    subcmds=(
        'help:Show help message'
        'set:Set wallpaper(s) and exit'
        'start:Start the wallpaper randomizer'
        'blacklist:Blacklist the current wallpaper'
        'bl:Blacklist the current wallpaper'
        'download:Download wallpapers'
        'add:Add a wallpaper to a list'
        'list:List wallpapers'
        'view:View the wallpaper set'
    )

    _arguments '1: :->subcmds' && return 0

    case $words[2] in
        help)
            ;;
        set)
            _arguments \
                '--once[Set the wallpaper once and exit]' \
                '--interval[Set the interval between wallpaper changes]:interval (seconds):' \
                '(-l --list)'{-l,--list}'[Set wallpaper from a list of images]:list:' \
                '(-d --display)'{-d,--display}'[Specify a single digit for display]:digit:' \
                '--no-track[Do not track the last wallpaper set]' \
                '--ignore-track[Ignore the last wallpaper set]' \
                '*:dir containing images:_files -/' 
            ;;
        add)
            local list_dir="$HOME/.local/share/wallpaper/var/lists"
            _arguments \
                '2:the display with the wallpaper (digit)' \
                '*:file:_files' \
                '3:list:->lists'
            if [[ $state == lists ]]; then
                local -a list_files
                list_files=($list_dir/*(.:t:r))

                _describe 'lists' list_files
            fi
            ;;
        bl|blacklist)
            _arguments \
                '2:the display with the wallpaper (digit)'
            ;;
        list)
            _arguments \
                '-v[Verbose output]' \
                '-h[Display help message]' \
                '2:list:->lists'
            if [[ $state == lists ]]; then
                local list_dir="$HOME/.local/share/wallpaper/var/lists"
                local -a list_files
                list_files=($list_dir/*(.:t:r))

                _describe 'lists' list_files
            fi
            ;;
        view)
            _arguments '2:display'
            ;;
        *)
            _describe 'subcommands' subcmds
            ;;
    esac
}

compdef _walls_completion walls.sh

