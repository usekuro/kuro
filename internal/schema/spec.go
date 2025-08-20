package schema

type Meta struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// HTTP route
type Route struct {
	Path     string             `json:"path"`
	Method   string             `json:"method"`
	Response ResponseDefinition `json:"response"`
}

type ResponseDefinition struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// TCP / WS conditional logic
type OnMessageRule struct {
	If      string `json:"if"`
	Respond string `json:"respond"`
}

type OnMessage struct {
	Match      string          `json:"match"`
	Conditions []OnMessageRule `json:"conditions"`
	Else       string          `json:"else"`
}

// SFTP file system
type FileEntry struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type SFTPAuth struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	PublicKeyPath string `json:"publicKeyPath"` // optional
}

type Session struct {
	Timeout string `json:"timeout"`
}

type Context struct {
	Variables map[string]any `json:"variables"`
}

type MockDefinition struct {
	Protocol  string            `json:"protocol"` // http, tcp, ws, sftp
	Port      int               `json:"port"`
	Meta      Meta              `json:"meta"`
	Routes    []Route           `json:"routes"`    // http
	OnMessage *OnMessage        `json:"onMessage"` // tcp/ws
	Files     []FileEntry       `json:"files"`     // sftp
	SFTPAuth  *SFTPAuth         `json:"sftpAuth"`  // sftp credentials
	Session   *Session          `json:"session"`   // optional
	Context   *Context          `json:"context"`   // optional
	Functions map[string]string `json:"functions"` // optional
	Import    []string          `json:"import"`    // optional
}
