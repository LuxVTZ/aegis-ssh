package ssh

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/sftp"
)

// ReadFile reads the contents of a file via SFTP
func (c *Client) ReadFile(path string) (string, error) {
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return "", fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer sftpClient.Close()

	file, err := sftpClient.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %v", path, err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", path, err)
	}

	return string(content), nil
}

// UploadFile uploads content to a remote path
func (c *Client) UploadFile(path, content string) error {
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer sftpClient.Close()

	file, err := sftpClient.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("failed to open file for writing %s: %v", path, err)
	}
	defer file.Close()

	_, err = file.Write([]byte(content))
	return err
}

// ListFiles lists directory contents
func (c *Client) ListFiles(dirPath string) ([]string, error) {
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return nil, fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer sftpClient.Close()

	files, err := sftpClient.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory %s: %v", dirPath, err)
	}

	var results []string
	for _, f := range files {
		results = append(results, fmt.Sprintf("%s (IsDir: %v, Size: %d)", f.Name(), f.IsDir(), f.Size()))
	}
	return results, nil
}
