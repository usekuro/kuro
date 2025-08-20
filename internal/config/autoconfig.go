package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

// AutoConfig manages automatic configuration setup for UseKuro server
type AutoConfig struct {
	SettingsPath     string `json:"settings_path"`
	UserDataPath     string `json:"user_data_path"`
	WorkspacesPath   string `json:"workspaces_path"`
	DefaultUser      string `json:"default_user"`
	DefaultPassword  string `json:"default_password"`
	HostKeyPath      string `json:"host_key_path"`
	HostKeyPubPath   string `json:"host_key_pub_path"`
	PublicConfigPath string `json:"public_config_path"`
	CreatedAt        string `json:"created_at"`
}

// PublicConfig contains publicly accessible configuration for client connections
type PublicConfig struct {
	SFTP struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		HostKey  string `json:"host_key_fingerprint"`
	} `json:"sftp"`
	SSH struct {
		HostKey       string `json:"host_key"`
		HostKeyPub    string `json:"host_key_pub"`
		Fingerprint   string `json:"fingerprint"`
		KnownHostLine string `json:"known_host_line"`
	} `json:"ssh"`
	Workspace struct {
		DefaultPath   string `json:"default_path"`
		UserConfigDir string `json:"user_config_dir"`
	} `json:"workspace"`
	Info struct {
		Version     string `json:"version"`
		GeneratedAt string `json:"generated_at"`
		ServerName  string `json:"server_name"`
	} `json:"info"`
}

// Initialize sets up the auto-configuration system and generates required keys/directories
func Initialize() (*AutoConfig, error) {
	config := &AutoConfig{
		SettingsPath:     "settings",
		UserDataPath:     "user_data",
		WorkspacesPath:   "workspaces",
		DefaultUser:      "usekuro",
		DefaultPassword:  "kuro123",
		HostKeyPath:      "settings/host_key",
		HostKeyPubPath:   "settings/host_key.pub",
		PublicConfigPath: "settings/public_config.json",
		CreatedAt:        time.Now().Format(time.RFC3339),
	}

	if err := config.createDirectories(); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	if err := config.generateHostKeys(); err != nil {
		return nil, fmt.Errorf("failed to generate host keys: %w", err)
	}

	if err := config.createPublicConfig(); err != nil {
		return nil, fmt.Errorf("failed to create public config: %w", err)
	}

	return config, nil
}

