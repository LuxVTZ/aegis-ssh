package tools

import (
	"context"

	"github.com/LuxVTZ/aegis-ssh/internal/config"
	"github.com/LuxVTZ/aegis-ssh/internal/playbook"
	"github.com/LuxVTZ/aegis-ssh/internal/ssh"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterPlaybookTools(s *server.MCPServer, cfgManager *config.Manager) {
	s.AddTool(mcp.NewTool("run_playbook",
		mcp.WithDescription("Run a YAML playbook to ideopotently configure a server"),
		mcp.WithString("host", mcp.Required(), mcp.Description("Name of the host")),
		mcp.WithString("yaml_content", mcp.Required(), mcp.Description("Raw YAML string of the playbook")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments map"), nil
		}

		host, _ := args["host"].(string)
		yamlContent, _ := args["yaml_content"].(string)

		machine, err := cfgManager.GetMachine(host)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		client, err := ssh.NewClient(machine)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer client.Close()

		result, err := playbook.RunPlaybook(client, yamlContent)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(result), nil
	})
}
