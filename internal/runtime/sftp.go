package runtime

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	runtime2 "runtime"

	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"github.com/usekuro/usekuro/internal/schema"
	"golang.org/x/crypto/ssh"
)

type SFTPHandler struct {
	port     int
	config   *ssh.ServerConfig
	listener net.Listener
	root     string
}

// Crea una nueva instancia
func NewSFTPHandler() *SFTPHandler {
	return &SFTPHandler{}
}

// Inicia el servidor
func (h *SFTPHandler) Start(def *schema.MockDefinition) error {
	h.port = def.Port
	h.root = "sftp_root"

	// Configuración de autenticación
	h.config = &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			logrus.WithField("user", c.User()).Info("🔐 Password auth")
			return nil, nil
		},
		PublicKeyCallback: func(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			logrus.WithFields(logrus.Fields{
				"user": c.User(),
				"key":  key.Type(),
			}).Info("🔑 Public key auth")
			return nil, nil
		},
	}

	// Cargar clave privada del host
	hostKeyPath := getSettingsSftp("host_key")
	privateBytes, err := os.ReadFile(hostKeyPath)
	if err != nil {
		return fmt.Errorf("❌ failed to load host key at %s: %w", hostKeyPath, err)
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return fmt.Errorf("❌ failed to parse host key: %w", err)
	}
	h.config.AddHostKey(private)

	// Preparar directorio raíz
	if err := os.MkdirAll(h.root, 0755); err != nil {
		return fmt.Errorf("❌ failed to create root dir: %w", err)
	}
	for _, f := range def.Files {
		fullPath := filepath.Join(h.root, f.Path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			logrus.WithError(err).Warn("⚠️ Could not create intermediate dirs")
		}
		if err := os.WriteFile(fullPath, []byte(f.Content), 0644); err != nil {
			logrus.WithError(err).Warnf("⚠️ Could not write file %s", fullPath)
		}
	}

	// Iniciar listener TCP
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", h.port))
	if err != nil {
		return fmt.Errorf("❌ failed to listen on port %d: %w", h.port, err)
	}
	h.listener = listener

	logrus.Infof("🚀 SFTP server listening on port %d", h.port)

	go h.acceptConnections()
	return nil
}

// Acepta conexiones entrantes
func (h *SFTPHandler) acceptConnections() {
	for {
		conn, err := h.listener.Accept()
		if err != nil {
			logrus.WithError(err).Error("❌ Failed to accept connection")
			return
		}
		logrus.Info("📥 Incoming TCP connection")
		go h.handleConn(conn)
	}
}

// Maneja una conexión SSH y lanza subsistemas
func (h *SFTPHandler) handleConn(nConn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithField("panic", r).Error("💥 Panic recovered in handleConn")
		}
	}()

	sshConn, chans, reqs, err := ssh.NewServerConn(nConn, h.config)
	if err != nil {
		logrus.WithError(err).Error("❌ SSH handshake failed")
		nConn.Close()
		return
	}
	logrus.WithField("user", sshConn.User()).Info("✅ SSH connection established")
	defer func() {
		sshConn.Close()
		logrus.WithField("user", sshConn.User()).Info("👋 SSH connection closed")
	}()

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			logrus.Warnf("⚠️ Rejected unknown channel type: %s", newChannel.ChannelType())
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			logrus.WithError(err).Error("❌ Could not accept channel")
			continue
		}

		go func() {
			for req := range requests {
				if req.Type == "subsystem" && string(req.Payload[4:]) == "sftp" {
					logrus.Info("📦 Starting SFTP subsystem")
					server, err := sftp.NewServer(channel)
					if err != nil {
						logrus.WithError(err).Error("❌ Failed to start SFTP subsystem")
						channel.Close()
						return
					}

					if err := server.Serve(); err == io.EOF {
						logrus.Info("✅ SFTP session ended cleanly (EOF)")
					} else if err != nil {
						logrus.WithError(err).Error("❌ SFTP session error")
					} else {
						logrus.Info("✅ SFTP session ended normally")
					}
					channel.Close()
				}
			}
		}()
	}
}

// Detiene el servidor SFTP
func (h *SFTPHandler) Stop() error {
	if h.listener != nil {
		logrus.Info("🛑 Stopping SFTP server")
		return h.listener.Close()
	}
	return nil
}

// Devuelve la ruta del archivo de configuración
func getSettingsSftp(filename string) string {
	_, currentFile, _, _ := runtime2.Caller(0)
	baseDir := filepath.Join(filepath.Dir(currentFile), "..", "..", "settings")
	return filepath.Join(baseDir, filename)
}
