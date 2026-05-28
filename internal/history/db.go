package history

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Manager struct {
	db *sql.DB
}

func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dbDir := filepath.Join(homeDir, ".sshmcp")
	if err := os.MkdirAll(dbDir, 0700); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(dbDir, "audit.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Create table if not exists
	query := `
	CREATE TABLE IF NOT EXISTS command_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		host TEXT,
		command TEXT,
		exit_code INTEGER,
		duration_ms INTEGER
	);`

	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	return &Manager{db: db}, nil
}

func (m *Manager) LogCommand(host, command string, exitCode int, durationMs int64) error {
	query := `INSERT INTO command_history (host, command, exit_code, duration_ms) VALUES (?, ?, ?, ?)`
	_, err := m.db.Exec(query, host, command, exitCode, durationMs)
	return err
}

type AuditLog struct {
	ID         int
	Timestamp  string
	Host       string
	Command    string
	ExitCode   int
	DurationMs int64
}

func (m *Manager) GetRecentLogs(limit int) ([]AuditLog, error) {
	query := `SELECT id, timestamp, host, command, exit_code, duration_ms FROM command_history ORDER BY timestamp DESC LIMIT ?`
	rows, err := m.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(&l.ID, &l.Timestamp, &l.Host, &l.Command, &l.ExitCode, &l.DurationMs); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (m *Manager) Close() error {
	return m.db.Close()
}

// Global instance for simple access from tools
var GlobalManager *Manager

func InitGlobalManager() error {
	m, err := NewManager()
	if err != nil {
		return err
	}
	GlobalManager = m
	return nil
}
