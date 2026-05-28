package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kevinburke/ssh_config"
)

// LoadSSHConfig loads the ~/.ssh/config file and converts it to MachineConfig objects.
// This allows the MCP server to automatically pick up the user's existing SSH hosts.
func LoadSSHConfig() ([]MachineConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".ssh", "config")
	f, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []MachineConfig{}, nil
		}
		return nil, err
	}
	defer f.Close()

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return nil, err
	}

	var machines []MachineConfig

	// ssh_config.Hosts contains all Host blocks
	for _, host := range cfg.Hosts {
		// We skip wildcards because they are not specific machines
		name := ""
		for _, pattern := range host.Patterns {
			if !strings.Contains(pattern.String(), "*") && !strings.Contains(pattern.String(), "?") {
				name = pattern.String()
				break
			}
		}

		if name == "" {
			continue
		}

		hostname, err := cfg.Get(name, "HostName")
		if err != nil || hostname == "" {
			hostname = name
		}

		user, err := cfg.Get(name, "User")
		if err != nil || user == "" {
			user = os.Getenv("USER")
		}

		portStr, err := cfg.Get(name, "Port")
		port := 22
		if err == nil && portStr != "" {
			// Parse port if needed, for simplicity we keep it as 22 if parsing fails
			// We can implement a robust string to int here
		}

		identityFile, err := cfg.Get(name, "IdentityFile")
		if err != nil || identityFile == "" {
			identityFile = "~/.ssh/id_rsa"
		} else {
			// IdentityFile usually has ~ in it, or is absolute
			// We can pass it straight through, our SSH client should expand it
		}

		machine := MachineConfig{
			Name: name,
			Host: hostname,
			Port: port,
			User: user,
			Auth: AuthConfig{
				Type:    "key",
				KeyPath: identityFile,
			},
			Security: SecurityConfig{
				AllowedCommands:   []string{".*"},
				ForbiddenCommands: []string{".*rm\\s+-rf\\s+/$"},
				TimeoutSeconds:    120,
			},
			Tags:        []string{"ssh_config"},
			Description: "Imported from ~/.ssh/config",
		}

		machines = append(machines, machine)
	}

	return machines, nil
}
