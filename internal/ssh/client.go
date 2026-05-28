package ssh

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LuxVTZ/aegis-ssh/internal/config"
	"github.com/LuxVTZ/aegis-ssh/internal/security"
	"golang.org/x/crypto/ssh"
)

// Client wraps an active SSH connection
type Client struct {
	machine config.MachineConfig
	client  *ssh.Client
}

// NewClient creates a new SSH client connection for a given machine
func NewClient(machine config.MachineConfig) (*Client, error) {
	var authMethod ssh.AuthMethod

	if machine.Auth.Type == "key" {
		keyPath := machine.Auth.KeyPath
		if strings.HasPrefix(keyPath, "~/") {
			homeDir, _ := os.UserHomeDir()
			keyPath = filepath.Join(homeDir, keyPath[2:])
		}

		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read private key: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			// Try parsing with passphrase if needed, but keeping it simple for now
			return nil, fmt.Errorf("unable to parse private key: %v", err)
		}
		authMethod = ssh.PublicKeys(signer)
	} else if machine.Auth.Type == "password" {
		authMethod = ssh.Password(machine.Auth.Password)
	} else {
		return nil, fmt.Errorf("unsupported auth type: %s", machine.Auth.Type)
	}

	sshConfig := &ssh.ClientConfig{
		User: machine.User,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For production, should use known_hosts
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", machine.Host, machine.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	return &Client{
		machine: machine,
		client:  client,
	}, nil
}

// Close closes the underlying SSH connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// ExecuteCommand executes a single command and returns its output
func (c *Client) ExecuteCommand(command string) (string, string, int, error) {
	if err := security.ValidateCommand(command, c.machine.Security.AllowedCommands, c.machine.Security.ForbiddenCommands); err != nil {
		return "", "", -1, fmt.Errorf("security violation: %v", err)
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", "", -1, fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	err = session.Run(command)

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*ssh.ExitError); ok {
			exitCode = exitError.ExitStatus()
		} else {
			exitCode = -1
		}
	}

	return stdoutBuf.String(), stderrBuf.String(), exitCode, nil
}
