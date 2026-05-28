package ssh

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"math/rand"
	"sync"
	"time"
)

type PTYManager struct {
	mu       sync.Mutex
	sessions map[string]*PTYSession
}

var GlobalPTYManager = &PTYManager{
	sessions: make(map[string]*PTYSession),
}

type PTYSession struct {
	ID        string
	client    *Client
	session   *ssh.Session
	stdin     io.WriteCloser
	stdoutBuf *bytes.Buffer
	mu        sync.Mutex
	lastUsed  time.Time
}

func (m *PTYManager) CreateSession(client *Client) (*PTYSession, error) {
	session, err := client.client.NewSession()
	if err != nil {
		return nil, err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm", 24, 80, modes); err != nil {
		session.Close()
		return nil, err
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		return nil, err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, err
	}

	if err := session.Shell(); err != nil {
		session.Close()
		return nil, err
	}

	id := fmt.Sprintf("pty-%d", rand.Intn(1000000))

	ptySession := &PTYSession{
		ID:        id,
		client:    client,
		session:   session,
		stdin:     stdin,
		stdoutBuf: &bytes.Buffer{},
		lastUsed:  time.Now(),
	}

	// Read loop
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				ptySession.mu.Lock()
				ptySession.stdoutBuf.Write(buf[:n])
				ptySession.mu.Unlock()
			}
			if err != nil {
				break
			}
		}
		m.CloseSession(id)
	}()

	m.mu.Lock()
	m.sessions[id] = ptySession
	m.mu.Unlock()

	return ptySession, nil
}

func (m *PTYManager) GetSession(id string) (*PTYSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session %s not found", id)
	}
	session.lastUsed = time.Now()
	return session, nil
}

func (m *PTYManager) CloseSession(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, exists := m.sessions[id]; exists {
		session.session.Close()
		session.client.Close()
		delete(m.sessions, id)
	}
}

func (s *PTYSession) Write(data string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastUsed = time.Now()
	_, err := s.stdin.Write([]byte(data))
	return err
}

func (s *PTYSession) Read() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastUsed = time.Now()
	data := s.stdoutBuf.String()
	s.stdoutBuf.Reset()
	return data
}
