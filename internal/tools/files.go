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

func RegisterFileTools(s *server.MCPServer, cfgManager *config.Manager) {
	s.AddTool(mcp.NewTool("read_file",
		mcp.WithDescription("Read file content from remote VPS server"),
		mcp.WithString("host", mcp.Required(), mcp.Description("Name of the host")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the file on remote server")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments map"), nil
		}
		host, _ := args["host"].(string)
		path, _ := args["path"].(string)

		machine, err := cfgManager.GetMachine(host)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		client, err := ssh.NewClient(machine)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer client.Close()

		content, err := client.ReadFile(path)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(content), nil
	})

	s.AddTool(mcp.NewTool("upload_file",
		mcp.WithDescription("Upload file to remote VPS server"),
		mcp.WithString("host", mcp.Required(), mcp.Description("Name of the host")),
		mcp.WithString("path", mcp.Required(), mcp.Description("Destination path")),
		mcp.WithString("content", mcp.Required(), mcp.Description("File content")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := request.Params.Arguments.(map[string]interface{})
		host, _ := args["host"].(string)
		path, _ := args["path"].(string)
		content, _ := args["content"].(string)

		machine, err := cfgManager.GetMachine(host)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		client, err := ssh.NewClient(machine)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer client.Close()

		if err := client.UploadFile(path, content); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Successfully uploaded to %s", path)), nil
	})

	s.AddTool(mcp.NewTool("list_files",
		mcp.WithDescription("List files in directory on remote VPS server"),
		mcp.WithString("host", mcp.Required(), mcp.Description("Name of the host")),
		mcp.WithString("directory", mcp.Required(), mcp.Description("Path to directory")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := request.Params.Arguments.(map[string]interface{})
		host, _ := args["host"].(string)
		dir, _ := args["directory"].(string)

		machine, err := cfgManager.GetMachine(host)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		client, err := ssh.NewClient(machine)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer client.Close()

		files, err := client.ListFiles(dir)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(strings.Join(files, "\n")), nil
	})
}
