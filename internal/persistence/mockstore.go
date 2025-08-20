package persistence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/usekuro/usekuro/internal/schema"
)

// MockStore handles persistence of user-created mock configurations
type MockStore struct {
	BasePath      string
	UserDataPath  string
	WorkspacePath string
}

// SavedMock represents a mock configuration saved by a user
type SavedMock struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Protocol    string                 `json:"protocol"`
	Port        int                    `json:"port"`
	Description string                 `json:"description"`
	Definition  *schema.MockDefinition `json:"definition,omitempty"`
	Content     string                 `json:"content,omitempty"`
	UserID      string                 `json:"user_id"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Source      string                 `json:"source"` // "frontend", "file", "import"
	FilePath    string                 `json:"file_path,omitempty"`
}

// MockMetadata contains summary information about a saved mock
type MockMetadata struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Protocol    string    `json:"protocol"`
	Port        int       `json:"port"`
	Description string    `json:"description"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Source      string    `json:"source"`
	HasContent  bool      `json:"has_content"`
}

// NewMockStore creates a new mock store instance
func NewMockStore(basePath, userDataPath, workspacePath string) *MockStore {
	return &MockStore{
		BasePath:      basePath,
		UserDataPath:  userDataPath,
		WorkspacePath: workspacePath,
	}
}

// SaveMock persists a mock configuration to the user's workspace
func (ms *MockStore) SaveMock(mock *SavedMock) error {
	if mock.UserID == "" {
		mock.UserID = "default"
	}

	// Ensure user directory exists
	userDir := filepath.Join(ms.WorkspacePath, mock.UserID, "mocks")
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("failed to create user directory: %w", err)
	}

	mock.UpdatedAt = time.Now()
	if mock.CreatedAt.IsZero() {
		mock.CreatedAt = mock.UpdatedAt
	}

	// Save mock definition as .kuro file
	kuruPath := filepath.Join(userDir, mock.ID+".kuro")
	if mock.Content != "" {
		if err := os.WriteFile(kuruPath, []byte(mock.Content), 0644); err != nil {
			return fmt.Errorf("failed to save mock content: %w", err)
		}
		mock.FilePath = kuruPath
	} else if mock.Definition != nil {
		// Convert definition to YAML content
		content, err := ms.definitionToYAML(mock.Definition)
		if err != nil {
			return fmt.Errorf("failed to convert definition to YAML: %w", err)
		}
		if err := os.WriteFile(kuruPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to save mock file: %w", err)
		}
		mock.FilePath = kuruPath
	}

	// Save metadata
	metadataPath := filepath.Join(userDir, mock.ID+".meta.json")
	metadataJSON, err := json.MarshalIndent(mock, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal mock metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		return fmt.Errorf("failed to save mock metadata: %w", err)
	}

	return nil
}

// LoadMock retrieves a saved mock configuration from user's workspace
func (ms *MockStore) LoadMock(userID, mockID string) (*SavedMock, error) {
	if userID == "" {
		userID = "default"
	}

	metadataPath := filepath.Join(ms.WorkspacePath, userID, "mocks", mockID+".meta.json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mock metadata: %w", err)
	}

	var mock SavedMock
	if err := json.Unmarshal(data, &mock); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mock metadata: %w", err)
	}

	// Load content if file exists
	kuruPath := filepath.Join(ms.WorkspacePath, userID, "mocks", mockID+".kuro")
	if content, err := os.ReadFile(kuruPath); err == nil {
		mock.Content = string(content)
	}

	return &mock, nil
}

// DeleteMock removes a mock configuration from user's workspace
func (ms *MockStore) DeleteMock(userID, mockID string) error {
	if userID == "" {
		userID = "default"
	}

	userDir := filepath.Join(ms.WorkspacePath, userID, "mocks")

	// Delete .kuro file
	kuruPath := filepath.Join(userDir, mockID+".kuro")
	if err := os.Remove(kuruPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete mock file: %w", err)
	}

	// Delete metadata file
	metadataPath := filepath.Join(userDir, mockID+".meta.json")
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete mock metadata: %w", err)
	}

	return nil
}

