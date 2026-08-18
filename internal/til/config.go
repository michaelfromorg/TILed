package til

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrConfigNotFound = errors.New("no TIL repository found")

func LoadConfig(dir string) (Config, error) {
	root, err := findConfigRoot(dir)
	if err != nil {
		return Config{DataDir: dir}, err
	}

	config := Config{DataDir: root}
	configFile := filepath.Join(root, metadataDirectoryName, "config")
	file, err := os.Open(configFile)
	if err != nil {
		return config, fmt.Errorf("open configuration: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), "=")
		if !ok {
			continue
		}

		switch strings.TrimSpace(key) {
		case "SYNC_TO_NOTION":
			config.SyncToNotion = strings.TrimSpace(value) == "true"
		case "NOTION_API_KEY":
			config.NotionAPIKey = strings.TrimSpace(value)
		case "NOTION_DB_ID":
			config.NotionDBID = strings.TrimSpace(value)
		case "SYNC_TO_GIT":
			config.SyncToGit = strings.TrimSpace(value) == "true"
		case "GIT_REMOTE_URL":
			config.GitRemoteURL = strings.TrimSpace(value)
		}
	}
	if err := scanner.Err(); err != nil {
		return config, fmt.Errorf("read configuration: %w", err)
	}

	return config, nil
}

func SaveConfig(config Config) error {
	for name, value := range map[string]string{
		"NOTION_API_KEY": config.NotionAPIKey,
		"NOTION_DB_ID":   config.NotionDBID,
		"GIT_REMOTE_URL": config.GitRemoteURL,
	} {
		if strings.ContainsAny(value, "\r\n") {
			return fmt.Errorf("%s cannot contain a newline", name)
		}
	}

	var content bytes.Buffer
	fmt.Fprintf(&content, "SYNC_TO_NOTION=%t\n", config.SyncToNotion)
	if config.SyncToNotion {
		fmt.Fprintf(&content, "NOTION_API_KEY=%s\n", config.NotionAPIKey)
		fmt.Fprintf(&content, "NOTION_DB_ID=%s\n", config.NotionDBID)
	}
	fmt.Fprintf(&content, "SYNC_TO_GIT=%t\n", config.SyncToGit)
	if config.SyncToGit {
		fmt.Fprintf(&content, "GIT_REMOTE_URL=%s\n", config.GitRemoteURL)
	}

	configFile := filepath.Join(config.DataDir, metadataDirectoryName, "config")
	if err := writeFileAtomic(configFile, content.Bytes(), 0600); err != nil {
		return fmt.Errorf("save configuration: %w", err)
	}
	return nil
}

func findConfigRoot(start string) (string, error) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}

	for {
		configPath := filepath.Join(current, metadataDirectoryName, "config")
		info, statErr := os.Stat(configPath)
		if statErr == nil && info.Mode().IsRegular() {
			return current, nil
		}
		if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
			return "", fmt.Errorf("inspect configuration: %w", statErr)
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", fmt.Errorf("%w from %s; run 'til init' first", ErrConfigNotFound, start)
}
