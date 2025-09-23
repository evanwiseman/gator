package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const configFileName = ".gatorconfig.json"

// Get the config path from the users home directory
func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting user home dir: %v", err)
	}
	return home + "/" + configFileName, nil
}

// JSON representation of the config file
type Config struct {
	DBURL    string `json:"db_url"`
	UserName string `json:"current_user_name"`
}

// Read the config file from the home directory and return the config and any errors
func Read() (Config, error) {
	// Get the users config path from their home dir
	filePath, err := getConfigPath()
	if err != nil {
		return Config{}, fmt.Errorf("error getting config path: %v", err)
	}

	// Open the file and defer its Close()
	file, err := os.Open(filePath)
	if err != nil {
		return Config{}, fmt.Errorf("error opening config file: %v", err)
	}
	defer file.Close()

	// Get bytes from the file
	bytes, err := io.ReadAll(file)
	if err != nil {
		return Config{}, fmt.Errorf("error reading config file: %v", err)
	}

	// Unpack the bytes into the Config struct
	var cfg Config
	err = json.Unmarshal(bytes, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("error unmarshalling config bytes: %v", err)
	}

	return cfg, nil
}

// Write the config file to the home directory and return any errors
func write(cfg Config) error {
	// Pack the bytes from the Config struct
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error marshalling config: %v", err)
	}

	// Get the users config path from their home dir
	filePath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("error getting config path: %v", err)
	}

	// Write the bytes to the file path with RD WR perms
	err = os.WriteFile(filePath, data, os.FileMode(os.O_RDWR))
	if err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	return nil
}

// Set the user name in the config and write to the config file, return any errors
func (cfg *Config) SetUserName(userName string) error {
	cfg.UserName = userName
	err := write(*cfg)
	if err != nil {
		return fmt.Errorf("error writing user name: %v", err)
	}

	return nil
}
