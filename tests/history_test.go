package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LuxVTZ/aegis-ssh/internal/history"
)

func TestHistoryManager(t *testing.T) {
	// Temporarily override the DB directory to a tmp path for tests
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir) // Mock HOME directory so DB writes to tmp

	manager, err := history.NewManager()
	if err != nil {
		t.Fatalf("Failed to initialize History Manager: %v", err)
	}
	defer manager.Close()

	// Verify DB file was created
	dbPath := filepath.Join(tmpDir, ".sshmcp", "audit.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file was not created at %s", dbPath)
	}

	// Test Insert
	err = manager.LogCommand("prod-server", "ls -la", 0, 150)
	if err != nil {
		t.Errorf("Failed to log command: %v", err)
	}
}
