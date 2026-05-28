package config

import (
	"os"
	"testing"
)

func TestLoadSSHConfig_NoFile(t *testing.T) {
	// If the file doesn't exist, it should return empty, not error
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", "/tmp/nonexistent_home_dir")
	defer os.Setenv("HOME", oldHome)

	machines, err := LoadSSHConfig()
	if err != nil {
		t.Fatalf("Expected no error when file doesn't exist, got: %v", err)
	}
	if len(machines) != 0 {
		t.Fatalf("Expected 0 machines, got %d", len(machines))
	}
}

func TestManager_GetMachine(t *testing.T) {
	m := &Manager{
		machines: map[string]MachineConfig{
			"test": {Name: "test", Host: "127.0.0.1"},
		},
	}

	machine, err := m.GetMachine("test")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if machine.Host != "127.0.0.1" {
		t.Fatalf("Expected host 127.0.0.1, got %s", machine.Host)
	}

	_, err = m.GetMachine("unknown")
	if err == nil {
		t.Fatal("Expected error for unknown machine, got nil")
	}
}
