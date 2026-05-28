package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/LuxVTZ/aegis-ssh/internal/config"
	"github.com/LuxVTZ/aegis-ssh/internal/history"
	mcp_server "github.com/LuxVTZ/aegis-ssh/internal/mcp"
	"github.com/LuxVTZ/aegis-ssh/internal/ws"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aegis-ssh",
	Short: "Aegis SSH - Secure MCP server for VPS management",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the MCP server (default mode for AI clients)",
	Run: func(cmd *cobra.Command, args []string) {
		cfgManager, err := config.NewManager("")
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		if err := history.InitGlobalManager(); err != nil {
			log.Printf("Warning: Failed to init audit DB: %v", err)
		} else {
			defer history.GlobalManager.Close()
		}

		if err := mcp_server.Start(cfgManager); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured servers",
	Run: func(cmd *cobra.Command, args []string) {
		cfgManager, err := config.NewManager("")
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		machines := cfgManager.ListMachines()
		fmt.Printf("Found %d servers:\n", len(machines))
		for _, m := range machines {
			fmt.Printf("- %s (%s@%s:%d)\n", m.Name, m.User, m.Host, m.Port)
		}
	},
}

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the WebSocket HTTP server for frontend terminals",
	Run: func(cmd *cobra.Command, args []string) {
		cfgManager, err := config.NewManager("")
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		if err := history.InitGlobalManager(); err != nil {
			log.Printf("Warning: Failed to init audit DB: %v", err)
		} else {
			defer history.GlobalManager.Close()
		}

		port, _ := cmd.Flags().GetInt("port")
		if err := ws.StartServer(port, cfgManager); err != nil {
			log.Fatalf("WebSocket server error: %v", err)
		}
	},
}

func init() {
	webCmd.Flags().IntP("port", "p", 8080, "Port to run the WebSocket server on")
}

func Execute() {
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(webCmd)
	
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
