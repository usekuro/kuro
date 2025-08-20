package web

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/usekuro/usekuro/internal/config"
	"github.com/usekuro/usekuro/internal/loader"
	"github.com/usekuro/usekuro/internal/persistence"
	"github.com/usekuro/usekuro/internal/runtime"
	"github.com/usekuro/usekuro/internal/schema"
)

// MockService represents a running mock service instance
type MockService struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Protocol    string                 `json:"protocol"`
	Port        int                    `json:"port"`
	Description string                 `json:"description"`
	Running     bool                   `json:"running"`
	LastStarted time.Time              `json:"lastStarted,omitempty"`
	Definition  *schema.MockDefinition `json:"-"`
	Handler     interface{}            `json:"-"`
	Context     context.Context        `json:"-"`
	Cancel      context.CancelFunc     `json:"-"`
}

// Server manages the web interface and API endpoints for UseKuro
type Server struct {
	router        *mux.Router
	serverMutex   sync.Mutex
	serverRunning bool
	mocks         map[string]*MockService
	mocksMutex    sync.RWMutex
	mockFiles     map[string]string // mockID -> file path
	autoConfig    *config.AutoConfig
	mockStore     *persistence.MockStore
}

// NewServer creates and configures a new web server instance
func NewServer() *Server {
	autoConfig, err := config.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize auto-config: %v", err)
	}

	mockStore := persistence.NewMockStore(".", "user_data", "workspaces")

	s := &Server{
		router:        mux.NewRouter(),
		serverRunning: false,
		mocks:         make(map[string]*MockService),
		mockFiles:     make(map[string]string),
		autoConfig:    autoConfig,
		mockStore:     mockStore,
	}

	config.PrintConnectionInfo(autoConfig)
	s.routes()
	return s
}

// Start initializes example mocks and starts the web server on specified port
func (s *Server) Start(port int) error {
	s.loadExampleMocks()

	addr := fmt.Sprintf(":%d", port)
	log.Printf("🌐 Web interface starting on http://localhost%s", addr)
	return http.ListenAndServe(addr, s.router)
}

// loadExampleMocks loads mock definitions from examples and mocks directories
func (s *Server) loadExampleMocks() {
	s.mocksMutex.Lock()
	defer s.mocksMutex.Unlock()

	s.loadMocksFromDirectory("examples")
	s.loadMocksFromDirectory("mocks")
}

// loadMocksFromDirectory scans a directory for .kuro files and loads them as mock services
func (s *Server) loadMocksFromDirectory(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Warning: Could not read directory %s: %v", dir, err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, ".kuro") {
			continue
		}

		filePath := filepath.Join(dir, filename)
		mockDef, err := loader.LoadMockFromFile(filePath)
		if err != nil {
			log.Printf("Warning: Could not load mock from %s: %v", filePath, err)
			continue
		}

		mockID := strings.TrimSuffix(filename, ".kuro")
		if dir != "examples" {
			mockID = fmt.Sprintf("%s_%s", dir, mockID)
		}

		name := mockDef.Meta.Name
		if name == "" {
			name = strings.TrimSuffix(filename, ".kuro")
		}

		description := mockDef.Meta.Description
		if description == "" {
			description = fmt.Sprintf("Mock service from %s", filePath)
		}

		s.mocks[mockID] = &MockService{
			ID:          mockID,
			Name:        name,
			Protocol:    mockDef.Protocol,
			Port:        mockDef.Port,
			Description: description,
			Running:     false,
			Definition:  mockDef,
		}
		s.mockFiles[mockID] = filePath

		log.Printf("Loaded mock: %s (%s:%d)", name, mockDef.Protocol, mockDef.Port)
	}
}

func (s *Server) handleGetMocks(w http.ResponseWriter, r *http.Request) {
	s.mocksMutex.RLock()
	defer s.mocksMutex.RUnlock()

	mocks := make([]*MockService, 0, len(s.mocks))
	for _, m := range s.mocks {
		mocks = append(mocks, m)
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"mocks":        mocks,
		"serverStatus": s.serverRunning,
	})
}

