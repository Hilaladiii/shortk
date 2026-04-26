package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ConfigDir     = func() string {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "shortk")
	}()
	AliasesFile   = filepath.Join(ConfigDir, "aliases.json")
	LocalFilename = ".shortk"
)

func EnsureConfigDir() error {
	if _, err := os.Stat(ConfigDir); os.IsNotExist(err) {
		return os.MkdirAll(ConfigDir, 0755)
	}
	return nil
}

func LoadConfig() (map[string]string, error) {
	if err := EnsureConfigDir(); err != nil {
		return nil, err
	}
	if _, err := os.Stat(AliasesFile); os.IsNotExist(err) {
		return make(map[string]string), nil
	}
	data, err := os.ReadFile(AliasesFile)
	if err != nil {
		return nil, err
	}
	var config map[string]string
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return config, nil
}

func SaveConfig(config map[string]string) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(AliasesFile, data, 0644)
}

func FindLocalConfigPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	currentDir := cwd
	for {
		localPath := filepath.Join(currentDir, LocalFilename)
		if _, err := os.Stat(localPath); err == nil {
			return localPath
		}
		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			break
		}
		currentDir = parent
	}
	return ""
}

func LoadLocalConfig() (map[string]string, error) {
	localPath := FindLocalConfigPath()
	if localPath == "" {
		return make(map[string]string), nil
	}
	file, err := os.Open(localPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" {
				config[key] = value
			}
		}
	}
	return config, scanner.Err()
}

func SaveLocalConfig(config map[string]string) error {
	localPath := FindLocalConfigPath()
	if localPath == "" {
		cwd, _ := os.Getwd()
		localPath = filepath.Join(cwd, LocalFilename)
	}
	var lines []string
	for k, v := range config {
		lines = append(lines, fmt.Sprintf("%s=%s", k, v))
	}
	data := strings.Join(lines, "\n")
	err := os.WriteFile(localPath, []byte(data), 0644)
	if err != nil {
		return err
	}
	return AddToGitIgnore(localPath)
}

func AddToGitIgnore(shortkPath string) error {
	dir := filepath.Dir(shortkPath)
	gitignorePath := filepath.Join(dir, ".gitignore")
	
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return os.WriteFile(gitignorePath, []byte(LocalFilename+"\n"), 0644)
		}
		return nil
	}

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		return err
	}
	if !strings.Contains(string(data), LocalFilename) {
		f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString("\n" + LocalFilename + "\n")
		return err
	}
	return nil
}
