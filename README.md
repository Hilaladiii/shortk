# shortk

A blazing-fast, zero-dependency command line shortcut manager written in Go.

`shortk` allows you to create global and project-local shortcuts (aliases) for your frequently used terminal commands. It uses a "Dynamic Binaries" architecture, which means your shortcuts work instantly across Bash, Zsh, and Fish without needing to `source` your configuration or restart your terminal.

## Features

- **Blazing Fast**: Written in Go, zero dependencies, <10ms execution.
- **Dynamic Binaries**: Creates real executable wrappers in your PATH instead of relying on shell memory aliases.
- **Project-Local Shortcuts**: Define shortcuts that only apply when you are in a specific project directory (saves to `.shortk`).
- **Auto-load `.env`**: Automatically loads variables from `.env` files and injects them into your commands at runtime.
- **Tab Autocompletion**: Supports native tab completion for commands and your custom shortcuts.
- **Zero Configuration Reloads**: Shortcuts are available instantly in all open terminal tabs after adding them.

## Installation

You can install `shortk` using the provided install script:

```bash
# Clone the repository
git clone https://github.com/username/shortk.git
cd shortk

# Run the install script
./install.sh
```

After installation, reload your shell profile to apply the PATH changes and autocompletion:
```bash
source ~/.zshrc # or source ~/.bashrc
```

## Uninstallation

To completely remove `shortk`, its binary, and global configurations:

```bash
./uninstall.sh
```

*(Note: Local `.shortk` files in your projects are kept intact).*

## Usage & Examples

### 1. Global Shortcuts
Add a shortcut that works everywhere on your system.

```bash
# Add a global shortcut
shortk add dcu "docker compose up -d"

# Use it immediately in any terminal
dcu
```

### 2. Project-Local Shortcuts
Add a shortcut that only works in the current directory and its subdirectories. This creates a `.shortk` file in your directory and automatically adds it to `.gitignore`.

```bash
# Add a local shortcut
shortk add migrate "go run main.go migrate" --local

# Use it
migrate
```

### 3. Environment Variables Injection
`shortk` automatically detects `.env` files and evaluates variables at runtime, not when you define the alias.

```bash
# Suppose you have a .env file with DB_USER=admin and DB_PASS=secret
# Define the shortcut using the variable names directly (escape $ for shell if needed):
shortk add db "psql -U \$DB_USER -W \$DB_PASS" --local

# When you run 'db', shortk will source the .env file and execute the command
db
```

### 4. Tab Autocompletion
`shortk` integrates with Bash and Zsh autocomplete.

```bash
# Type shortk and press TAB
shortk [TAB]
# Suggestions: init add remove list status completion help

# Type shortk remove and press TAB
shortk remove [TAB]
# Suggestions: dcu db migrate
```

### 5. Check Status
Find out where a shortcut is defined (Global or Local) and if a local shortcut is overriding a global one.

```bash
shortk status dcu
# Output:
# Status for 'dcu':
#   GLOBAL: docker compose up -d
#   LOCAL:  (none found in directory tree)

shortk status migrate
# Output:
# Status for 'migrate':
#   LOCAL:  go run main.go migrate (from /path/to/project/.shortk)
```

### 6. List Shortcuts
View all your configured shortcuts.

```bash
# List all global shortcuts
shortk list

# List local shortcuts active in the current directory tree
shortk list --local
```

### 7. Remove Shortcuts
Remove a configured shortcut.

```bash
# Remove global shortcut
shortk remove dcu

# Remove local shortcut
shortk remove migrate --local
```

## License
ISC