func (s *Server) handleAddMock(w http.ResponseWriter, r *http.Request) {
	var mock MockService
	if err := json.NewDecoder(r.Body).Decode(&mock); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if mock.Name == "" || mock.Port <= 0 || mock.Port > 65535 {
		respondWithError(w, http.StatusBadRequest, "Invalid mock configuration")
		return
	}

	mock.ID = fmt.Sprintf("custom_%d", time.Now().Unix())
	mock.Running = false

	s.mocksMutex.Lock()
	s.mocks[mock.ID] = &mock
	s.mocksMutex.Unlock()

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"mock":    mock,
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.serverMutex.Lock()
	serverRunning := s.serverRunning
	s.serverMutex.Unlock()

	s.mocksMutex.RLock()
	runningMocks := 0
	totalMocks := len(s.mocks)
	for _, mock := range s.mocks {
		if mock.Running {
			runningMocks++
		}
	}
	s.mocksMutex.RUnlock()

	status := "healthy"
	if totalMocks == 0 {
		status = "no_mocks"
	} else if runningMocks == 0 {
		status = "all_stopped"
	}

	response := map[string]interface{}{
		"status":        status,
		"timestamp":     time.Now().Format(time.RFC3339),
		"server":        serverRunning,
		"mocks_total":   totalMocks,
		"mocks_running": runningMocks,
		"version":       "1.0.0",
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (s *Server) handleToggleMock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mockID := vars["id"]

	s.mocksMutex.Lock()
	defer s.mocksMutex.Unlock()

	mock, exists := s.mocks[mockID]
	if !exists {
		respondWithError(w, http.StatusNotFound, "Mock not found")
		return
	}

	if mock.Running {
		// Stop the mock
		if mock.Cancel != nil {
			mock.Cancel()
		}

		if mock.Handler != nil {
			switch handler := mock.Handler.(type) {
			case *runtime.HTTPHandler:
				handler.Stop()
			case *runtime.TCPHandler:
				handler.Stop()
			case *runtime.WSHandler:
				handler.Stop()
			case *runtime.SFTPHandler:
				handler.Stop()
			}
		}

		mock.Running = false
		mock.Handler = nil
		mock.Context = nil
		mock.Cancel = nil
	} else {
		// Start the mock
		if mock.Definition == nil {
			respondWithError(w, http.StatusInternalServerError, "Mock definition not found")
			return
		}

		// Check for port conflicts
		for _, existingMock := range s.mocks {
			if existingMock.Running && existingMock.Port == mock.Port && existingMock.ID != mock.ID {
				respondWithError(w, http.StatusConflict, fmt.Sprintf("Port %d is already in use by mock '%s'", mock.Port, existingMock.Name))
				return
			}
		}

		// Create context for the mock
		ctx, cancel := context.WithCancel(context.Background())
		mock.Context = ctx
		mock.Cancel = cancel

		// Create and start the appropriate handler
		var handler runtime.ProtocolHandler

		switch strings.ToLower(mock.Protocol) {
		case "http", "https":
			handler = runtime.NewHTTPHandler()
		case "tcp":
			handler = runtime.NewTCPHandler()
		case "ws", "websocket":
			handler = runtime.NewWSHandler()
		case "sftp":
			handler = runtime.NewSFTPHandler()
		default:
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Unsupported protocol: %s", mock.Protocol))
			return
		}

		if err := handler.Start(mock.Definition); err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to start mock: %v", err))
			return
		}

		mock.Handler = handler
		mock.Running = true
		mock.LastStarted = time.Now()
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"running": mock.Running,
		"mock":    mock,
	})
}

func (s *Server) handleToggleServer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Action string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	s.serverMutex.Lock()
	defer s.serverMutex.Unlock()

	switch strings.ToLower(req.Action) {
	case "start":
		s.serverRunning = true
	case "stop":
		s.serverRunning = false
		s.mocksMutex.Lock()
		for _, m := range s.mocks {
			if m.Running {
				// Stop the handler properly
				if m.Handler != nil {
					switch handler := m.Handler.(type) {
					case *runtime.HTTPHandler:
						handler.Stop()
					case *runtime.TCPHandler:
						handler.Stop()
					case *runtime.WSHandler:
						handler.Stop()
					case *runtime.SFTPHandler:
						handler.Stop()
					}
				}
				if m.Cancel != nil {
					m.Cancel()
				}
				m.Running = false
				m.Handler = nil
				m.Context = nil
				m.Cancel = nil
			}
		}
		s.mocksMutex.Unlock()
	default:
		respondWithError(w, http.StatusBadRequest, "Invalid action")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"running": s.serverRunning,
	})
}

