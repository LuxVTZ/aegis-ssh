package monitoring

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/LuxVTZ/aegis-ssh/internal/config"
	"github.com/LuxVTZ/aegis-ssh/internal/ssh"
	"github.com/mark3labs/mcp-go/server"
)

func StartDaemon(s *server.MCPServer, cfgManager *config.Manager) {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			checkServers(s, cfgManager)
		}
	}()
}

func checkServers(s *server.MCPServer, cfgManager *config.Manager) {
	machines := cfgManager.ListMachines()
	for _, m := range machines {
		client, err := ssh.NewClient(m)
		if err != nil {
			notify(s, fmt.Sprintf("Alert: Server %s is unreachable: %v", m.Name, err))
			continue
		}
		
		// Check CPU load
		stdout, _, _, _ := client.ExecuteCommand("uptime | awk -F'load average:' '{ print $2 }' | cut -d, -f1")
		load := strings.TrimSpace(stdout)
		if load != "" {
			notify(s, fmt.Sprintf("Metric: Server %s load is %s", m.Name, load))
		}
		
		client.Close()
	}
}

func notify(s *server.MCPServer, message string) {
	log.Println("[DAEMON ALERT]", message)
}
