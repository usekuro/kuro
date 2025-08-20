package extensions

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
)

func LoadKurof(source string) (string, error) {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		resp, err := http.Get(source)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return "", errors.New("failed to fetch remote kurof")
		}
		body, err := io.ReadAll(resp.Body)
		return string(body), err
	}
	// local file
	data, err := os.ReadFile(source)
	return string(data), err
}
