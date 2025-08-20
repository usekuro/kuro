package bootloader

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/usekuro/usekuro/internal/extensions"
	"github.com/usekuro/usekuro/internal/loader"
	"github.com/usekuro/usekuro/internal/runtime"
)

func BootFromFolder(path string) {
	handlers := []runtime.ProtocolHandler{}

	err := filepath.Walk(path, func(file string, info fs.FileInfo, err error) error {
		if strings.HasSuffix(file, ".kuro") && !strings.Contains(file, "/functions/") {
			mock, err := loader.LoadMockFromFile(file)
			if err != nil {
				log.Printf("❌ Error loading %s: %v", file, err)
				return nil
			}

			// Register common functions from /functions folder
			funcPath := filepath.Join(path, "functions")
			if _, err := os.Stat(funcPath); err == nil {
				files, _ := os.ReadDir(funcPath)
				for _, f := range files {
					if strings.HasSuffix(f.Name(), ".kurof") {
						_, err := extensions.LoadKurof(filepath.Join(funcPath, f.Name()))
						if err == nil {
							// agregar al mock como import local
							mock.Import = append(mock.Import, filepath.Join(funcPath, f.Name()))
						}
					}
				}
			}

			// Iniciar handler
			var handler runtime.ProtocolHandler
			switch mock.Protocol {
			case "http":
				handler = runtime.NewHTTPHandler()
			case "tcp":
				handler = runtime.NewTCPHandler()
			case "ws":
				handler = runtime.NewWSHandler()
			case "sftp":
				handler = runtime.NewSFTPHandler()
			default:
				log.Printf("⚠️ Protocolo no reconocido: %s", mock.Protocol)
				return nil
			}

			if err := handler.Start(mock); err != nil {
				log.Printf("🚨 Error starting mock %s: %v", file, err)
				return nil
			}
			log.Printf("✅ Mock started: %s (%s)", file, mock.Protocol)
			handlers = append(handlers, handler)
		}
		return nil
	})

	if err != nil {
		log.Fatal("Error scanning folder:", err)
	}
}
