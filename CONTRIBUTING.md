# Contributing to UseKuro 🐱

Thank you for your interest in contributing to UseKuro! We're excited to have you join our community of developers working to make protocol mocking accessible and powerful.

## 🌟 Quick Start for Contributors

1. **Fork** the repository on GitHub
2. **Clone** your fork locally
3. **Create** a feature branch
4. **Make** your changes
5. **Test** your changes
6. **Submit** a pull request

```bash
# Fork and clone
git clone https://github.com/yourusername/kuro.git
cd kuro

# Install dependencies
go mod download

# Run tests to ensure everything works
make test

# Create your feature branch
git checkout -b feature/amazing-feature
```

## 🎯 Ways to Contribute

### 🐛 Bug Reports

Found a bug? Help us fix it!

- **Search** existing issues first
- **Use** our bug report template
- **Include** reproduction steps
- **Attach** logs and configuration files

### 💡 Feature Requests

Have an idea for improvement?

- **Check** the roadmap first
- **Open** a feature request issue
- **Describe** the use case clearly
- **Consider** implementation approaches

### 📝 Documentation

Documentation is crucial for adoption:

- **Fix** typos and clarify confusing sections
- **Add** examples for edge cases
- **Improve** API documentation
- **Translate** to other languages

### 🔧 Code Contributions

#### Areas We Need Help

- **🌐 New Protocol Support**
  - gRPC implementation
  - UDP protocol handler
  - MQTT broker simulation
  - Custom protocol plugins

- **🎨 Web UI Improvements**
  - Real-time monitoring
  - Mock designer interface
  - Better error handling
  - Mobile responsiveness

- **🚀 Performance Optimizations**
  - Template caching
  - Concurrent request handling
  - Memory usage optimization
  - Startup time improvements

- **🧪 Testing & Quality**
  - Integration test scenarios
  - Load testing utilities
  - Error handling improvements
  - Edge case coverage

## 🛠️ Development Setup

### Prerequisites

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Git** - Version control
- **Make** - Build automation
- **Docker** (optional) - Container testing

### Development Environment

```bash
# Clone the repository
git clone https://github.com/usekuro/kuro.git
cd kuro

# Install dependencies
go mod download

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/swaggo/swag/cmd/swag@latest

# Run all tests
make test

# Build the application
make build

# Run with live reload during development
make dev
```

### Project Structure

```
usekuro/
├── cmd/usekuro/              # Main application entry point
├── internal/
│   ├── bootloader/           # Multi-mock orchestration
│   ├── extensions/           # Extension system
│   ├── loader/               # .kuro file loading
│   ├── runtime/              # Protocol implementations
│   │   ├── http.go           # HTTP protocol handler
│   │   ├── tcp.go            # TCP protocol handler
│   │   ├── ws.go             # WebSocket handler
│   │   └── sftp.go           # SFTP handler
│   ├── schema/               # Schema definitions
│   ├── template/             # Template engine
│   └── web/                  # Web management interface
├── examples/                 # Example .kuro files
├── tests/                    # Integration tests
├── docs/                     # Documentation
└── scripts/                  # Build and deployment scripts
```

## 📋 Development Guidelines

### Code Style

We follow Go best practices and use automated tools:

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Check for security issues
gosec ./...
```

### Commit Messages

We use [Conventional Commits](https://conventionalcommits.org/):

```
feat: add gRPC protocol support
fix: resolve template rendering issue with arrays
docs: update API documentation
test: add integration tests for WebSocket
refactor: improve error handling in HTTP runtime
```

### Branch Naming

- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation updates
- `refactor/description` - Code refactoring

### Pull Request Process

1. **Update** documentation if needed
2. **Add** tests for new functionality
3. **Ensure** all tests pass
4. **Update** CHANGELOG.md
5. **Request** review from maintainers

## 🧪 Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test package
go test -v ./internal/runtime/

# Run integration tests
go test -v ./tests/integration/

# Run benchmarks
make bench
```

### Writing Tests

#### Unit Tests

```go
func TestHTTPHandler_Start(t *testing.T) {
    tests := []struct {
        name    string
        mock    *schema.MockDefinition
        wantErr bool
    }{
        {
            name: "valid HTTP mock",
            mock: &schema.MockDefinition{
                Protocol: "http",
                Port:     8080,
                Routes: []schema.Route{
                    {
                        Path:   "/test",
                        Method: "GET",
                        Response: schema.ResponseDefinition{
                            Status: 200,
                            Body:   "OK",
                        },
                    },
                },
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            h := NewHTTPHandler()
            err := h.Start(tt.mock)
            if (err != nil) != tt.wantErr {
                t.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
            }
            if err == nil {
                defer h.Stop()
            }
        })
    }
}
```

#### Integration Tests

```go
func TestHTTPMockIntegration(t *testing.T) {
    // Load mock from file
    mock, err := loader.LoadMockFromFile("testdata/api.kuro")
    require.NoError(t, err)

    // Start handler
    handler := runtime.NewHTTPHandler()
    err = handler.Start(mock)
    require.NoError(t, err)
    defer handler.Stop()

    // Test endpoints
    resp, err := http.Get("http://localhost:8080/health")
    require.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, 200, resp.StatusCode)
}
```

## 📚 Adding New Protocols

Want to add support for a new protocol? Here's how:

### 1. Define the Protocol Interface

