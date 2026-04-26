package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"shortk/internal/config"
	"strings"
)

var BinDir = filepath.Join(config.ConfigDir, "bin")

func EnsureBinDir() error {
	if _, err := os.Stat(BinDir); os.IsNotExist(err) {
		return os.MkdirAll(BinDir, 0755)
	}
	return nil
}

func CreateWrapper(short, long string) error {
	if err := EnsureBinDir(); err != nil {
		return err
	}
	
	// 1. Create bash wrapper (always, for WSL/Git Bash/Linux compatibility)
	scriptPath := filepath.Join(BinDir, short)
	escapedLong := strings.ReplaceAll(long, "'", "'\\''")

	content := fmt.Sprintf(`#!/usr/bin/env bash

SHORT="%s"
GLOBAL_CMD='%s'

load_env() {
  if [ -f "$1" ]; then
    set -a
    source "$1"
    set +a
  fi
}

# Traverse up to find .shortk
dir="$PWD"
while [ "$dir" != "/" ] && [ "$dir" != "." ]; do
  if [ -f "$dir/.shortk" ]; then
    # Use awk to find the exact match key=value
    LOCAL_CMD=$(awk -F'=' -v k="$SHORT" '$1==k {sub(/^[^=]+=/,""); print; exit}' "$dir/.shortk")
    if [ -n "$LOCAL_CMD" ]; then
      load_env "$dir/.env"
      eval "$LOCAL_CMD \"$@\""
      exit $?
    fi
  fi
  dir=$(dirname "$dir")
done

if [ -n "$GLOBAL_CMD" ] && [ "$GLOBAL_CMD" != "undefined" ]; then
  load_env "$PWD/.env"
  eval "$GLOBAL_CMD \"$@\""
else
  echo "shortk: '$SHORT' is not defined globally and no local override found."
  exit 1
fi
`, short, escapedLong)

	if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
		return err
	}

	// 2. Create Windows wrappers
	if runtime.GOOS == "windows" {
		// .bat wrapper
		batPath := scriptPath + ".bat"
		batContent := fmt.Sprintf("@echo off\nshortk _exec \"%%~n0\" %%*\n")
		if err := os.WriteFile(batPath, []byte(batContent), 0755); err != nil {
			return err
		}

		// .ps1 wrapper
		ps1Path := scriptPath + ".ps1"
		ps1Content := fmt.Sprintf("$short = $MyInvocation.MyCommand.Name.Replace(\".ps1\", \"\")\n& shortk _exec $short $args\nexit $LASTEXITCODE\n")
		if err := os.WriteFile(ps1Path, []byte(ps1Content), 0755); err != nil {
			return err
		}
	}

	return nil
}

func RemoveWrapper(short string) error {
	scriptPath := filepath.Join(BinDir, short)
	
	// Remove bash wrapper
	if _, err := os.Stat(scriptPath); err == nil {
		os.Remove(scriptPath)
	}

	// Remove Windows wrappers
	if runtime.GOOS == "windows" {
		os.Remove(scriptPath + ".bat")
		os.Remove(scriptPath + ".ps1")
	}

	return nil
}


func SyncWrappers() error {
	if err := EnsureBinDir(); err != nil {
		return err
	}
	files, err := os.ReadDir(BinDir)
	if err == nil {
		for _, file := range files {
			os.Remove(filepath.Join(BinDir, file.Name()))
		}
	}

	aliases, err := config.LoadConfig()
	if err != nil {
		return err
	}
	for short, long := range aliases {
		if err := CreateWrapper(short, long); err != nil {
			return err
		}
	}
	return nil
}

func InitShellProfile() error {
	home, _ := os.UserHomeDir()
	profiles := []string{
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".bashrc"),
	}

	if runtime.GOOS == "windows" {
		// PowerShell profiles
		profiles = append(profiles, filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"))
		profiles = append(profiles, filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1"))
	}

	startMarker := "# <<< shortk initialize <<<"
	endMarker := "# >>> shortk initialize >>>"

	for _, profile := range profiles {
		if _, err := os.Stat(profile); err == nil {
			data, err := os.ReadFile(profile)
			if err != nil {
				continue
			}
			content := string(data)

			var integrationCode string
			if strings.HasSuffix(profile, ".ps1") {
				integrationCode = fmt.Sprintf("\n\n%s\n$env:PATH = \"%s;\" + $env:PATH\nshortk completion | Out-String | Invoke-Expression\n%s\n", startMarker, BinDir, endMarker)
			} else {
				integrationCode = fmt.Sprintf("\n\n%s\nexport PATH=\"%s:$PATH\"\nsource <(shortk completion)\n%s\n", startMarker, BinDir, endMarker)
			}
			
			if strings.Contains(content, startMarker) {
				// Replace
				lines := strings.Split(content, "\n")
				var newLines []string
				inside := false
				for _, line := range lines {
					if strings.Contains(line, startMarker) {
						inside = true
						newLines = append(newLines, integrationCode)
						continue
					}
					if strings.Contains(line, endMarker) {
						inside = false
						continue
					}
					if !inside {
						newLines = append(newLines, line)
					}
				}
				content = strings.Join(newLines, "\n")
			} else {
				content += integrationCode
			}
			
			err = os.WriteFile(profile, []byte(content), 0644)
			if err != nil {
				fmt.Printf("Error updating %s: %v\n", profile, err)
			} else {
				fmt.Printf("Updated %s\n", profile)
			}
		}
	}
	fmt.Println("\nIMPORTANT: To apply changes, please restart your terminal or run source on your profile.")
	return nil
}

