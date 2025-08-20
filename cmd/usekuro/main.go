package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/usekuro/usekuro/internal/bootloader"
	"github.com/usekuro/usekuro/internal/extensions"
	"github.com/usekuro/usekuro/internal/loader"
	runtimepkg "github.com/usekuro/usekuro/internal/runtime"
	"github.com/usekuro/usekuro/internal/template"
	"github.com/usekuro/usekuro/internal/web"
)

func init() {
	// Setup structured logging
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})

	// Set log level from environment or default to Info
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		if parsedLevel, err := logrus.ParseLevel(level); err == nil {
			logrus.SetLevel(parsedLevel)
		}
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		if len(os.Args) < 3 {
			log.Fatal("You must specify a `.kuro` file")
		}
		runMock(os.Args[2])

	case "boot":
		if len(os.Args) < 3 {
			log.Fatal("You must specify the backup folder")
		}
		bootloader.BootFromFolder(os.Args[2])
		waitForExit()

	case "validate":
		if len(os.Args) < 3 {
			log.Fatal("You must specify a `.kuro` file")
		}
		validateMock(os.Args[2])

	case "web":
		port := 3000
		if len(os.Args) >= 3 {
			if p, err := strconv.Atoi(os.Args[2]); err == nil {
				port = p
			}
		}

		server := web.NewServer()
		log.Fatal(server.Start(port))

	default:
		log.Fatalf("Unknown command: %s", os.Args[1])
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  usekuro run file.kuro          # Run a mock")
	fmt.Println("  usekuro boot folder/           # Run multiple mocks from backup folder")
	fmt.Println("  usekuro validate file.kuro     # Validate schema without running")
	fmt.Println("  usekuro web [port]             # Start web interface (default port 8798)")
}

func runMock(path string) {
	logger := logrus.WithField("component", "mock-runner")
	logger.Infof("Loading mock from file: %s", path)

	mock, err := loader.LoadMockFromFile(path)
	if err != nil {
		logger.Fatalf("Error loading file: %v", err)
	}

	logger.WithFields(logrus.Fields{
		"protocol": mock.Protocol,
		"port":     mock.Port,
		"routes":   len(mock.Routes),
	}).Info("Mock loaded successfully")

	// Register imports
	reg := extensions.NewRegistry()
	for _, src := range mock.Import {
		logger.Debugf("Loading import: %s", src)
		code, err := extensions.LoadKurof(src)
		if err == nil {
			reg.Register(src, code, src)
			logger.Infof("Import loaded: %s", src)
		} else {
			logger.Warnf("Failed to load import %s: %v", src, err)
		}
	}

	ctx := template.MergeContext(nil, nil, mock.Context.Variables)
	if _, err := template.NewRuntime(ctx, reg); err != nil {
		logger.Errorf("Template runtime initialization failed: %v", err)
	}

	var handler runtimepkg.ProtocolHandler

	switch mock.Protocol {
	case "http":
		handler = runtimepkg.NewHTTPHandler()
	case "tcp":
		handler = runtimepkg.NewTCPHandler()
	case "ws":
		handler = runtimepkg.NewWSHandler()
	case "sftp":
		handler = runtimepkg.NewSFTPHandler()
	default:
		logger.Fatalf("Unsupported protocol: %s", mock.Protocol)
	}

	logger.Info("Starting mock handler...")
	if err := handler.Start(mock); err != nil {
		logger.Fatalf("Error starting handler: %v", err)
	}

	logger.WithFields(logrus.Fields{
		"file":     path,
		"protocol": mock.Protocol,
		"port":     mock.Port,
	}).Info("✅ Mock started successfully")

	// Log available endpoints for HTTP mocks
	if mock.Protocol == "http" {
		logger.Info("Available endpoints:")
		logger.Info("  GET /health   - Health check")
		logger.Info("  GET /healthz  - Health check (alias)")
		for _, route := range mock.Routes {
			logger.Infof("  %s %s", route.Method, route.Path)
		}
	}

	waitForExit()
}

func validateMock(path string) {
	_, err := loader.LoadMockFromFile(path)
	if err != nil {
		log.Fatalf("❌ Loading error: %v", err)
	}
	log.Println("✅ Valid file:", path)
}

func waitForExit() {
	logger := logrus.WithField("component", "main")
	logger.Info("Mock is running. Press Ctrl+C to stop...")

	// Set up channel to listen for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for signal
	sig := <-sigChan
	logger.Infof("Received signal: %v", sig)
	logger.Info("Shutting down gracefully...")

	// Give a moment for cleanup
	time.Sleep(100 * time.Millisecond)
	logger.Info("✅ Mock stopped")
}