func (s *Server) handleUpdateMock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mockID := vars["id"]

	var req struct {
		Name        *string `json:"name"`
		Protocol    *string `json:"protocol"`
		Port        *int    `json:"port"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Protocol != nil {
		p := strings.ToLower(strings.TrimSpace(*req.Protocol))
		if p == "" {
			respondWithError(w, http.StatusBadRequest, "Invalid protocol")
			return
		}
		*req.Protocol = p
	}

	// Update mock
	s.mocksMutex.Lock()
	defer s.mocksMutex.Unlock()

	mock, ok := s.mocks[mockID]
	if !ok {
		respondWithError(w, http.StatusNotFound, "Mock not found")
		return
	}

	if req.Name != nil {
		mock.Name = strings.TrimSpace(*req.Name)
	}
	if req.Protocol != nil {
		mock.Protocol = *req.Protocol
	}
	if req.Port != nil {
		mock.Port = *req.Port
	}
	if req.Description != nil {
		mock.Description = *req.Description
	}

	// Save to persistent storage
	savedMock := &persistence.SavedMock{
		ID:          mock.ID,
		Name:        mock.Name,
		Protocol:    mock.Protocol,
		Port:        mock.Port,
		Description: mock.Description,
		UserID:      "default",
		Definition:  mock.Definition,
		Source:      "frontend",
	}

	if err := s.mockStore.SaveMock(savedMock); err != nil {
		log.Printf("Warning: Failed to save mock to persistent storage: %v", err)
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"mock":    mock,
	})
}

// handlePublicConfig exposes server configuration and connection details
func (s *Server) handlePublicConfig(w http.ResponseWriter, r *http.Request) {
	config, err := s.autoConfig.GetPublicConfig()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get public configuration")
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	respondWithJSON(w, http.StatusOK, config)
}

// handleUserMocks returns all mocks for a specific user workspace
func (s *Server) handleUserMocks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	if userID == "" {
		userID = "default"
	}

	// Ensure workspace exists
	if err := s.mockStore.CreateUserWorkspace(userID); err != nil {
		log.Printf("Warning: Failed to ensure workspace exists for %s: %v", userID, err)
	}

	mocks, err := s.mockStore.ListUserMocks(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to list user mocks")
		return
	}

	// Always get example mocks from default workspace for reference
	exampleMocks := make([]*persistence.MockMetadata, 0)
	if userID != "default" {
		defaultMocks, err := s.mockStore.ListUserMocks("default")
		if err == nil {
			for _, mock := range defaultMocks {
				if !strings.HasPrefix(mock.ID, "user_") {
					mock.Source = "example"
					exampleMocks = append(exampleMocks, mock)
				}
			}
		}
	}

	// Separate user mocks from example mocks in current workspace
	userMocks := make([]*persistence.MockMetadata, 0)

	for _, mock := range mocks {
		if strings.HasPrefix(mock.ID, "user_") {
			mock.Source = "user"
			userMocks = append(userMocks, mock)
		} else if userID == "default" {
			// In default workspace, non-user mocks are examples
			mock.Source = "example"
			exampleMocks = append(exampleMocks, mock)
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"mocks":             userMocks,    // Only user-created mocks
		"example_mocks":     exampleMocks, // Example mocks (read-only)
		"user_id":           userID,
		"total_user":        len(userMocks),
		"total_examples":    len(exampleMocks),
		"current_workspace": userID,
		"workspace_type": func() string {
			if userID == "default" {
				return "default"
			}
			return "user"
		}(),
	})
}

// handleCreateUserMock creates a new mock in user's workspace
func (s *Server) handleCreateUserMock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	if userID == "" {
		userID = "default"
	}

	var req struct {
		Name        string `json:"name"`
		Protocol    string `json:"protocol"`
		Port        int    `json:"port"`
		Description string `json:"description"`
		Content     string `json:"content,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Name == "" || req.Port <= 0 || req.Port > 65535 {
		respondWithError(w, http.StatusBadRequest, "Invalid mock configuration")
		return
	}

	// Check for port conflicts with running mocks
	s.mocksMutex.Lock()
	for _, existingMock := range s.mocks {
		if existingMock.Running && existingMock.Port == req.Port {
			s.mocksMutex.Unlock()
			respondWithError(w, http.StatusConflict, fmt.Sprintf("Port %d is already in use by mock '%s'", req.Port, existingMock.Name))
			return
		}
	}
	s.mocksMutex.Unlock()

	// Create mock definition automatically - this fixes the "Mock definition not found" error
	definition := s.generateMockDefinition(req.Protocol, req.Port, req.Name, req.Description)

	// Generate unique ID with timestamp
	mockID := fmt.Sprintf("user_%s_%d", userID, time.Now().UnixNano())

	mock := &persistence.SavedMock{
		ID:          mockID,
		Name:        req.Name,
		Protocol:    req.Protocol,
		Port:        req.Port,
		Description: req.Description,
		Definition:  definition,
		Content:     req.Content,
		UserID:      userID,
		Source:      "user",
	}

	// Save to persistent storage first
	if err := s.mockStore.SaveMock(mock); err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to save mock: %v", err))
		return
	}

	// Add to runtime mocks for immediate use (this allows starting without restart)
	runtimeMock := &MockService{
		ID:          mock.ID,
		Name:        mock.Name,
		Protocol:    mock.Protocol,
		Port:        mock.Port,
		Description: mock.Description,
		Running:     false,
		Definition:  definition,
	}

	s.mocksMutex.Lock()
	s.mocks[mock.ID] = runtimeMock
	s.mocksMutex.Unlock()

	log.Printf("Created user mock: %s (ID: %s) in workspace: %s", mock.Name, mock.ID, userID)

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"mock":    mock,
		"message": "Mock created successfully with auto-generated definition",
	})
}

