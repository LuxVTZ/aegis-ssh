package tools

import (
	"context"

	"github.com/LuxVTZ/aegis-ssh/internal/config"
	"github.com/LuxVTZ/aegis-ssh/internal/ssh"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterShellTools(s *server.MCPServer, cfgManager *config.Manager) {
	s.AddTool(mcp.NewTool("shell_start",
		mcp.WithDescription("Start an interactive shell session"),
		mcp.WithString("host", mcp.Required(), mcp.Description("Name of the host")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := request.Params.Arguments.(map[string]interface{})
		host, _ := args["host"].(string)

		machine, err := cfgManager.GetMachine(host)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		client, err := ssh.NewClient(machine)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		session, err := ssh.GlobalPTYManager.CreateSession(client)
		if err != nil {
			client.Close()
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText("Session started. ID: " + session.ID), nil
	})

	s.AddTool(mcp.NewTool("shell_send",
		mcp.WithDescription("Send text to interactive shell"),
		mcp.WithString("session_id", mcp.Required(), mcp.Description("Session ID")),
		mcp.WithString("data", mcp.Required(), mcp.Description("Text to send")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := request.Params.Arguments.(map[string]interface{})
		id, _ := args["session_id"].(string)
		data, _ := args["data"].(string)

		session, err := ssh.GlobalPTYManager.GetSession(id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if err := session.Write(data + "\n"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText("Sent."), nil
	})

	s.AddTool(mcp.NewTool("shell_recv",
		mcp.WithDescription("Receive output from interactive shell"),
		mcp.WithString("session_id", mcp.Required(), mcp.Description("Session ID")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := request.Params.Arguments.(map[string]interface{})
		id, _ := args["session_id"].(string)

		session, err := ssh.GlobalPTYManager.GetSession(id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		output := session.Read()
		return mcp.NewToolResultText(output), nil
	})

	s.AddTool(mcp.NewTool("shell_close",
		mcp.WithDescription("Close an interactive shell"),
		mcp.WithString("session_id", mcp.Required(), mcp.Description("Session ID")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := request.Params.Arguments.(map[string]interface{})
		id, _ := args["session_id"].(string)

		ssh.GlobalPTYManager.CloseSession(id)
		return mcp.NewToolResultText("Session closed."), nil
	})
}
