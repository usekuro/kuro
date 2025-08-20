package runtime

import (
	"regexp"

	"github.com/sirupsen/logrus"
	"github.com/usekuro/usekuro/internal/extensions"
	"github.com/usekuro/usekuro/internal/schema"
)

type ProtocolHandler interface {
	Start(def *schema.MockDefinition) error
	Stop() error
}

func loadExtensions(imports []string, logger *logrus.Entry) *extensions.Registry {
	registry := extensions.NewRegistry()
	for _, src := range imports {
		code, err := extensions.LoadKurof(src)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"file": src,
				"err":  err,
			}).Warn("failed to load .kurof file")
			continue
		}
		logger.WithField("file", src).Info("loaded .kurof file")
		registry.Register(src, code, src)
	}
	return registry
}

func extractVars(input, pattern string) map[string]any {
	if pattern == "" {
		return map[string]any{"msg": input}
	}
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(input)
	names := re.SubexpNames()

	result := map[string]any{}
	for i, name := range names {
		if i != 0 && name != "" && i < len(match) {
			result[name] = match[i]
		}
	}
	return result
}
