package ssh

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type TunnelManager struct {
	mu      sync.Mutex
	tunnels map[string]*ActiveTunnel
}

var GlobalTunnelManager = &TunnelManager{
	tunnels: make(map[string]*ActiveTunnel),
}

type ActiveTunnel struct {
	ID         string
	Client     *Client
	Listener   net.Listener
	RemotePort int
	LocalPort  int
	Host       string
}

// StartRemoteForwarding exposes a local port to a remote server
func (m *TunnelManager) StartRemoteForwarding(client *Client, remotePort, localPort int) (*ActiveTunnel, error) {
	listener, err := client.client.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", remotePort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on remote port %d: %v", remotePort, err)
	}

	id := fmt.Sprintf("tunnel-%s-%d", client.machine.Name, remotePort)
	tunnel := &ActiveTunnel{
		ID:         id,
		Client:     client,
		Listener:   listener,
		RemotePort: remotePort,
		LocalPort:  localPort,
		Host:       client.machine.Name,
	}

	m.mu.Lock()
	m.tunnels[id] = tunnel
	m.mu.Unlock()

	go func() {
		for {
			remoteConn, err := listener.Accept()
			if err != nil {
				break
			}

			go func(rConn net.Conn) {
				defer rConn.Close()
				localConn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
				if err != nil {
					log.Printf("Failed to connect to local port %d: %v", localPort, err)
					return
				}
				defer localConn.Close()

				var wg sync.WaitGroup
				wg.Add(2)

				go func() {
					io.Copy(localConn, rConn)
					wg.Done()
				}()
				go func() {
					io.Copy(rConn, localConn)
					wg.Done()
				}()

				wg.Wait()
			}(remoteConn)
		}
		m.CloseTunnel(id)
	}()

	return tunnel, nil
}

func (m *TunnelManager) CloseTunnel(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if tunnel, exists := m.tunnels[id]; exists {
		tunnel.Listener.Close()
		tunnel.Client.Close()
		delete(m.tunnels, id)
	}
}
