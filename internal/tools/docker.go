package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/LuxVTZ/aegis-ssh/internal/config"
	"github.com/LuxVTZ/aegis-ssh/internal/ssh"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterDockerTools(s *server.MCPServer, cfgManager *config.Manager) {
	s.AddTool(mcp.NewTool("docker_ps",
		mcp.WithDescription("Get a structured JSON list of running Docker containers on the remote host"),
		mcp.WithString("host", mcp.Required(), mcp.Description("Name of the host")),
		mcp.WithBoolean("all", mcp.Description("Show all containers (default shows just running)")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments map"), nil
		}

		host, ok := args["host"].(string)
		if !ok {
			return mcp.NewToolResultError("invalid host argument"), nil
		}
		
		allFlag := ""
		if all, ok := args["all"].(bool); ok && all {
			allFlag = "-a"
		}

		cmd := fmt.Sprintf("sudo docker ps %s --format '{{json .}}'", allFlag)
		
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
		if err != nil || exitCode != 0 {
			return mcp.NewToolResultError(fmt.Sprintf("Docker command failed: %v\n%s", err, stderr)), nil
		}

		// Docker formats each line as a separate JSON object. We'll wrap them in an array.
		lines := strings.Split(strings.TrimSpace(stdout), "\n")
		jsonArray := "[\n" + strings.Join(lines, ",\n") + "\n]"
		
		return mcp.NewToolResultText(jsonArray), nil
	})

	s.AddTool(mcp.NewTool("docker_logs",
		mcp.WithDescription("Get logs from a specific Docker container"),
		mcp.WithString("host", mcp.Required(), mcp.Description("Name of the host")),
		mcp.WithString("container", mcp.Required(), mcp.Description("Container name or ID")),
		mcp.WithNumber("lines", mcp.Description("Number of lines to tail (default 100)")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := request.Params.Arguments.(map[string]interface{})
		host, _ := args["host"].(string)
		container, _ := args["container"].(string)
		
		lines := 100.0
		if l, ok := args["lines"].(float64); ok {
			lines = l
		}

		cmd := fmt.Sprintf("sudo docker logs --tail %d %s", int(lines), container)
		
		machine, err := cfgManager.GetMachine(host)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		client, err := ssh.NewClient(machine)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer client.Close()

		stdout, stderr, _, _ := client.ExecuteCommand(cmd)
		output := stdout
		if output == "" {
			output = stderr // Docker often outputs logs to stderr
		}

		return mcp.NewToolResultText(output), nil
	})
}
