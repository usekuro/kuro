package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/usekuro/usekuro/internal/extensions"
	"github.com/usekuro/usekuro/internal/schema"
	"github.com/usekuro/usekuro/internal/template"
)

type HTTPHandler struct {
	server *http.Server
	logger *logrus.Entry
}

func NewHTTPHandler() *HTTPHandler {
	return &HTTPHandler{
		logger: logrus.WithField("protocol", "http"),
	}
}

func (h *HTTPHandler) Start(def *schema.MockDefinition) error {
	h.logger.Infof("starting HTTP mock on port %d", def.Port)

	// Single extensions registry for all routes of this mock
	registry := extensions.NewRegistry()
	for _, src := range def.Import {
		code, err := extensions.LoadKurof(src)
		if err != nil {
			h.logger.WithField("file", src).WithError(err).Warn("failed to load .kurof file")
			continue
		}
		registry.Register(src, code, src)
		h.logger.WithField("file", src).Info("loaded .kurof file")
	}

	mux := http.NewServeMux()

	// Skip default health endpoints if custom ones exist
	hasCustomHealth := false
	for _, route := range def.Routes {
		if route.Path == "/health" || route.Path == "/healthz" {
			hasCustomHealth = true
			break
		}
	}

	if !hasCustomHealth {
		// ---------------------------
		// Default HEALTH endpoint(s)
		// ---------------------------
		healthHandler := func(w http.ResponseWriter, r *http.Request) {
			// CORS-friendly by default (useful for dashboards)
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Cache-Control", "no-store")

			switch r.Method {
			case http.MethodOptions:
				w.WriteHeader(http.StatusNoContent)
				return
			case http.MethodHead:
				w.WriteHeader(http.StatusOK)
				return
			default:
				// GET
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
			}
		}
		mux.HandleFunc("/health", healthHandler)
		mux.HandleFunc("/healthz", healthHandler)
	}

	registeredPaths := make(map[string]bool)
	routeHandlers := make(map[string][]schema.Route)

	// Create initial template runtime for path processing
	contextVars := def.Context.Variables
	if contextVars == nil {
		contextVars = make(map[string]interface{})
	}

	// Create full context structure that matches .kuro file expectations
	// Template expects .context.apiVersion, so we need to create a flattened structure
	fullContext := make(map[string]interface{})

	// Add all variables at root level for backward compatibility
	for k, v := range contextVars {
		fullContext[k] = v
	}

	// Create the context structure with direct access to variables
	contextStruct := make(map[string]interface{})
	for k, v := range contextVars {
		contextStruct[k] = v
	}
	fullContext["context"] = contextStruct

	initialTpl, err := template.NewRuntime(fullContext, registry)
	if err != nil {
		return fmt.Errorf("failed to create template runtime: %w", err)
	}

	// Group routes by path, processing templates in paths
	for _, route := range def.Routes {
		routePath := route.Path

		// Process template in route path if it contains template syntax
		if strings.Contains(routePath, "{{") {

			processedPath, err := initialTpl.Render("route-path", routePath)
			if err != nil {
				h.logger.WithError(err).Warnf("failed to process template in path %s, using original", routePath)
			} else {

				routePath = processedPath
			}
		}

		routeHandlers[routePath] = append(routeHandlers[routePath], route)
	}

	// Register each unique path once
	for path, routes := range routeHandlers {
		if registeredPaths[path] {
			h.logger.Warnf("skipping duplicate path registration: %s", path)
			continue
		}
		registeredPaths[path] = true

		h.logger.WithField("path", path).Info("registering route")

		// capture loop variable
		routesCopy := routes

		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			// Find the matching route for this method
			var routeCopy schema.Route
			found := false
			for _, rt := range routesCopy {
				// Empty rt.Method = wildcard (any method)
				if strings.EqualFold(rt.Method, r.Method) || rt.Method == "" {
					routeCopy = rt
					found = true
					break
				}
			}

			if !found {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			// Parse request body for POST/PUT/PATCH requests (JSON only)
			var inputVars map[string]interface{}
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				// tolerate content-type with charset (e.g., application/json; charset=utf-8)
				ct := r.Header.Get("Content-Type")
				if ct != "" && strings.HasPrefix(strings.ToLower(ct), "application/json") {
					decoder := json.NewDecoder(r.Body)
					if err := decoder.Decode(&inputVars); err != nil {
						h.logger.WithError(err).Warn("failed to parse JSON body")
					}
				}
			}

			// Prepare context with request data
			var contextVars map[string]interface{}
			if def.Context != nil {
				contextVars = def.Context.Variables
			}

			// Merge all contexts with priority: input > route params (nil here) > context vars
			ctx := template.MergeContext(inputVars, nil, contextVars)

			tpl, err := template.NewRuntime(ctx, registry)
			if err != nil {
				h.logger.WithError(err).Error("template runtime error")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Dynamic headers with error handling
			for k, v := range routeCopy.Response.Headers {
				hdr, err := tpl.Render("hdr", v)
				if err != nil {
					h.logger.WithError(err).Warnf("failed to render header %s, using raw value", k)
					hdr = v // fallback to raw value
				}
				h.logger.WithFields(logrus.Fields{
					"header": k,
					"value":  hdr,
				}).Debug("rendered header")
				w.Header().Set(k, hdr)
			}

			// Dynamic body with error handling
			body, err := tpl.Render("body", routeCopy.Response.Body)
			if err != nil {
				h.logger.WithError(err).Error("failed to render response body")
				body = `{"error": "template rendering failed"}`
				w.Header().Set("Content-Type", "application/json")
			}

			h.logger.WithFields(logrus.Fields{
				"method": r.Method,
				"path":   r.URL.Path,
				"status": routeCopy.Response.Status,
			}).Info("sending HTTP response")

			w.WriteHeader(routeCopy.Response.Status)
			_, _ = w.Write([]byte(body))
		})
	}

	h.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", def.Port),
		Handler: mux,
	}

	// Start server in background with proper error handling
	errChan := make(chan error, 1)
	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.logger.WithError(err).Error("HTTP server failed - attempting to continue")
			errChan <- err
		}
	}()

	// Give server time to start and check for immediate failures
	select {
	case err := <-errChan:
		// Server failed to start (e.g., port in use)
		return fmt.Errorf("failed to start HTTP server: %w", err)
	case <-time.After(100 * time.Millisecond):
		// Server started successfully
		h.logger.Info("HTTP server started successfully")
	}

	return nil
}

func (h *HTTPHandler) Stop() error {
	if h.server != nil {
		h.logger.Info("stopping HTTP mock")

		// Give the server 5 seconds to gracefully shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.server.Shutdown(ctx); err != nil {
			h.logger.WithError(err).Warn("graceful shutdown failed, forcing close")
			return h.server.Close()
		}

		h.logger.Info("HTTP server stopped gracefully")
	}
	return nil
}
