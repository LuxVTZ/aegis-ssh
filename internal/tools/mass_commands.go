package tools

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/LuxVTZ/aegis-ssh/internal/config"
	"github.com/LuxVTZ/aegis-ssh/internal/ssh"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterMassCommandTools(s *server.MCPServer, cfgManager *config.Manager) {
	s.AddTool(mcp.NewTool("execute_on_multiple",
		mcp.WithDescription("Execute a command on multiple VPS servers simultaneously"),
		mcp.WithString("hosts", mcp.Required(), mcp.Description("Comma separated list of host names, or 'tag:my_tag', or '*' for all")),
		mcp.WithString("command", mcp.Required(), mcp.Description("Shell command to execute")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments map"), nil
		}

		hostsStr, ok := args["hosts"].(string)
		if !ok {
			return mcp.NewToolResultError("invalid hosts argument"), nil
		}
		cmd, ok := args["command"].(string)
		if !ok {
			return mcp.NewToolResultError("invalid command argument"), nil
		}

		var targetMachines []config.MachineConfig
		allMachines := cfgManager.ListMachines()

		if hostsStr == "*" {
			targetMachines = allMachines
		} else if strings.HasPrefix(hostsStr, "tag:") {
			tag := strings.TrimPrefix(hostsStr, "tag:")
			for _, m := range allMachines {
				for _, t := range m.Tags {
					if t == tag {
						targetMachines = append(targetMachines, m)
						break
					}
				}
			}
		} else {
			hostNames := strings.Split(hostsStr, ",")
			for _, name := range hostNames {
				name = strings.TrimSpace(name)
				if m, err := cfgManager.GetMachine(name); err == nil {
					targetMachines = append(targetMachines, m)
				}
			}
		}

		if len(targetMachines) == 0 {
			return mcp.NewToolResultText("No machines matched the hosts criteria."), nil
		}

		var wg sync.WaitGroup
		var mu sync.Mutex
		results := make([]string, 0, len(targetMachines))

		for _, machine := range targetMachines {
			wg.Add(1)
			go func(m config.MachineConfig) {
				defer wg.Done()

				client, err := ssh.NewClient(m)
				if err != nil {
					mu.Lock()
					results = append(results, fmt.Sprintf("Host: %s\nError connecting: %v\n", m.Name, err))
					mu.Unlock()
					return
				}
				defer client.Close()

				stdout, stderr, exitCode, err := client.ExecuteCommand(cmd)

				mu.Lock()
				if err != nil {
					results = append(results, fmt.Sprintf("Host: %s\nError executing: %v\n", m.Name, err))
				} else {
					results = append(results, fmt.Sprintf("Host: %s (Exit %d)\nSTDOUT:\n%s\nSTDERR:\n%s\n", m.Name, exitCode, strings.TrimSpace(stdout), strings.TrimSpace(stderr)))
				}
				mu.Unlock()

			}(machine)
		}

		wg.Wait()

		return mcp.NewToolResultText(strings.Join(results, "\n-------------------\n")), nil
	})
}
