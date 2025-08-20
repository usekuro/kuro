package extensions

type Extension struct {
	Name    string
	Source  string // URL o path local
	Content string
}

type Registry struct {
	Extensions map[string]Extension
}

func NewRegistry() *Registry {
	return &Registry{Extensions: make(map[string]Extension)}
}

func (r *Registry) Register(name, content, source string) {
	r.Extensions[name] = Extension{
		Name:    name,
		Source:  source,
		Content: content,
	}
}

func (r *Registry) Get(name string) (Extension, bool) {
	ext, ok := r.Extensions[name]
	return ext, ok
}
