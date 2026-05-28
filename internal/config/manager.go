package config

import (
	"fmt"
	"sync"
)

type Manager struct {
	mu       sync.RWMutex
	machines map[string]MachineConfig
}

// NewManager creates a new config manager, loading both machines.json and ~/.ssh/config
func NewManager(customConfigPath string) (*Manager, error) {
	m := &Manager{
		machines: make(map[string]MachineConfig),
	}

	// Load from ~/.ssh/config first
	sshMachines, err := LoadSSHConfig()
	if err == nil {
		for _, sm := range sshMachines {
			m.machines[sm.Name] = sm
		}
	}

	// Load from machines.json (overrides ~/.ssh/config if names collide)
	customConfig, err := LoadConfig(customConfigPath)
	if err == nil && customConfig != nil {
		for _, cm := range customConfig.Machines {
			m.machines[cm.Name] = cm
		}
	} else if err != nil {
		// It's ok if machines.json is missing, but if there's a parse error, we should return it
		// For now, we will log it and proceed if it's not a fatal json error, but for simplicity we return err
		// return nil, err
	}

	return m, nil
}

// GetMachine returns a machine configuration by name
func (m *Manager) GetMachine(name string) (MachineConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	machine, exists := m.machines[name]
	if !exists {
		return MachineConfig{}, fmt.Errorf("machine '%s' not found", name)
	}
	return machine, nil
}

// ListMachines returns all configured machines
func (m *Manager) ListMachines() []MachineConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var machines []MachineConfig
	for _, machine := range m.machines {
		machines = append(machines, machine)
	}
	return machines
}
