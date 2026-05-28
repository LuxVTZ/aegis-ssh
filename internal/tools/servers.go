package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/LuxVTZ/aegis-ssh/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterServerTools registers tools related to server management
func RegisterServerTools(s *server.MCPServer, cfgManager *config.Manager) {
	s.AddTool(mcp.NewTool("list_servers",
		mcp.WithDescription("List all configured VPS servers"),
		mcp.WithString("tag", mcp.Description("Optional tag to filter servers")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		args, ok := request.Params.Arguments.(map[string]interface{})
		tagFilter := ""
		if ok && args["tag"] != nil {
			if tagArg, ok := args["tag"].(string); ok {
				tagFilter = tagArg
			}
		}

		machines := cfgManager.ListMachines()
		var output strings.Builder

		count := 0
		for _, m := range machines {
			if tagFilter != "" {
				hasTag := false
				for _, t := range m.Tags {
					if t == tagFilter {
						hasTag = true
						break
					}
				}
				if !hasTag {
					continue
				}
			}

			output.WriteString(fmt.Sprintf("- %s (%s@%s:%d)\n  Tags: %v\n  Desc: %s\n\n",
				m.Name, m.User, m.Host, m.Port, m.Tags, m.Description))
			count++
		}

		if count == 0 {
			return mcp.NewToolResultText("No servers found matching criteria."), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Found %d servers:\n\n%s", count, output.String())), nil
	})
}