```go
// internal/runtime/myprotocol.go
type MyProtocolHandler struct {
    server *MyProtocolServer
    logger *logrus.Entry
}

func NewMyProtocolHandler() *MyProtocolHandler {
    return &MyProtocolHandler{
        logger: logrus.WithField("protocol", "myprotocol"),
    }
}

func (h *MyProtocolHandler) Start(def *schema.MockDefinition) error {
    h.logger.Infof("starting MyProtocol mock on port %d", def.Port)
    
    // Implementation here
    
    return nil
}

func (h *MyProtocolHandler) Stop() error {
    if h.server != nil {
        h.logger.Info("stopping MyProtocol mock")
        return h.server.Close()
    }
    return nil
}
```

### 2. Add to Protocol Factory

```go
// cmd/usekuro/main.go
case "myprotocol":
    handler = runtimepkg.NewMyProtocolHandler()
```

### 3. Add Schema Support

```go
// internal/schema/mock.go
// Add validation rules for your protocol
```

### 4. Create Examples

```yaml
# examples/myprotocol_example.kuro
protocol: myprotocol
port: 9090
meta:
  name: "MyProtocol Example"
  description: "Example mock for MyProtocol"

# Protocol-specific configuration
myprotocol:
  timeout: 30s
  
# Your protocol handlers
```

### 5. Add Tests

Create comprehensive tests covering:
- Protocol-specific features
- Error handling
- Integration scenarios
- Performance benchmarks

## 🎨 Web UI Development

The web interface uses vanilla JavaScript and modern CSS:

```bash
# Start development server
make web-dev

# Build assets
make web-build

# The web UI is located in:
# - web/index.html
# - web/static/css/
# - web/static/js/
```

### UI Guidelines

- **Responsive** design for mobile and desktop
- **Accessible** with proper ARIA labels
- **Fast** with minimal JavaScript
- **Consistent** with the UseKuro design system

## 🚀 Performance Considerations

### Benchmarking

```bash
# Run performance benchmarks
make bench

# Profile specific operations
go test -bench=. -cpuprofile=cpu.prof ./internal/runtime/
go tool pprof cpu.prof
```

### Memory Usage

- Use object pooling for frequent allocations
- Implement proper cleanup in Stop() methods
- Monitor goroutine leaks

### Concurrency

- All protocol handlers should be thread-safe
- Use context.Context for cancellation
- Implement proper shutdown sequences

## 📖 Documentation Standards

### Code Documentation

```go
// LoadMockFromFile loads a mock definition from a .kuro file.
// It supports both YAML and JSON formats, with automatic detection
// based on file extension.
//
// The function validates the schema after loading and returns
// detailed error messages for invalid configurations.
//
// Example:
//   mock, err := LoadMockFromFile("api.kuro")
//   if err != nil {
//       log.Fatal(err)
//   }
func LoadMockFromFile(path string) (*schema.MockDefinition, error) {
    // Implementation
}
```

### Example Documentation

Each example should include:
- Clear description of the use case
- Step-by-step setup instructions
- Expected behavior
- Common variations

## 🤝 Community Guidelines

### Communication

- **Be respectful** and inclusive
- **Help newcomers** get started
- **Share knowledge** and experiences
- **Provide constructive feedback**

### Getting Help

- **GitHub Discussions** - General questions and ideas
- **Discord** - Real-time chat and support
- **GitHub Issues** - Bug reports and feature requests
- **Stack Overflow** - Tag questions with `usekuro`

### Recognition

Contributors are recognized through:
- **Contributors list** in README
- **Release notes** mentions
- **Special badges** for significant contributions
- **Maintainer status** for consistent contributors

## 🎯 Roadmap Priorities

### Current Focus (v1.2)

- [ ] **gRPC Protocol Support** - High demand feature
- [ ] **Plugin System** - Allow external protocols
- [ ] **Performance Optimization** - Template caching
- [ ] **Advanced Debugging** - Request/response inspection

### Future Goals (v2.0)

- [ ] **Distributed Mocking** - Multi-node orchestration
- [ ] **AI-Powered Generation** - Auto-generate mocks
- [ ] **Visual Designer** - Drag-and-drop interface
- [ ] **Cloud Integration** - Deploy to cloud providers

## 📋 Release Process

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):
- **MAJOR** - Breaking changes
- **MINOR** - New features (backward compatible)
- **PATCH** - Bug fixes

### Release Checklist

- [ ] Update CHANGELOG.md
- [ ] Run full test suite
- [ ] Update documentation
- [ ] Create release notes
- [ ] Tag version in Git
- [ ] Build and publish binaries
- [ ] Update Docker images
- [ ] Announce release

## ❓ FAQ

### How do I debug template rendering issues?

Use debug logging to see template execution:

```bash
LOG_LEVEL=debug usekuro run your-mock.kuro
```

### Can I contribute without Go knowledge?

Absolutely! We need help with:
- Documentation improvements
- Example .kuro files
- UI/UX design
- Testing and bug reports

### How do I propose architectural changes?

1. Open a GitHub Discussion
2. Describe the problem and proposed solution
3. Get feedback from maintainers
4. Create a detailed design document
5. Submit implementation in phases

### What if my feature request is rejected?

- We'll explain our reasoning
- You can implement it as a plugin
- Consider creating a fork
- Keep the discussion open for future consideration

## 🙏 Thank You

Every contribution, no matter how small, makes UseKuro better for everyone. Thank you for being part of our community!

---

**Happy coding! 🐱**

For questions about contributing, reach out to us on [Discord](https://discord.gg/usekuro) or open a [GitHub Discussion](https://github.com/usekuro/kuro/discussions).