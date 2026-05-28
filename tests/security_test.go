package tests

import (
	"testing"

	"github.com/LuxVTZ/aegis-ssh/internal/security"
)

func TestValidateCommand_Whitelist(t *testing.T) {
	allowed := []string{"^ls.*", "^echo.*"}
	forbidden := []string{}

	if err := security.ValidateCommand("ls -la", allowed, forbidden); err != nil {
		t.Errorf("Expected command 'ls -la' to be allowed, got error: %v", err)
	}

	if err := security.ValidateCommand("echo hello", allowed, forbidden); err != nil {
		t.Errorf("Expected command 'echo hello' to be allowed, got error: %v", err)
	}

	if err := security.ValidateCommand("rm -rf /", allowed, forbidden); err == nil {
		t.Errorf("Expected command 'rm -rf /' to be forbidden, but it was allowed")
	}
}

func TestValidateCommand_Blacklist(t *testing.T) {
	allowed := []string{".*"}
	forbidden := []string{".*rm\\s+-rf.*", "^reboot"}

	if err := security.ValidateCommand("rm -rf /var/www", allowed, forbidden); err == nil {
		t.Errorf("Expected 'rm -rf' to be caught by blacklist")
	}

	if err := security.ValidateCommand("reboot", allowed, forbidden); err == nil {
		t.Errorf("Expected 'reboot' to be caught by blacklist")
	}

	if err := security.ValidateCommand("cat /var/log/syslog", allowed, forbidden); err != nil {
		t.Errorf("Expected 'cat' to be allowed, got error: %v", err)
	}
}
