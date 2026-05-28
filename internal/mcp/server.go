package mcp

import (
	"github.com/LuxVTZ/aegis-ssh/internal/config"
	"github.com/LuxVTZ/aegis-ssh/internal/monitoring"
	"github.com/LuxVTZ/aegis-ssh/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

// Start starts the MCP Server via stdio
func Start(cfgManager *config.Manager) error {
	// Create MCP server
	s := server.NewMCPServer(
		"aegis-ssh",
		"1.0.0",
		server.WithLogging(),
	)

	// Register tools
	tools.RegisterCommandTools(s, cfgManager)
	tools.RegisterMassCommandTools(s, cfgManager)
	tools.RegisterServerTools(s, cfgManager)
	tools.RegisterFileTools(s, cfgManager)
	tools.RegisterProcessTools(s, cfgManager)
	tools.RegisterShellTools(s, cfgManager)
	tools.RegisterDockerTools(s, cfgManager)
	tools.RegisterTunnelTools(s, cfgManager)
	tools.RegisterPlaybookTools(s, cfgManager)

	monitoring.StartDaemon(s, cfgManager)

	// Start the stdio server
	return server.ServeStdio(s)
}
