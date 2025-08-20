package runtime

import (
	"errors"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/usekuro/usekuro/internal/schema"
	"github.com/usekuro/usekuro/internal/template"
)

type TCPHandler struct {
	Port   int
	ln     net.Listener
	logger *logrus.Entry
}

func NewTCPHandler() *TCPHandler {
	return &TCPHandler{
		logger: logrus.WithField("protocol", "tcp"),
	}
}

func (h *TCPHandler) Start(def *schema.MockDefinition) error {
	var err error
	h.ln, err = net.Listen("tcp", fmt.Sprintf(":%d", def.Port))
	if err != nil {
		h.logger.WithError(err).Error("failed to start TCP listener")
		return err
	}

	h.logger.Infof("TCP mock listening on port %d", def.Port)

	go func() {
		for {
			conn, err := h.ln.Accept()
			if err != nil {
				// Si el listener fue cerrado, salimos limpio
				var opErr *net.OpError
				if errors.As(err, &opErr) && opErr.Err.Error() == "use of closed network connection" {
					h.logger.WithError(err).Info("TCP listener cerrado")
					return
				}
				h.logger.WithError(err).Error("failed to accept TCP connection")
				continue
			}
			go h.handleConnection(conn, def)
		}
	}()

	return nil
}

func (h *TCPHandler) Stop() error {
	if h.ln != nil {
		h.logger.Info("stopping TCP mock")
		return h.ln.Close()
	}
	return nil
}

func (h *TCPHandler) handleConnection(conn net.Conn, def *schema.MockDefinition) {
	defer conn.Close()

	if def == nil {
		h.logger.Warn("No TCP definition found for message")
		conn.Write([]byte("error: def is nil\n"))
		return
	}

	// Defensive: If OnMessage is nil, respond with error and log
	if def.OnMessage == nil {
		h.logger.Error("OnMessage is nil in TCP mock definition")
		conn.Write([]byte("error: OnMessage is nil\n"))
		return
	}

	defer conn.Close()

	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		h.logger.WithError(err).Warn("failed to read from TCP client")
		return
	}

	rawInput := string(buf[:n])
	h.logger.WithField("input", rawInput).Info("received message")

	matches := extractVars(rawInput, def.OnMessage.Match)
	registry := loadExtensions(def.Import, h.logger)
	ctx := template.MergeContext(matches, nil, def.Context.Variables)

	tpl, err := template.NewRuntime(ctx, registry)
	if err != nil {
		h.logger.WithError(err).Error("template runtime creation failed")
		conn.Write([]byte("error de template"))
		return
	}

	for i, cond := range def.OnMessage.Conditions {
		result, _ := tpl.Render(fmt.Sprintf("cond_%d", i), cond.If)
		h.logger.WithFields(logrus.Fields{
			"condition": i,
			"if":        cond.If,
			"result":    result,
		}).Debug("evaluated condition")

		if result == "true" {
			resp, _ := tpl.Render(fmt.Sprintf("resp_%d", i), cond.Respond)
			h.logger.WithField("response", resp).Info("sending matched response")
			if len(resp) > 0 && resp[len(resp)-1] != '\n' {
				resp += "\n"
			}
			if len(resp) > 0 && resp[len(resp)-1] != '\n' {
				resp += "\n"
			}
			conn.Write([]byte(resp))
			return
		}
	}

	if def.OnMessage.Else != "" {
		resp, _ := tpl.Render("else", def.OnMessage.Else)
		h.logger.WithField("response", resp).Info("sending fallback response")
		if len(resp) > 0 && resp[len(resp)-1] != '\n' {
			resp += "\n"
		}
		if len(resp) > 0 && resp[len(resp)-1] != '\n' {
			resp += "\n"
		}
		conn.Write([]byte(resp))
	}
}
