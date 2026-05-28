package tools

import (
	"context"
	"fmt"

	"github.com/LuxVTZ/aegis-ssh/internal/config"
	"github.com/LuxVTZ/aegis-ssh/internal/ssh"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterProcessTools(s *server.MCPServer, cfgManager *config.Manager) {
	s.AddTool(mcp.NewTool("manage_process",
		mcp.WithDescription("Manage processes on remote VPS server (systemd, pm2)"),
		mcp.WithString("host", mcp.Required(), mcp.Description("Name of the host")),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action (start, stop, restart, status)")),
		mcp.WithString("process_name", mcp.Required(), mcp.Description("Name of the process")),
		mcp.WithString("service_manager", mcp.Description("systemd or pm2")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments map"), nil
		}

		host, _ := args["host"].(string)
		action, _ := args["action"].(string)
		process, _ := args["process_name"].(string)
		manager, ok := args["service_manager"].(string)
		if !ok || manager == "" || manager == "auto" {
			manager = "systemd" // Default
		}

		var cmd string
		if manager == "systemd" {
			cmd = fmt.Sprintf("sudo systemctl %s %s", action, process)
		} else if manager == "pm2" {
			cmd = fmt.Sprintf("pm2 %s %s", action, process)
		} else {
			return mcp.NewToolResultError("unsupported manager"), nil
		}

		machine, err := cfgManager.GetMachine(host)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		client, err := ssh.NewClient(machine)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer client.Close()

		stdout, stderr, exitCode, err := client.ExecuteCommand(cmd)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed: %v", err)), nil
		}

		result := fmt.Sprintf("Manager: %s\nAction: %s\nProcess: %s\nExit Code: %d\nStdout:\n%s\nStderr:\n%s", manager, action, process, exitCode, stdout, stderr)
		return mcp.NewToolResultText(result), nil
	})
}
