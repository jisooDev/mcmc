package sftp

import (
	"fmt"
	"io"
	"mcmc/config"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Client เป็นสำหรับจัดการการเชื่อมต่อ SFTP
type Client struct {
	sshClient  *ssh.Client
	sftpClient *sftp.Client
	config     config.SFTPConfig
}

// NewClient สร้างการเชื่อมต่อใหม่กับ SFTP server
func NewClient(cfg config.SFTPConfig) (*Client, error) {
	// สร้าง SSH client config
	sshConfig := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(cfg.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// เชื่อมต่อกับ SSH server
	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", cfg.Host, cfg.Port), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อกับ SSH server ได้: %v", err)
	}

	// สร้าง SFTP client
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("ไม่สามารถสร้าง SFTP client ได้: %v", err)
	}

	return &Client{
		sshClient:  sshClient,
		sftpClient: sftpClient,
		config:     cfg,
	}, nil
}

// Close ปิดการเชื่อมต่อทั้งหมด
func (c *Client) Close() error {
	err1 := c.sftpClient.Close()
	err2 := c.sshClient.Close()

	if err1 != nil {
		return err1
	}
	return err2
}

// FindLatestFileByPrefix หาไฟล์ล่าสุดตาม prefix
func (c *Client) FindLatestFileByPrefix(prefix string) (os.FileInfo, error) {
	files, err := c.sftpClient.ReadDir(c.config.RemotePath)
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถอ่านรายการไฟล์ได้: %v", err)
	}

	var matchedFiles []os.FileInfo
	for _, file := range files {
		if strings.HasPrefix(file.Name(), prefix) {
			matchedFiles = append(matchedFiles, file)
		}
	}

	if len(matchedFiles) == 0 {
		return nil, fmt.Errorf("ไม่พบไฟล์ที่ขึ้นต้นด้วย %s", prefix)
	}

	// เรียงไฟล์ตามวันที่แก้ไข (ใหม่สุดอยู่ข้างบน)
	sort.Slice(matchedFiles, func(i, j int) bool {
		return matchedFiles[i].ModTime().After(matchedFiles[j].ModTime())
	})

	return matchedFiles[0], nil
}

// DownloadFile ดาวน์โหลดไฟล์จาก SFTP server
func (c *Client) DownloadFile(remoteFilePath, localFilePath string) (int64, error) {
	// เปิดไฟล์บนเซิร์ฟเวอร์
	remoteFile, err := c.sftpClient.Open(remoteFilePath)
	if err != nil {
		return 0, fmt.Errorf("ไม่สามารถเปิดไฟล์บนเซิร์ฟเวอร์ได้: %v", err)
	}
	defer remoteFile.Close()

	// สร้างไฟล์บนเครื่องของเรา
	localDir := filepath.Dir(localFilePath)
	if _, err := os.Stat(localDir); os.IsNotExist(err) {
		os.MkdirAll(localDir, 0755)
	}

	localFile, err := os.Create(localFilePath)
	if err != nil {
		return 0, fmt.Errorf("ไม่สามารถสร้างไฟล์บนเครื่องได้: %v", err)
	}
	defer localFile.Close()

	// คัดลอกข้อมูลจากไฟล์บนเซิร์ฟเวอร์มายังไฟล์บนเครื่อง
	n, err := io.Copy(localFile, remoteFile)
	if err != nil {
		return 0, fmt.Errorf("ไม่สามารถคัดลอกข้อมูลได้: %v", err)
	}

	return n, nil
}