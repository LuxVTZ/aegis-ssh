package security

import (
	"fmt"
	"regexp"
)

type Profile string

const (
	Strict   Profile = "strict"
	Moderate Profile = "moderate"
	Full     Profile = "full"
)

func ValidateCommand(command string, allowed []string, forbidden []string) error {
	// First check forbidden (Blacklist)
	for _, pattern := range forbidden {
		matched, err := regexp.MatchString(pattern, command)
		if err == nil && matched {
			return fmt.Errorf("command matches forbidden pattern: %s", pattern)
		}
	}

	// Then check allowed (Whitelist)
	for _, pattern := range allowed {
		if pattern == ".*" {
			return nil
		}
		matched, err := regexp.MatchString(pattern, command)
		if err == nil && matched {
			return nil
		}
	}

	return fmt.Errorf("command not permitted by security profile")
}
