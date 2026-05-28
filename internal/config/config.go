package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// AuthConfig represents the authentication settings for a machine
type AuthConfig struct {
	Type     string `json:"type"` // "key" or "password"
	KeyPath  string `json:"key_path,omitempty"`
	Password string `json:"password,omitempty"`
}

// SecurityConfig represents security policies for a machine
type SecurityConfig struct {
	AllowedCommands   []string `json:"allowed_commands"`
	ForbiddenCommands []string `json:"forbidden_commands"`
	TimeoutSeconds    int      `json:"timeout_seconds"`
}

// MachineConfig represents a single server configuration
type MachineConfig struct {
	Name        string         `json:"name"`
	Host        string         `json:"host"`
	Port        int            `json:"port"`
	User        string         `json:"user"`
	Auth        AuthConfig     `json:"auth"`
	Security    SecurityConfig `json:"security"`
	Tags        []string       `json:"tags,omitempty"`
	Description string         `json:"description,omitempty"`
}

// MachinesConfig is the root configuration structure
type MachinesConfig struct {
	Machines []MachineConfig `json:"machines"`
}

// GetDefaultConfigPath returns the default path: ~/.sshmcp/machines.json
func GetDefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".sshmcp", "machines.json"), nil
}

// LoadConfig loads the machines configuration from the specified path or default path
func LoadConfig(path string) (*MachinesConfig, error) {
	if path == "" {
		defaultPath, err := GetDefaultConfigPath()
		if err != nil {
			return nil, err
		}
		path = defaultPath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Return empty config if file doesn't exist
			return &MachinesConfig{Machines: []MachineConfig{}}, nil
		}
		return nil, err
	}

	var config MachinesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
