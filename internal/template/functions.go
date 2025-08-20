package template

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

func FuncMap() map[string]any {
	return map[string]any{
		"now":  func() string { return time.Now().Format(time.RFC3339) },
		"uuid": func() string { return uuid.NewString() },
		"toJSON": func(v any) string {
			b, err := json.Marshal(v)
			if err != nil {
				return fmt.Sprintf(`"error: %v"`, err)
			}
			return string(b)
		},
		"contains":   safeContains,
		"regexMatch": safeRegexMatch,
		"upper":      safeUpper,
		"lower":      safeLower,
		"title":      safeTitle,
		"trim":       safeTrim,
		"split":      safeSplit,
		"join":       safeJoin,
		"replace":    safeReplace,
		"len":        safeLen,
		"default":    safeDefault,
	}
}

func safeContains(s, substr any) bool {
	str, ok1 := s.(string)
	sub, ok2 := substr.(string)
	return ok1 && ok2 && strings.Contains(str, sub)
}

func safeRegexMatch(pat, val any) bool {
	p, ok1 := pat.(string)
	v, ok2 := val.(string)
	if !ok1 || !ok2 {
		return false
	}
	m, _ := regexp.MatchString(p, v)
	return m
}

func safeUpper(val any) string {
	return strings.ToUpper(fmt.Sprint(val))
}

func safeLower(val any) string {
	s, ok := val.(string)
	if !ok {
		return "error: not a string"
	}
	return strings.ToLower(s)
}

func safeTitle(val any) string {
	s, ok := val.(string)
	if !ok {
		return "error: not a string"
	}
	return strings.Title(s)
}

func safeTrim(val any) string {
	s, ok := val.(string)
	if !ok {
		return "error: not a string"
	}
	return strings.TrimSpace(s)
}

func safeSplit(str, sep any) []string {
	s, ok1 := str.(string)
	sepStr, ok2 := sep.(string)
	if !ok1 || !ok2 {
		return []string{}
	}
	return strings.Split(s, sepStr)
}

func safeJoin(arr any, sep string) string {
	list, ok := arr.([]string)
	if !ok {
		return "error: expected []string"
	}
	return strings.Join(list, sep)
}

func safeReplace(s, old, new any) string {
	str, ok1 := s.(string)
	oldStr, ok2 := old.(string)
	newStr, ok3 := new.(string)
	if !ok1 || !ok2 || !ok3 {
		return "error: invalid args"
	}
	return strings.ReplaceAll(str, oldStr, newStr)
}

func safeLen(v any) int {
	val := reflect.ValueOf(v)
	if !val.IsValid() {
		return 0
	}
	switch val.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return val.Len()
	default:
		return 0
	}
}

func safeDefault(value, defaultValue any) any {
	if value == nil {
		return defaultValue
	}

	// Check if value is empty string
	if str, ok := value.(string); ok && str == "" {
		return defaultValue
	}

	// Check if value is zero value using reflection
	val := reflect.ValueOf(value)
	if !val.IsValid() || val.IsZero() {
		return defaultValue
	}

	return value
}