// ListUserMocks returns metadata for all mocks in a user's workspace
func (ms *MockStore) ListUserMocks(userID string) ([]*MockMetadata, error) {
	if userID == "" {
		userID = "default"
	}

	userDir := filepath.Join(ms.WorkspacePath, userID, "mocks")

	entries, err := os.ReadDir(userDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*MockMetadata{}, nil
		}
		return nil, fmt.Errorf("failed to read user mocks directory: %w", err)
	}

	var mocks []*MockMetadata
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".meta.json") {
			continue
		}

		metadataPath := filepath.Join(userDir, entry.Name())
		data, err := os.ReadFile(metadataPath)
		if err != nil {
			continue
		}

		var mock SavedMock
		if err := json.Unmarshal(data, &mock); err != nil {
			continue
		}

		metadata := &MockMetadata{
			ID:          mock.ID,
			Name:        mock.Name,
			Protocol:    mock.Protocol,
			Port:        mock.Port,
			Description: mock.Description,
			UserID:      mock.UserID,
			CreatedAt:   mock.CreatedAt,
			UpdatedAt:   mock.UpdatedAt,
			Source:      mock.Source,
			HasContent:  mock.Content != "" || mock.FilePath != "",
		}

		mocks = append(mocks, metadata)
	}

	return mocks, nil
}

// UpdateMock applies updates to an existing mock configuration
func (ms *MockStore) UpdateMock(userID, mockID string, updates map[string]interface{}) error {
	mock, err := ms.LoadMock(userID, mockID)
	if err != nil {
		return fmt.Errorf("failed to load mock for update: %w", err)
	}

	// Update fields
	if name, ok := updates["name"].(string); ok {
		mock.Name = name
	}
	if protocol, ok := updates["protocol"].(string); ok {
		mock.Protocol = protocol
	}
	if port, ok := updates["port"].(float64); ok {
		mock.Port = int(port)
	}
	if description, ok := updates["description"].(string); ok {
		mock.Description = description
	}
	if content, ok := updates["content"].(string); ok {
		mock.Content = content
	}

	return ms.SaveMock(mock)
}

// ExportMock exports a mock configuration as downloadable content
func (ms *MockStore) ExportMock(userID, mockID string) ([]byte, error) {
	mock, err := ms.LoadMock(userID, mockID)
	if err != nil {
		return nil, err
	}

	if mock.Content != "" {
		return []byte(mock.Content), nil
	}

	kuruPath := filepath.Join(ms.WorkspacePath, userID, "mocks", mockID+".kuro")
	return os.ReadFile(kuruPath)
}

// ImportMock creates a new mock from imported content
func (ms *MockStore) ImportMock(userID string, content []byte, metadata map[string]interface{}) (*SavedMock, error) {
	mock := &SavedMock{
		ID:       fmt.Sprintf("import_%d", time.Now().Unix()),
		UserID:   userID,
		Content:  string(content),
		Source:   "import",
		Protocol: "http", // default protocol
		Port:     8080,   // default port
	}

	if name, ok := metadata["name"].(string); ok && name != "" {
		mock.Name = name
	} else {
		mock.Name = "Imported Mock"
	}

	if protocol, ok := metadata["protocol"].(string); ok {
		mock.Protocol = protocol
	}

	if port, ok := metadata["port"].(float64); ok {
		mock.Port = int(port)
	}

	if description, ok := metadata["description"].(string); ok {
		mock.Description = description
	}

	return mock, ms.SaveMock(mock)
}

