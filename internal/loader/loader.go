package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/usekuro/usekuro/internal/schema"
	"gopkg.in/yaml.v3"
)

func LoadMockFromFile(path string) (*schema.MockDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", path, err)
	}

	def := &schema.MockDefinition{}
	isJSON := strings.HasSuffix(path, ".json")

	if isJSON {
		if err := json.Unmarshal(data, def); err != nil {
			return nil, fmt.Errorf("invalid JSON format: %w", err)
		}
	} else {
		if err := yaml.Unmarshal(data, def); err != nil {
			return nil, fmt.Errorf("invalid YAML format: %w", err)
		}
	}

	if err := schema.Validate(def); err != nil {
		return nil, fmt.Errorf("schema validation failed: %w", err)
	}

	return def, nil
}
