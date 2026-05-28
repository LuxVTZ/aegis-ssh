package ws

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/LuxVTZ/aegis-ssh/internal/config"
	"github.com/LuxVTZ/aegis-ssh/internal/history"
	"github.com/LuxVTZ/aegis-ssh/internal/ssh"
	"github.com/gorilla/websocket"
)

//go:embed web/*
var webFiles embed.FS

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all for now, in prod should restrict
	},
}

type WsMessage struct {
	Type   string `json:"type"`
	Data   string `json:"data,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

func StartServer(port int, cfgManager *config.Manager) error {
	addr := fmt.Sprintf(":%d", port)

	// API Routes
	http.HandleFunc("/api/servers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		machines := cfgManager.ListMachines()
		if machines == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(machines)
	})

	http.HandleFunc("/api/audit", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if history.GlobalManager != nil {
			logs, _ := history.GlobalManager.GetRecentLogs(50)
			json.NewEncoder(w).Encode(logs)
		} else {
			json.NewEncoder(w).Encode([]history.AuditLog{})
		}
	})

	// WebSockets
	http.HandleFunc("/ws/shell/", func(w http.ResponseWriter, r *http.Request) {
		host := strings.TrimPrefix(r.URL.Path, "/ws/shell/")
		if host == "" {
			http.Error(w, "Host required", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Upgrade error: %v", err)
			return
		}
		defer conn.Close()

		handleShell(conn, host, cfgManager)
	})

	// Serve Embedded Frontend
	webRoot, _ := fs.Sub(webFiles, "web")
	http.Handle("/", http.FileServer(http.FS(webRoot)))

	log.Printf("Aegis Dashboard & WS starting on http://localhost:%d", port)
	return http.ListenAndServe(addr, nil)
}

func handleShell(conn *websocket.Conn, host string, cfgManager *config.Manager) {
	machine, err := cfgManager.GetMachine(host)
	if err != nil {
		conn.WriteJSON(map[string]string{"type": "error", "message": "Unknown host: " + host})
		return
	}

	client, err := ssh.NewClient(machine)
	if err != nil {
		conn.WriteJSON(map[string]string{"type": "error", "message": "SSH Connection failed: " + err.Error()})
		return
	}
	defer client.Close()

	session, err := ssh.GlobalPTYManager.CreateSession(client)
	if err != nil {
		conn.WriteJSON(map[string]string{"type": "error", "message": "PTY allocation failed: " + err.Error()})
		return
	}
	defer ssh.GlobalPTYManager.CloseSession(session.ID)

	done := make(chan bool)

	// Reader goroutine: SSH Stdout -> WebSocket
	go func() {
		for {
			data := session.Read()
			if len(data) > 0 {
				err := conn.WriteJSON(map[string]string{"type": "output", "data": data})
				if err != nil {
					break
				}
			}
		}
		done <- true
	}()

	// Writer loop: WebSocket -> SSH Stdin
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg WsMessage
		if err := json.Unmarshal(p, &msg); err == nil {
			if msg.Type == "input" {
				session.Write(msg.Data)
			} else if msg.Type == "close" {
				break
			}
			// Note: resizing PTY is possible via x/crypto/ssh but requires passing WindowChangeRequest
		}
	}

	conn.WriteJSON(map[string]string{"type": "closed"})
}
