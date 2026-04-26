package shell

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"shortk/internal/config"
	"strings"
)

func ExecuteAlias(short string, args []string) {
	localCmd := ""
	localEnvPath := ""
	cwd, _ := os.Getwd()
	dir := cwd

	// 1. Traverse up to find .shortk
	for dir != "" && dir != "." {
		shortkPath := filepath.Join(dir, config.LocalFilename)
		if _, err := os.Stat(shortkPath); err == nil {
			file, err := os.Open(shortkPath)
			if err == nil {
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					line := scanner.Text()
					if strings.HasPrefix(line, short+"=") {
						localCmd = strings.TrimPrefix(line, short+"=")
						localEnvPath = filepath.Join(dir, ".env")
						break
					}
				}
				file.Close()
			}
		}
		if localCmd != "" {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// 2. Fallback to global
	if localCmd == "" {
		aliases, err := config.LoadConfig()
		if err == nil {
			if cmd, exists := aliases[short]; exists && cmd != "" {
				localCmd = cmd
				localEnvPath = filepath.Join(cwd, ".env")
			}
		}
	}

	if localCmd == "" {
		fmt.Printf("shortk: '%s' is not defined globally and no local override found.\n", short)
		os.Exit(1)
	}

	// 3. Load .env if exists
	if _, err := os.Stat(localEnvPath); err == nil {
		file, err := os.Open(localEnvPath)
		if err == nil {
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				// Support both KEY=VALUE and export KEY=VALUE
				line = strings.TrimPrefix(line, "export ")
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					// Basic quote removal
					value = strings.Trim(value, `"'`)
					os.Setenv(key, value)
				}
			}
			file.Close()
		}
	}

	// 4. Execute
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		fullCmd := localCmd
		for _, arg := range args {
			if strings.Contains(arg, " ") {
				fullCmd += fmt.Sprintf(` "%s"`, arg)
			} else {
				fullCmd += " " + arg
			}
		}
		cmd = exec.Command("cmd", "/c", fullCmd)
	} else {
		fullCmd := localCmd
		for _, arg := range args {
			// Basic escaping for bash - simplified
			if strings.ContainsAny(arg, " \"'$") {
				fullCmd += fmt.Sprintf(" %q", arg)
			} else {
				fullCmd += " " + arg
			}
		}
		cmd = exec.Command("bash", "-c", fullCmd)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}
}
