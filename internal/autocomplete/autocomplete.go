package autocomplete

func GenerateCompletionScript() string {
	return `
# shortk completion script

_shortk_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="init add remove list status completion help"

    # If we are at the first argument, suggest sub-commands
    if [[ ${COMP_CWORD} -eq 1 ]]; then
        COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
        return 0
    fi

    # Command-specific completion for the second argument
    case "${prev}" in
        remove|status)
            # Try to get keys from shortk.
            local keys=$(shortk _list-keys 2>/dev/null)
            if [ -n "${keys}" ]; then
                COMPREPLY=( $(compgen -W "${keys}" -- "${cur}") )
            else
                # Provide a dummy completion to prevent file fallback
                if type compopt &>/dev/null; then
                    compopt +o default
                else
                    COMPREPLY=('')
                fi
            fi
            return 0
            ;;
        add)
            if [[ ${cur} == -* ]] ; then
                COMPREPLY=( $(compgen -W "--local" -- "${cur}") )
            fi
            return 0
            ;;
        *)
            if [[ ${cur} == -* ]] ; then
                COMPREPLY=( $(compgen -W "--local --help" -- "${cur}") )
                return 0
            fi
            ;;
    esac

    return 0
}

# Bash completion
if [ -n "$BASH_VERSION" ]; then
    complete -F _shortk_completion shortk
fi

# Zsh completion (compatibility mode)
if [ -n "$ZSH_VERSION" ]; then
    autoload -Uz bashcompinit && bashcompinit
    complete -F _shortk_completion shortk
fi
`
}
