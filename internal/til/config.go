package til

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// LoadConfig loads the configuration from the .til/config file
func LoadConfig(dir string) (Config, error) {
	config := Config{
		DataDir: dir,
	}

	// Read configuration file
	configFile := filepath.Join(dir, ".til", "config")
	file, err := os.Open(configFile)
	if err != nil {
		return config, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "SYNC_TO_NOTION":
			config.SyncToNotion = value == "true"
		case "NOTION_API_KEY":
			config.NotionAPIKey = value
		case "NOTION_DB_ID":
			config.NotionDBID = value
		}
	}

	return config, scanner.Err()
}

// SaveConfig saves the configuration to the .til/config file
func SaveConfig(config Config) error {
	configDir := filepath.Join(config.DataDir, ".til")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "config")
	f, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write configuration to file
	if config.SyncToNotion {
		f.WriteString("SYNC_TO_NOTION=true\n")
		f.WriteString("NOTION_API_KEY=" + config.NotionAPIKey + "\n")
		f.WriteString("NOTION_DB_ID=" + config.NotionDBID + "\n")
	} else {
		f.WriteString("SYNC_TO_NOTION=false\n")
	}

	return nil
}
