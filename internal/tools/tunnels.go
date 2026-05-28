package tools

import (
	"context"
	"fmt"

	"github.com/LuxVTZ/aegis-ssh/internal/config"
	"github.com/LuxVTZ/aegis-ssh/internal/ssh"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterTunnelTools(s *server.MCPServer, cfgManager *config.Manager) {
	s.AddTool(mcp.NewTool("tunnel_expose",
		mcp.WithDescription("Expose a local port to the internet via a remote VPS (Reverse Tunneling)"),
		mcp.WithString("host", mcp.Required(), mcp.Description("Name of the host")),
		mcp.WithNumber("local_port", mcp.Required(), mcp.Description("Local port on this machine")),
		mcp.WithNumber("remote_port", mcp.Required(), mcp.Description("Port to open on the remote server")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments map"), nil
		}

		host, _ := args["host"].(string)
		localPortFloat, _ := args["local_port"].(float64)
		remotePortFloat, _ := args["remote_port"].(float64)

		localPort := int(localPortFloat)
		remotePort := int(remotePortFloat)

		machine, err := cfgManager.GetMachine(host)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		client, err := ssh.NewClient(machine)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		tunnel, err := ssh.GlobalTunnelManager.StartRemoteForwarding(client, remotePort, localPort)
		if err != nil {
			client.Close()
			return mcp.NewToolResultError(err.Error()), nil
		}

		msg := fmt.Sprintf("Tunnel established: Remote %s:%d -> Local 127.0.0.1:%d\nTunnel ID: %s", machine.Host, remotePort, localPort, tunnel.ID)
		return mcp.NewToolResultText(msg), nil
	})

	s.AddTool(mcp.NewTool("tunnel_close",
		mcp.WithDescription("Close an active tunnel"),
		mcp.WithString("tunnel_id", mcp.Required(), mcp.Description("ID of the tunnel")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := request.Params.Arguments.(map[string]interface{})
		id, _ := args["tunnel_id"].(string)

		ssh.GlobalTunnelManager.CloseTunnel(id)
		return mcp.NewToolResultText("Tunnel closed"), nil
	})
}
