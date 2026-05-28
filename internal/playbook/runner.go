package playbook

import (
	"fmt"
	"strings"

	"github.com/LuxVTZ/aegis-ssh/internal/ssh"
	"gopkg.in/yaml.v3"
)

type Playbook struct {
	Name  string `yaml:"name"`
	Tasks []Task `yaml:"tasks"`
}

type Task struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type"` // "shell", "apt", "service"
	Command string `yaml:"command"`
}

func RunPlaybook(client *ssh.Client, yamlContent string) (string, error) {
	var pb Playbook
	if err := yaml.Unmarshal([]byte(yamlContent), &pb); err != nil {
		return "", fmt.Errorf("invalid playbook yaml: %v", err)
	}

	var results strings.Builder
	results.WriteString(fmt.Sprintf("Running playbook: %s\n", pb.Name))

	for _, task := range pb.Tasks {
		results.WriteString(fmt.Sprintf("\n[TASK] %s\n", task.Name))
		cmd := ""

		switch task.Type {
		case "shell":
			cmd = task.Command
		case "apt":
			cmd = fmt.Sprintf("sudo DEBIAN_FRONTEND=noninteractive apt-get install -y %s", task.Command)
		case "service":
			cmd = fmt.Sprintf("sudo systemctl restart %s", task.Command)
		default:
			results.WriteString(fmt.Sprintf("FAILED: Unknown task type '%s'\n", task.Type))
			continue
		}

		stdout, stderr, exitCode, err := client.ExecuteCommand(cmd)
		if err != nil {
			results.WriteString(fmt.Sprintf("ERROR: %v\n", err))
		} else if exitCode != 0 {
			results.WriteString(fmt.Sprintf("FAILED (Code %d): %s\n", exitCode, stderr))
		} else {
			results.WriteString(fmt.Sprintf("SUCCESS:\n%s\n", stdout))
		}
	}

	return results.String(), nil
}
