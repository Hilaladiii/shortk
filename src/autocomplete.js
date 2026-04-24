/**
 * Generates a shell completion script for Bash and Zsh.
 * @returns {string} The completion script.
 */
export function generateCompletionScript() {
  return `
# shortk completion script

_shortk_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="\${COMP_WORDS[COMP_CWORD]}"
    prev="\${COMP_WORDS[COMP_CWORD-1]}"
    opts="init add remove list status completion help"

    # Handle sub-commands completion (first argument)
    if [[ \${COMP_CWORD} -eq 1 ]]; then
        COMPREPLY=( $(compgen -W "\${opts}" -- "\${cur}") )
        return 0
    fi

    # Handle command-specific arguments (second argument)
    case "\${prev}" in
        remove|status)
            # Try to fetch existing keys. 
            # We use 'command -v shortk' to ensure the binary is reachable.
            if command -v shortk >/dev/null 2>&1; then
                local keys=$(shortk _list-keys 2>/dev/null)
                if [ -n "\${keys}" ]; then
                    COMPREPLY=( $(compgen -W "\${keys}" -- "\${cur}") )
                fi
            fi
            
            # If no matches found, prevent shell from falling back to file completion
            if [[ \${#COMPREPLY[@]} -eq 0 ]]; then
                # On newer bash, we can use compopt. 
                # On older/zsh, we return a dummy to silence it.
                if type compopt &>/dev/null; then
                    compopt +o default
                else
                    COMPREPLY=('')
                fi
            fi
            return 0
            ;;
        add)
            # Suggest --local if the user starts typing a flag
            if [[ \${cur} == -* ]]; then
                COMPREPLY=( $(compgen -W "--local" -- "\${cur}") )
            fi
            return 0
            ;;
    esac

    # Generic flag suggestion for any other position
    if [[ \${cur} == -* ]]; then
        COMPREPLY=( $(compgen -W "--local --help" -- "\${cur}") )
        return 0
    fi

    return 0
}

# Bash integration
if [ -n "$BASH_VERSION" ]; then
    complete -F _shortk_completion shortk
fi

# Zsh integration (via bashcompinit)
if [ -n "$ZSH_VERSION" ]; then
    autoload -Uz bashcompinit && bashcompinit
    complete -F _shortk_completion shortk
fi
`;
}
