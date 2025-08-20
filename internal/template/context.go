package template

func MergeContext(input, session, global map[string]any) map[string]any {
	if session == nil {
		session = map[string]any{}
	}
	if global == nil {
		global = map[string]any{}
	}
	if input == nil {
		input = map[string]any{}
	}
	return map[string]any{
		"input":   input,
		"session": session,
		"context": global,
	}
}