// createDirectories creates all necessary directories for UseKuro operation
func (c *AutoConfig) createDirectories() error {
	dirs := []string{
		c.SettingsPath,
		c.UserDataPath,
		c.WorkspacesPath,
		filepath.Join(c.UserDataPath, "configs"),
		filepath.Join(c.UserDataPath, "uploads"),
		filepath.Join(c.WorkspacesPath, "default"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// generateHostKeys creates RSA host key pair for SFTP if they don't exist
func (c *AutoConfig) generateHostKeys() error {
	if _, err := os.Stat(c.HostKeyPath); err == nil {
		fmt.Printf("Host key already exists at %s\n", c.HostKeyPath)
		return nil
	}

	fmt.Printf("Generating SSH host key pair...\n")

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	if err := os.WriteFile(c.HostKeyPath, privateKeyPEM, 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to generate public key: %w", err)
	}

	publicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)
	if err := os.WriteFile(c.HostKeyPubPath, publicKeyBytes, 0644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	fmt.Printf("✅ SSH host key pair generated successfully:\n")
	fmt.Printf("   Private key: %s\n", c.HostKeyPath)
	fmt.Printf("   Public key:  %s\n", c.HostKeyPubPath)

	return nil
}

// createPublicConfig generates the public configuration file with connection details
func (c *AutoConfig) createPublicConfig() error {
	pubKeyBytes, err := os.ReadFile(c.HostKeyPubPath)
	if err != nil {
		return fmt.Errorf("failed to read public key: %w", err)
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(pubKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	fingerprint := ssh.FingerprintSHA256(pubKey)
	knownHostLine := fmt.Sprintf("localhost ssh-rsa %s", ssh.MarshalAuthorizedKey(pubKey))

	publicConfig := PublicConfig{
		SFTP: struct {
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Username string `json:"username"`
			Password string `json:"password"`
			HostKey  string `json:"host_key_fingerprint"`
		}{
			Host:     "localhost",
			Port:     2222,
			Username: c.DefaultUser,
			Password: c.DefaultPassword,
			HostKey:  fingerprint,
		},
		SSH: struct {
			HostKey       string `json:"host_key"`
			HostKeyPub    string `json:"host_key_pub"`
			Fingerprint   string `json:"fingerprint"`
			KnownHostLine string `json:"known_host_line"`
		}{
			HostKey:       string(pubKeyBytes),
			HostKeyPub:    string(pubKeyBytes),
			Fingerprint:   fingerprint,
			KnownHostLine: knownHostLine,
		},
		Workspace: struct {
			DefaultPath   string `json:"default_path"`
			UserConfigDir string `json:"user_config_dir"`
		}{
			DefaultPath:   c.WorkspacesPath,
			UserConfigDir: c.UserDataPath,
		},
		Info: struct {
			Version     string `json:"version"`
			GeneratedAt string `json:"generated_at"`
			ServerName  string `json:"server_name"`
		}{
			Version:     "1.0.0",
			GeneratedAt: time.Now().Format(time.RFC3339),
			ServerName:  "UseKuro Development Server",
		},
	}

	configJSON, err := json.MarshalIndent(publicConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal public config: %w", err)
	}

	if err := os.WriteFile(c.PublicConfigPath, configJSON, 0644); err != nil {
		return fmt.Errorf("failed to write public config: %w", err)
	}

	fmt.Printf("✅ Public configuration created at: %s\n", c.PublicConfigPath)
	return nil
}

// GetPublicConfig loads and returns the public configuration
func (c *AutoConfig) GetPublicConfig() (*PublicConfig, error) {
	data, err := os.ReadFile(c.PublicConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public config: %w", err)
	}

	var config PublicConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal public config: %w", err)
	}

	return &config, nil
}

// CreateUserWorkspace creates a dedicated workspace directory for a user
func (c *AutoConfig) CreateUserWorkspace(userID string) (string, error) {
	userPath := filepath.Join(c.WorkspacesPath, userID)

	dirs := []string{
		userPath,
		filepath.Join(userPath, "mocks"),
		filepath.Join(userPath, "configs"),
		filepath.Join(userPath, "uploads"),
		filepath.Join(userPath, "exports"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("failed to create user directory %s: %w", dir, err)
		}
	}

	defaultConfig := map[string]interface{}{
		"user_id":    userID,
		"created_at": time.Now().Format(time.RFC3339),
		"settings": map[string]interface{}{
			"theme":       "dark",
			"auto_save":   true,
			"auto_backup": true,
		},
	}

	configPath := filepath.Join(userPath, "user.json")
	configJSON, _ := json.MarshalIndent(defaultConfig, "", "  ")
	os.WriteFile(configPath, configJSON, 0644)

	return userPath, nil
}

// SaveMockConfig saves a mock configuration to user's workspace
func (c *AutoConfig) SaveMockConfig(userID, mockID string, content []byte) error {
	userPath := filepath.Join(c.WorkspacesPath, userID, "mocks")
	if err := os.MkdirAll(userPath, 0755); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s.kuro", mockID)
	filepath := filepath.Join(userPath, filename)

	return os.WriteFile(filepath, content, 0644)
}

// LoadMockConfig loads a mock configuration from user's workspace
func (c *AutoConfig) LoadMockConfig(userID, mockID string) ([]byte, error) {
	filename := fmt.Sprintf("%s.kuro", mockID)
	filepath := filepath.Join(c.WorkspacesPath, userID, "mocks", filename)

	return os.ReadFile(filepath)
}

// ListUserMocks returns a list of mock files for a specific user
func (c *AutoConfig) ListUserMocks(userID string) ([]string, error) {
	mocksPath := filepath.Join(c.WorkspacesPath, userID, "mocks")

	entries, err := os.ReadDir(mocksPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var mocks []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".kuro" {
			mocks = append(mocks, entry.Name())
		}
	}

	return mocks, nil
}

// PrintConnectionInfo displays server connection details and credentials
func PrintConnectionInfo(config *AutoConfig) {
	publicConfig, err := config.GetPublicConfig()
	if err != nil {
		fmt.Printf("❌ Error reading public config: %v\n", err)
		return
	}

	fmt.Printf("\n🚀 UseKuro Server Ready!\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("📊 Web Interface:    http://localhost:3000\n")
	fmt.Printf("🔧 Public Config:    http://localhost:3000/api/config\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("📁 SFTP Connection Details:\n")
	fmt.Printf("   Host:     %s\n", publicConfig.SFTP.Host)
	fmt.Printf("   Port:     %d\n", publicConfig.SFTP.Port)
	fmt.Printf("   Username: %s\n", publicConfig.SFTP.Username)
	fmt.Printf("   Password: %s\n", publicConfig.SFTP.Password)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("🔑 SSH Host Key Fingerprint:\n")
	fmt.Printf("   %s\n", publicConfig.SSH.Fingerprint)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("💡 Quick SFTP Test:\n")
	fmt.Printf("   sftp -P %d %s@%s\n", publicConfig.SFTP.Port, publicConfig.SFTP.Username, publicConfig.SFTP.Host)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
}
