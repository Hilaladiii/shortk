package cli

import (
	"fmt"
	"shortk/internal/autocomplete"
	"shortk/internal/config"
	"shortk/internal/shell"
	"sort"
	"strings"
)

func ShowHelp() {
	fmt.Println(`
Usage: shortk <command> [args]

Commands:
  init                       Set up shell integration (PATH)
  add <short> "<long>"       Add a new short command
  remove <short>             Remove a short command
  list                       List all short commands
  status <short>             Check if command is local or global
  completion                 Generate shell completion script
  help                       Show this help message

Options:
  --local                    Apply command to local .shortk file`)
}

func ListAliases(isLocal bool) {
	var aliases map[string]string
	var err error
	var sourceInfo string

	if isLocal {
		localPath := config.FindLocalConfigPath()
		if localPath != "" {
			aliases, err = config.LoadLocalConfig()
			sourceInfo = fmt.Sprintf(" (from %s)", localPath)
		} else {
			aliases = make(map[string]string)
		}
	} else {
		aliases, err = config.LoadConfig()
	}

	if err != nil {
		fmt.Printf("Error loading aliases: %v\n", err)
		return
	}

	if len(aliases) == 0 {
		msg := "No short commands configured yet"
		if isLocal && sourceInfo == "" {
			msg += " (no .shortk file found)"
		}
		fmt.Printf("%s.\n", msg)
		return
	}

	fmt.Printf("Your %sshort commands%s:\n", func() string {
		if isLocal {
			return "local "
		}
		return ""
	}(), sourceInfo)

	var keys []string
	for k := range aliases {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Printf("  %-15s ->  %s\n", k, aliases[k])
	}
}

func AddAlias(short, long string, isLocal bool) {
	if short == "" || long == "" {
		fmt.Println("Error: Both <short> and <long> commands are required.")
		return
	}

	if isLocal {
		aliases, err := config.LoadLocalConfig()
		if err != nil {
			fmt.Printf("Error loading local config: %v\n", err)
			return
		}
		aliases[short] = long
		if err := config.SaveLocalConfig(aliases); err != nil {
			fmt.Printf("Error saving local config: %v\n", err)
			return
		}

		// Create global wrapper if it doesn't exist
		globalAliases, _ := config.LoadConfig()
		if _, ok := globalAliases[short]; !ok {
			shell.CreateWrapper(short, "")
		}
		fmt.Printf("Successfully added LOCAL: %s -> %s (Saved to %s)\n", short, long, config.LocalFilename)
	} else {
		aliases, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}
		aliases[short] = long
		if err := config.SaveConfig(aliases); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		if err := shell.CreateWrapper(short, long); err != nil {
			fmt.Printf("Error creating wrapper: %v\n", err)
			return
		}
		fmt.Printf("Successfully added GLOBAL: %s -> %s\n", short, long)
	}
}

func RemoveAlias(short string, isLocal bool) {
	if short == "" {
		fmt.Println("Error: <short> command is required.")
		return
	}

	if isLocal {
		aliases, err := config.LoadLocalConfig()
		if err != nil {
			fmt.Printf("Error loading local config: %v\n", err)
			return
		}
		if _, ok := aliases[short]; !ok {
			fmt.Printf("Error: Local short command \"%s\" not found in %s\n", short, config.LocalFilename)
			return
		}
		delete(aliases, short)
		if err := config.SaveLocalConfig(aliases); err != nil {
			fmt.Printf("Error saving local config: %v\n", err)
			return
		}
		fmt.Printf("Successfully removed LOCAL: %s\n", short)
	} else {
		aliases, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}
		if _, ok := aliases[short]; !ok {
			fmt.Printf("Error: Global short command \"%s\" not found.\n", short)
			return
		}
		delete(aliases, short)
		if err := config.SaveConfig(aliases); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}
		shell.RemoveWrapper(short)
		fmt.Printf("Successfully removed GLOBAL: %s\n", short)
	}
}

func CheckStatus(short string) {
	if short == "" {
		fmt.Println("Error: <short> command is required.")
		return
	}

	localPath := config.FindLocalConfigPath()
	var localCmd string
	if localPath != "" {
		localConfig, _ := config.LoadLocalConfig()
		localCmd = localConfig[short]
	}

	globalConfig, _ := config.LoadConfig()
	globalCmd := globalConfig[short]

	if localCmd == "" && globalCmd == "" {
		fmt.Printf("'%s' is not defined.\n", short)
		return
	}

	fmt.Printf("Status for '%s':\n", short)
	if localCmd != "" {
		fmt.Printf("  LOCAL:  %s (from %s)\n", localCmd, localPath)
		if globalCmd != "" {
			fmt.Printf("  GLOBAL: %s (OVERRIDDEN)\n", globalCmd)
		}
	} else {
		fmt.Printf("  GLOBAL: %s\n", globalCmd)
		fmt.Println("  LOCAL:  (none found in directory tree)")
	}
}

func ListKeys() {
	globalConfig, _ := config.LoadConfig()
	localConfig, _ := config.LoadLocalConfig()
	
	keySet := make(map[string]struct{})
	for k := range globalConfig {
		keySet[k] = struct{}{}
	}
	for k := range localConfig {
		keySet[k] = struct{}{}
	}

	var keys []string
	for k := range keySet {
		keys = append(keys, k)
	}
	fmt.Print(strings.Join(keys, " "))
}

func Run(args []string) {
	if len(args) == 0 {
		ShowHelp()
		return
	}

	isLocal := false
	var cleanArgs []string
	for _, arg := range args {
		if arg == "--local" {
			isLocal = true
		} else {
			cleanArgs = append(cleanArgs, arg)
		}
	}

	command := cleanArgs[0]
	rest := cleanArgs[1:]

	switch command {
	case "init":
		if err := shell.InitShellProfile(); err == nil {
			shell.SyncWrappers()
		}
	case "add":
		if len(rest) >= 2 {
			AddAlias(rest[0], rest[1], isLocal)
		} else {
			fmt.Println("Error: Both <short> and <long> commands are required.")
		}
	case "remove":
		if len(rest) >= 1 {
			RemoveAlias(rest[0], isLocal)
		} else {
			fmt.Println("Error: <short> command is required.")
		}
	case "list":
		ListAliases(isLocal)
	case "status":
		if len(rest) >= 1 {
			CheckStatus(rest[0])
		} else {
			fmt.Println("Error: <short> command is required.")
		}
	case "_list-keys":
		ListKeys()
	case "completion":
		fmt.Print(autocomplete.GenerateCompletionScript())
	case "help", "--help", "-h":
		ShowHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		ShowHelp()
	}
}