// handleUserMock handles GET, PUT, DELETE operations for individual user mocks
func (s *Server) handleUserMock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]
	mockID := vars["mockID"]

	if userID == "" {
		userID = "default"
	}

	switch r.Method {
	case "GET":
		mock, err := s.mockStore.LoadMock(userID, mockID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "Mock not found")
			return
		}

		respondWithJSON(w, http.StatusOK, mock)

	case "PUT":
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		if err := s.mockStore.UpdateMock(userID, mockID, updates); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to update mock")
			return
		}

		// Update runtime mock if it exists
		s.mocksMutex.Lock()
		if runtimeMock, exists := s.mocks[mockID]; exists {
			if name, ok := updates["name"].(string); ok {
				runtimeMock.Name = name
			}
			if protocol, ok := updates["protocol"].(string); ok {
				runtimeMock.Protocol = protocol
			}
			if port, ok := updates["port"].(float64); ok {
				runtimeMock.Port = int(port)
			}
			if description, ok := updates["description"].(string); ok {
				runtimeMock.Description = description
			}
		}
		s.mocksMutex.Unlock()

		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Mock updated successfully",
		})

	case "DELETE":
		// Only allow deletion of user-created mocks (with user_ prefix)
		if !strings.HasPrefix(mockID, "user_"+userID+"_") {
			respondWithError(w, http.StatusForbidden, "Cannot delete example mocks. Only user-created mocks can be deleted.")
			return
		}

		// Stop runtime mock if running
		s.mocksMutex.Lock()
		if runtimeMock, exists := s.mocks[mockID]; exists && runtimeMock.Running {
			if runtimeMock.Cancel != nil {
				runtimeMock.Cancel()
			}
			if runtimeMock.Handler != nil {
				switch handler := runtimeMock.Handler.(type) {
				case *runtime.HTTPHandler:
					handler.Stop()
				case *runtime.TCPHandler:
					handler.Stop()
				case *runtime.WSHandler:
					handler.Stop()
				case *runtime.SFTPHandler:
					handler.Stop()
				}
			}
		}
		delete(s.mocks, mockID)
		s.mocksMutex.Unlock()

		// Delete from storage
		if err := s.mockStore.DeleteMock(userID, mockID); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to delete mock")
			return
		}

		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "User mock deleted successfully",
		})
	}
}

// handleExportMock exports a user mock as downloadable .kuro file
func (s *Server) handleExportMock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]
	mockID := vars["mockID"]

	if userID == "" {
		userID = "default"
	}

	content, err := s.mockStore.ExportMock(userID, mockID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Mock not found")
		return
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.kuro", mockID))
	w.Write(content)
}

