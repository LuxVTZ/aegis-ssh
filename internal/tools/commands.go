package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/LuxVTZ/aegis-ssh/internal/config"
	"github.com/LuxVTZ/aegis-ssh/internal/history"
	"github.com/LuxVTZ/aegis-ssh/internal/ssh"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterCommandTools registers command execution tools
func RegisterCommandTools(s *server.MCPServer, cfgManager *config.Manager) {
	s.AddTool(mcp.NewTool("execute_command",
		mcp.WithDescription("Execute a command on remote VPS server via SSH"),
		mcp.WithString("host", mcp.Required(), mcp.Description("Name of the host")),
		mcp.WithString("command", mcp.Required(), mcp.Description("Shell command to execute")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments map"), nil
		}

		host, ok := args["host"].(string)
		if !ok {
			return mcp.NewToolResultError("invalid host argument"), nil
		}

		cmd, ok := args["command"].(string)
		if !ok {
			return mcp.NewToolResultError("invalid command argument"), nil
		}

		machine, err := cfgManager.GetMachine(host)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("machine error: %v", err)), nil
		}

		client, err := ssh.NewClient(machine)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("connection error: %v", err)), nil
		}
		defer client.Close()

		start := time.Now()
		stdout, stderr, exitCode, err := client.ExecuteCommand(cmd)
		duration := time.Since(start).Milliseconds()

		if history.GlobalManager != nil {
			go history.GlobalManager.LogCommand(host, cmd, exitCode, duration)
		}

		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("execution failed: %v", err)), nil
		}

		result := fmt.Sprintf("Exit Code: %d\n\nSTDOUT:\n%s\n\nSTDERR:\n%s", exitCode, stdout, stderr)
		return mcp.NewToolResultText(result), nil
	})
}
