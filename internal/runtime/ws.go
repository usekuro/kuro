package runtime

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/usekuro/usekuro/internal/schema"
	"github.com/usekuro/usekuro/internal/template"
)

type WSHandler struct {
	upgrader websocket.Upgrader
	logger   *logrus.Entry
}

func NewWSHandler() *WSHandler {
	return &WSHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},

		logger: logrus.WithField("protocol", "ws"),
	}
}

func (h *WSHandler) Start(def *schema.MockDefinition) error {
	h.logger.Logger.SetLevel(logrus.DebugLevel)
	h.logger.Infof("starting WebSocket mock on port %d", def.Port)

	registry := loadExtensions(def.Import, h.logger)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := h.upgrader.Upgrade(w, r, nil)
		if err != nil {
			h.logger.WithError(err).Error("failed to upgrade WebSocket connection")
			return
		}
		defer conn.Close()
		h.logger.Info("client connected")

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				h.logger.WithError(err).Info("client disconnected")
				break
			}
			raw := string(msg)
			h.logger.WithField("input", raw).Info("received message")

			matches := extractVars(raw, def.OnMessage.Match)
			ctx := template.MergeContext(matches, nil, def.Context.Variables)

			tpl, err := template.NewRuntime(ctx, registry)
			if err != nil {
				h.logger.WithError(err).Error("template runtime error")
				conn.WriteMessage(websocket.TextMessage, []byte("template error"))
				continue
			}

			sent := false
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
					conn.WriteMessage(websocket.TextMessage, []byte(resp))
					sent = true
					break
				}
			}

			if !sent && def.OnMessage.Else != "" {
				resp, _ := tpl.Render("else", def.OnMessage.Else)
				h.logger.WithField("response", resp).Info("sending fallback response")
				conn.WriteMessage(websocket.TextMessage, []byte(resp))
			}
		}
	})

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", def.Port), nil)
		if err != nil {
			h.logger.WithError(err).Fatal("failed to start WebSocket server")
		}
	}()
	return nil
}

func (h *WSHandler) Stop() error {
	// No-op for now
	return nil
}