// handleImportMock imports a .kuro file into user's workspace
func (s *Server) handleImportMock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	if userID == "" {
		userID = "default"
	}

	var req struct {
		Content  string                 `json:"content"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	mock, err := s.mockStore.ImportMock(userID, []byte(req.Content), req.Metadata)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to import mock")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"mock":    mock,
	})
}

// handleUserStats returns usage statistics for a user's workspace
func (s *Server) handleUserStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	if userID == "" {
		userID = "default"
	}

	stats, err := s.mockStore.GetUserStats(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get user stats")
		return
	}

	respondWithJSON(w, http.StatusOK, stats)
}

// routes configures all HTTP routes for the web server
func (s *Server) routes() {
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")
	s.router.HandleFunc("/healthz", s.handleHealth).Methods("GET")

	s.router.HandleFunc("/api/config", s.handlePublicConfig).Methods("GET")

	// Workspace management endpoints
	s.router.HandleFunc("/api/workspaces", s.handleListWorkspaces).Methods("GET")
	s.router.HandleFunc("/api/user/{userID}/workspace", s.handleCreateWorkspace).Methods("POST")
	s.router.HandleFunc("/api/user/{userID}/workspace", s.handleDeleteWorkspace).Methods("DELETE")

	// User mock management endpoints
	s.router.HandleFunc("/api/user/{userID}/mocks", s.handleUserMocks).Methods("GET")
	s.router.HandleFunc("/api/user/{userID}/mocks", s.handleCreateUserMock).Methods("POST")
	s.router.HandleFunc("/api/user/{userID}/mocks/{mockID}", s.handleUserMock).Methods("GET", "PUT", "DELETE")
	s.router.HandleFunc("/api/user/{userID}/mocks/{mockID}/export", s.handleExportMock).Methods("GET")
	s.router.HandleFunc("/api/user/{userID}/import", s.handleImportMock).Methods("POST")
	s.router.HandleFunc("/api/user/{userID}/stats", s.handleUserStats).Methods("GET")

	staticRoot := http.Dir("web/static")
	staticFS := http.FileServer(staticRoot)

	staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ext := filepath.Ext(r.URL.Path)
		if ext != "" {
			w.Header().Set("Content-Type", getContentType(ext))
			w.Header().Set("Cache-Control", "public, max-age=31536000")
		}
		staticFS.ServeHTTP(w, r)
	})

	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticHandler))

	api := s.router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/mocks", s.handleGetMocks).Methods("GET")
	api.HandleFunc("/mocks", s.handleAddMock).Methods("POST")
	api.HandleFunc("/mocks/{id}/toggle", s.handleToggleMock).Methods("POST")
	api.HandleFunc("/server/toggle", s.handleToggleServer).Methods("POST")
	api.HandleFunc("/mocks/{id}", s.handleUpdateMock).Methods("PUT")

	s.router.HandleFunc("/", s.handleIndex).Methods("GET")
}

// handleIndex serves the main web interface
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	indexPath := "web/index.html"
	http.ServeFile(w, r, indexPath)
}

// getContentType returns the appropriate MIME type for file extensions
func getContentType(ext string) string {
	switch ext {
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".html":
		return "text/html"
	default:
		return "application/octet-stream"
	}
}

// respondWithError sends a JSON error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON sends a JSON response with the given status code
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// generateMockDefinition creates a basic mock definition based on protocol
func (s *Server) generateMockDefinition(protocol string, port int, name, description string) *schema.MockDefinition {
	definition := &schema.MockDefinition{
		Protocol: strings.ToLower(protocol),
		Port:     port,
		Meta: schema.Meta{
			Name:        name,
			Description: description,
		},
	}

	switch strings.ToLower(protocol) {
	case "http", "https":
		definition.Routes = []schema.Route{
			{
				Path:   "/",
				Method: "GET",
				Response: schema.ResponseDefinition{
					Status: 200,
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: `{"message": "Hello from ` + name + `", "status": "ok"}`,
				},
			},
		}
	case "sftp":
		definition.SFTPAuth = &schema.SFTPAuth{
			Username: "usekuro",
			Password: "kuro123",
		}
		definition.Files = []schema.FileEntry{
			{
				Path:    "/welcome.txt",
				Content: "Welcome to " + name + " SFTP server!",
			},
			{
				Path:    "/data/sample.json",
				Content: `{"server": "` + name + `", "protocol": "sftp"}`,
			},
		}
	case "tcp":
		definition.OnMessage = &schema.OnMessage{
			Match: ".*",
			Conditions: []schema.OnMessageRule{
				{
					If:      "HELLO",
					Respond: "HELLO from " + name,
				},
			},
			Else: "Echo: {{message}}",
		}
	case "websocket", "ws":
		definition.OnMessage = &schema.OnMessage{
			Match: ".*",
			Conditions: []schema.OnMessageRule{
				{
					If:      "ping",
					Respond: "pong from " + name,
				},
			},
			Else: "Echo: {{message}}",
		}
	}

	return definition
}

// handleListWorkspaces returns a list of available workspaces
func (s *Server) handleListWorkspaces(w http.ResponseWriter, r *http.Request) {
	workspaces := []map[string]interface{}{
		{
			"id":          "default",
			"name":        "Default Workspace",
			"description": "Default workspace with example mocks",
			"protected":   true,
			"created_at":  "2024-01-01T00:00:00Z",
			"is_default":  true,
		},
	}

	// Get user workspaces from filesystem
	entries, err := os.ReadDir(s.mockStore.WorkspacePath)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() != "default" {
				configPath := filepath.Join(s.mockStore.WorkspacePath, entry.Name(), "config.json")
				workspace := map[string]interface{}{
					"id":          entry.Name(),
					"name":        fmt.Sprintf("Workspace %s", entry.Name()),
					"description": "User workspace",
					"protected":   false,
					"is_default":  false,
				}

				// Try to read config for more details
				if configData, err := os.ReadFile(configPath); err == nil {
					var config map[string]interface{}
					if json.Unmarshal(configData, &config) == nil {
						if createdAt, ok := config["created_at"].(string); ok {
							workspace["created_at"] = createdAt
						}
						if name, ok := config["display_name"].(string); ok && name != "" {
							workspace["name"] = name
						}
					}
				} else {
					workspace["created_at"] = time.Now().Format(time.RFC3339)
				}

				workspaces = append(workspaces, workspace)
			}
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"workspaces": workspaces,
		"current":    "default", // This could be dynamic based on user session
	})
}

// handleCreateWorkspace creates a new user workspace
func (s *Server) handleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	if userID == "" || userID == "default" {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID for workspace creation")
		return
	}

	// Validate workspace name (basic validation)
	if len(userID) < 2 || len(userID) > 50 {
		respondWithError(w, http.StatusBadRequest, "Workspace name must be between 2 and 50 characters")
		return
	}

	// Check if workspace already exists
	workspacePath := filepath.Join(s.mockStore.WorkspacePath, userID)
	if _, err := os.Stat(workspacePath); err == nil {
		respondWithError(w, http.StatusConflict, "Workspace already exists")
		return
	}

	// Create the workspace
	if err := s.mockStore.CreateUserWorkspace(userID); err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create workspace: %v", err))
		return
	}

	log.Printf("Created new workspace: %s", userID)

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Workspace created successfully",
		"workspace": map[string]interface{}{
			"id":          userID,
			"name":        fmt.Sprintf("Workspace %s", userID),
			"description": "User workspace",
			"protected":   false,
			"created_at":  time.Now().Format(time.RFC3339),
			"is_default":  false,
		},
	})
}

// handleDeleteWorkspace deletes a user workspace and all its content
func (s *Server) handleDeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	if userID == "" || userID == "default" {
		respondWithError(w, http.StatusBadRequest, "Cannot delete default workspace")
		return
	}

	workspacePath := filepath.Join(s.mockStore.WorkspacePath, userID)

	// Check if workspace exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		respondWithError(w, http.StatusNotFound, "Workspace not found")
		return
	}

	// Stop and remove ONLY user-created mocks from this workspace
	s.mocksMutex.Lock()
	var userMocksToDelete []string
	for mockID, mock := range s.mocks {
		// Only delete mocks that were created by this specific user
		if strings.HasPrefix(mockID, "user_"+userID+"_") {
			userMocksToDelete = append(userMocksToDelete, mockID)
			if mock.Running {
				if mock.Cancel != nil {
					mock.Cancel()
				}
				if mock.Handler != nil {
					switch handler := mock.Handler.(type) {
					case *runtime.HTTPHandler:
						handler.Stop()
					case *runtime.TCPHandler:
						handler.Stop()
					case *runtime.WSHandler:
						handler.Stop()
					case *runtime.SFTPHandler:
						handler.Stop()
					}
				}
			}
		}
	}

	// Delete user mocks from runtime after stopping them
	for _, mockID := range userMocksToDelete {
		delete(s.mocks, mockID)
		log.Printf("Removed user mock from runtime: %s", mockID)
	}
	s.mocksMutex.Unlock()

	// Remove workspace directory completely
	if err := os.RemoveAll(workspacePath); err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete workspace: %v", err))
		return
	}

	log.Printf("Deleted workspace: %s (removed %d user mocks)", userID, len(userMocksToDelete))

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success":       true,
		"message":       fmt.Sprintf("Workspace '%s' deleted successfully. Example mocks preserved.", userID),
		"deleted_mocks": len(userMocksToDelete),
	})
}