// CreateUserWorkspace initializes a complete workspace for a user
func (ms *MockStore) CreateUserWorkspace(userID string) error {
	if userID == "" {
		userID = "default"
	}

	userPath := filepath.Join(ms.WorkspacePath, userID)

	// Check if workspace already exists
	if _, err := os.Stat(userPath); err == nil {
		// Workspace exists, just ensure it has required structure
		return ms.ensureWorkspaceStructure(userID)
	}

	// Create workspace directories
	dirs := []string{
		userPath,
		filepath.Join(userPath, "mocks"),
		filepath.Join(userPath, "configs"),
		filepath.Join(userPath, "uploads"),
		filepath.Join(userPath, "exports"),
		filepath.Join(userPath, "custom"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create user config file with enhanced settings
	userConfig := map[string]interface{}{
		"user_id": userID,
		"display_name": func() string {
			if userID == "default" {
				return "Default Workspace"
			}
			return fmt.Sprintf("Workspace %s", userID)
		}(),
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
		"version":    "1.0",
		"workspace_type": func() string {
			if userID == "default" {
				return "default"
			}
			return "user"
		}(),
		"settings": map[string]interface{}{
			"theme":            "dark",
			"auto_save":        true,
			"auto_backup":      true,
			"default_protocol": "http",
			"default_port_range": map[string]int{
				"start": 8080,
				"end":   8999,
			},
		},
		"stats": map[string]interface{}{
			"total_mocks_created": 0,
			"last_activity":       time.Now().Format(time.RFC3339),
		},
	}

	configPath := filepath.Join(userPath, "config.json")
	configJSON, err := json.MarshalIndent(userConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, configJSON, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Create a README for user workspaces
	if userID != "default" {
		readmePath := filepath.Join(userPath, "README.md")
		readmeContent := fmt.Sprintf(`# %s

This is your personal workspace for managing custom mocks.

## Directory Structure

- mocks/ - Your custom mock definitions (.kuro files)
- configs/ - Configuration files
- uploads/ - File uploads for SFTP mocks
- exports/ - Exported mock configurations
- custom/ - Custom scripts and extensions

## Quick Start

1. Create mocks through the web interface
2. Edit .kuro files in the mocks/ directory
3. Export/import configurations as needed
4. Use example mocks from the default workspace as templates

Created: %s
`, userConfig["display_name"], time.Now().Format("2006-01-02 15:04:05"))

		os.WriteFile(readmePath, []byte(readmeContent), 0644)
	}

	return nil
}

// ensureWorkspaceStructure ensures existing workspace has all required directories
func (ms *MockStore) ensureWorkspaceStructure(userID string) error {
	userPath := filepath.Join(ms.WorkspacePath, userID)

	dirs := []string{
		filepath.Join(userPath, "mocks"),
		filepath.Join(userPath, "configs"),
		filepath.Join(userPath, "uploads"),
		filepath.Join(userPath, "exports"),
		filepath.Join(userPath, "custom"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to ensure directory %s: %w", dir, err)
		}
	}

	return nil
}

// GetUserStats calculates and returns statistics for a user's mocks
func (ms *MockStore) GetUserStats(userID string) (map[string]interface{}, error) {
	mocks, err := ms.ListUserMocks(userID)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_mocks":   len(mocks),
		"by_protocol":   make(map[string]int),
		"by_source":     make(map[string]int),
		"created_today": 0,
	}

	today := time.Now().Format("2006-01-02")
	protocolCounts := make(map[string]int)
	sourceCounts := make(map[string]int)

	for _, mock := range mocks {
		protocolCounts[mock.Protocol]++
		sourceCounts[mock.Source]++

		if mock.CreatedAt.Format("2006-01-02") == today {
			stats["created_today"] = stats["created_today"].(int) + 1
		}
	}

	stats["by_protocol"] = protocolCounts
	stats["by_source"] = sourceCounts

	return stats, nil
}

// definitionToYAML converts a mock definition to YAML format
// Note: Simplified implementation - production should use gopkg.in/yaml.v3
func (ms *MockStore) definitionToYAML(def *schema.MockDefinition) (string, error) {

	yaml := fmt.Sprintf("protocol: %s\n", def.Protocol)
	yaml += fmt.Sprintf("port: %d\n", def.Port)

	if def.Meta.Name != "" || def.Meta.Description != "" {
		yaml += "meta:\n"
		if def.Meta.Name != "" {
			yaml += fmt.Sprintf("  name: \"%s\"\n", def.Meta.Name)
		}
		if def.Meta.Description != "" {
			yaml += fmt.Sprintf("  description: \"%s\"\n", def.Meta.Description)
		}
	}

	// Add basic routes if available
	if len(def.Routes) > 0 {
		yaml += "\nroutes:\n"
		for _, route := range def.Routes {
			yaml += fmt.Sprintf("  - path: %s\n", route.Path)
			yaml += fmt.Sprintf("    method: %s\n", route.Method)
			yaml += "    response:\n"
			yaml += fmt.Sprintf("      status: %d\n", route.Response.Status)
			if len(route.Response.Headers) > 0 {
				yaml += "      headers:\n"
				for k, v := range route.Response.Headers {
					yaml += fmt.Sprintf("        %s: \"%s\"\n", k, v)
				}
			}
			if route.Response.Body != "" {
				yaml += fmt.Sprintf("      body: |\n        %s\n", route.Response.Body)
			}
		}
	}

	return yaml, nil
}
