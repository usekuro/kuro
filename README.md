<div align="center">

# 🐱 UseKuro

**Mock any protocol like a master. No coding required. Flexible, powerful, shareable.**

[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green?style=for-the-badge)](LICENSE)
[![Release](https://img.shields.io/github/v/release/usekuro/kuro?style=for-the-badge)](https://github.com/usekuro/kuro/releases)
[![Tests](https://img.shields.io/badge/tests-passing-brightgreen?style=for-the-badge)](#testing)
[![Go Report Card](https://goreportcard.com/badge/github.com/usekuro/kuro?style=for-the-badge)](https://goreportcard.com/report/github.com/usekuro/kuro)

[**Documentation**](https://usekuro.com) • [**Examples**](examples/) • [**Contributing**](#-contributing) • [**Discord**](https://discord.gg/usekuro)

</div>

---

## ⚡ Quick Start

```bash
# Install UseKuro
go install github.com/usekuro/kuro/cmd/usekuro@latest

# Run your first HTTP mock
usekuro run examples/http_api.kuro

# 🚀 Mock is now running at http://localhost:8080
# ✅ Health check: http://localhost:8080/health
```

## 🌟 Why UseKuro?

UseKuro is the **ultimate protocol mocking tool** that lets you simulate complex systems using declarative `.kuro` files with embedded Go Templates. Think Postman collections, but for **any protocol** and **any complexity**.

<table>
<tr>
<td>

### ✨ **Before UseKuro**
```javascript
// Complex mock server setup
const express = require('express');
const app = express();

app.get('/users', (req, res) => {
  res.json({
    users: mockUsers,
    timestamp: new Date().toISOString()
  });
});

app.post('/orders', (req, res) => {
  // 50+ lines of validation logic
  // Database simulation
  // Error handling
  // ...
});

app.listen(8080);
```

</td>
<td>

### 🚀 **With UseKuro**
```yaml
protocol: http
port: 8080

routes:
  - path: /users
    method: GET
    response:
      status: 200
      body: |
        {
          "users": {{ .context.users | toJSON }},
          "timestamp": "{{ now }}"
        }

  - path: /orders
    method: POST
    response:
      status: 201
      body: |
        {
          "id": "{{ uuid }}",
          "total": {{ .input.total }},
          "created": "{{ now }}"
        }
```

</td>
</tr>
</table>

## 🎯 Features

<details>
<summary><b>🌐 Multi-Protocol Support</b></summary>

- **HTTP/HTTPS** - REST APIs, webhooks, microservices
- **TCP** - Custom protocols, database connections, message queues  
- **WebSocket** - Real-time applications, chat systems, live dashboards
- **SFTP** - File transfer simulation, development environments
- **More coming** - gRPC, UDP, MQTT on the roadmap

</details>

<details>
<summary><b>🎨 Template Engine</b></summary>

- **Go Templates** embedded in any field
- **Dynamic responses** based on input
- **Custom functions** for complex logic
- **Context variables** shared across requests
- **External imports** for reusable code

</details>

<details>
<summary><b>🏗️ Professional Features</b></summary>

- **Health checks** built into every mock
- **Structured logging** with configurable levels
- **Graceful shutdown** and error handling
- **Hot reload** during development
- **Docker support** for containerized deployments
- **Web UI** for managing multiple mocks

</details>

<details>
<summary><b>🤝 Developer Experience</b></summary>

- **Validation** before running
- **Auto-complete** friendly YAML schema
- **Comprehensive examples** for every protocol
- **Detailed error messages** with suggestions
- **Visual debugging** with request/response logs

</details>

## 📚 Protocols & Examples

### 🌐 HTTP API Mock

Perfect for **microservices**, **API testing**, and **frontend development**.

```yaml
protocol: http
port: 8080
meta:
  name: "E-Commerce API"
  description: "Complete REST API simulation"

context:
  variables:
    products:
      - id: 1
        name: "MacBook Pro"
        price: 2499.99
        stock: 15

routes:
  - path: /products
    method: GET
    response:
      status: 200
      headers:
        Content-Type: application/json
        X-Total-Count: "{{ len .context.products }}"
      body: |
        {
          "products": {{ .context.products | toJSON }},
          "timestamp": "{{ now }}",
          "version": "v1"
        }

  - path: /products/{{.path.id}}
    method: GET
    response:
      status: 200
      body: |
        {
          "product": {{range .context.products}}{{if eq .id (.path.id | atoi)}}{{. | toJSON}}{{end}}{{end}},
          "related": []
        }
```

### 🔌 TCP Server Mock

Great for **custom protocols**, **IoT devices**, and **message queues**.

```yaml
protocol: tcp
port: 9090
meta:
  name: "IoT Device Simulator"

session:
  timeout: 30s
  keepAlive: true

onMessage:
  match: "(?P<command>\\w+)\\s*(?P<params>.*)"
  conditions:
    - if: '{{ eq .command "TEMP" }}'
      respond: "TEMP_OK {{ random 18.0 25.0 }}C"
    
    - if: '{{ eq .command "STATUS" }}'
      respond: |
        STATUS_OK
        Device: IoT-Sensor-001
        Uptime: {{ now }}
        Battery: {{ random 65 100 }}%
        Signal: {{ random -70 -40 }}dBm
    
    - if: '{{ eq .command "RESET" }}'
      respond: "RESET_OK Device restarting..."
      
  else: "ERROR Unknown command: {{ .command }}"
```

### 🔄 WebSocket Real-time Mock

Ideal for **dashboards**, **chat applications**, and **live updates**.

```yaml
protocol: ws
port: 8080
meta:
  name: "Real-time Dashboard"

context:
  variables:
    metrics:
      cpu_usage: 45.2
      memory_usage: 67.8
      disk_usage: 23.1

onMessage:
  conditions:
    - if: '{{ contains .input "subscribe" }}'
      respond: |
        {
          "type": "subscription_confirmed",
          "channels": ["metrics", "alerts", "logs"]
        }
      
    - if: '{{ contains .input "get_metrics" }}'
      respond: |
        {
          "type": "metrics_update",
          "data": {
            "cpu": {{ add .context.metrics.cpu_usage (random -5.0 5.0) }},
            "memory": {{ add .context.metrics.memory_usage (random -2.0 2.0) }},
            "disk": {{ .context.metrics.disk_usage }},
            "timestamp": "{{ now }}"
          }
        }
```

### 📁 SFTP File Server Mock

Perfect for **file transfer testing** and **development environments**.

```yaml
protocol: sftp
port: 2222
meta:
  name: "Development File Server"

sftpAuth:
  username: "developer"
  password: "dev123"
  keyFile: "~/.ssh/id_rsa.pub"

context:
  variables:
    projectName: "UseKuro"
    version: "1.0.0"

files:
  - path: /README.md
    content: |
      # {{ .context.projectName }}
      
      Version: {{ .context.version }}
      Generated: {{ now }}
      
      Welcome to the {{ .context.projectName }} development server!

  - path: /config/database.yml
    content: |
      development:
        adapter: postgresql
        host: localhost
        port: 5432
        database: {{ .context.projectName | lower }}_dev
        
  - path: /logs/app.log
    content: |
      [{{ now }}] INFO - Application started
      [{{ now }}] INFO - Database connected
      [{{ now }}] INFO - Server listening on port 8080
```

## 🚀 Installation

### Option 1: Go Install (Recommended)
```bash
go install github.com/usekuro/kuro/cmd/usekuro@latest
```

### Option 2: Download Binary
```bash
# Linux/Mac
curl -sfL https://raw.githubusercontent.com/usekuro/kuro/main/install.sh | sh

# Or download from releases
wget https://github.com/usekuro/kuro/releases/latest/download/usekuro-linux-amd64.tar.gz
```

### Option 3: Docker
```bash
docker run --rm -p 8080:8080 -v $(pwd)/mocks:/mocks usekuro/kuro run /mocks/api.kuro
```

### Option 4: From Source
```bash
git clone https://github.com/usekuro/kuro.git
cd kuro
make build
./bin/usekuro --help
```

## 💡 Usage

### Basic Commands

```bash
# Run a single mock
usekuro run examples/http_api.kuro

# Validate mock file without running
usekuro validate examples/http_api.kuro

# Run multiple mocks from a directory
usekuro boot mocks/development/

# Start web management interface
usekuro web 8798
```

### Development Workflow

```bash
# 1. Create your mock
cat > my-api.kuro << EOF
protocol: http
port: 8080
routes:
  - path: /hello
    method: GET
    response:
      body: '{"message": "Hello {{ .input.name | default "World" }}!"}'
EOF

# 2. Validate it
usekuro validate my-api.kuro

# 3. Run with debug logging
LOG_LEVEL=debug usekuro run my-api.kuro

# 4. Test it
curl "http://localhost:8080/hello?name=UseKuro"
curl "http://localhost:8080/health"
```

### Production Deployment

```bash
# Environment configuration
export LOG_LEVEL=info
export USEKURO_CONFIG=/etc/usekuro/

# Run with systemd, docker, or kubernetes
usekuro boot /etc/usekuro/mocks/

# Health checks available at /health on each mock
curl http://localhost:8080/health
```

## 🏗️ Architecture

```
UseKuro/
├── cmd/usekuro/              # CLI application
├── internal/
│   ├── bootloader/           # Multi-mock orchestrator
│   ├── loader/               # .kuro file parser
│   ├── runtime/              # Protocol implementations
│   │   ├── http.go           # HTTP server runtime
│   │   ├── tcp.go            # TCP server runtime  
│   │   ├── ws.go             # WebSocket runtime
│   │   └── sftp.go           # SFTP server runtime
│   ├── template/             # Go template engine
│   ├── schema/               # Schema validation
│   └── web/                  # Management web UI
├── examples/                 # Example .kuro files
├── tests/                    # Integration tests
└── docs/                     # Documentation
```

## 📖 Template Functions

UseKuro provides a rich set of template functions:

### Time & Dates
```yaml
"{{ now }}"                    # 2024-01-15T10:30:00Z
"{{ now | date "2006-01-02" }}" # 2024-01-15
"{{ unix now }}"               # 1705315800
```

### Data Manipulation  
```yaml
"{{ .input | toJSON }}"        # Convert to JSON
"{{ .data | toPrettyJSON }}"   # Pretty printed JSON
"{{ random 1 100 }}"           # Random number 1-100
"{{ uuid }}"                   # Generate UUID
```

### String Operations
```yaml
"{{ .name | upper }}"          # UPPERCASE
"{{ .name | lower }}"          # lowercase  
"{{ .text | trim }}"           # Trim whitespace
"{{ split .path "/" }}"        # Split string
```

### Logic & Control
```yaml
{{ if eq .method "POST" }}...{{ end }}
{{ range .items }}...{{ end }}
{{ with .user }}...{{ end }}
{{ default .value "fallback" }}
```

### Custom Functions
```yaml
# Define reusable functions
functions:
  greeting: |
    {{ define "greeting" }}
    Hello {{ .name }}, welcome to {{ .service }}!
    {{ end }}

# Use in responses
body: '{{ template "greeting" . }}'
```

## 🧪 Testing

UseKuro includes comprehensive testing:

```bash
# Run all tests
make test

# Run with coverage  
make test-coverage

# Test specific protocols
go test -v ./internal/runtime/

# Integration tests
go test -v ./tests/integration/

# Benchmark tests
make bench
```

### Writing Tests

```go
func TestHTTPMock(t *testing.T) {
    // Load mock definition
    mock, err := loader.LoadMockFromFile("testdata/api.kuro")
    require.NoError(t, err)
    
    // Start mock server
    handler := runtime.NewHTTPHandler()
    err = handler.Start(mock)
    require.NoError(t, err)
    defer handler.Stop()
    
    // Test endpoints
    resp, err := http.Get("http://localhost:8080/health")
    require.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
}
```

## 🐳 Docker & Kubernetes

### Docker

```dockerfile
FROM usekuro/kuro:latest
COPY mocks/ /mocks/
EXPOSE 8080-8090
CMD ["usekuro", "boot", "/mocks/"]
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: usekuro-mocks
spec:
  replicas: 2
  selector:
    matchLabels:
      app: usekuro
  template:
    metadata:
      labels:
        app: usekuro
    spec:
      containers:
      - name: usekuro
        image: usekuro/kuro:latest
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: mocks
          mountPath: /mocks
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
      volumes:
      - name: mocks
        configMap:
          name: usekuro-mocks
```

## 🤝 Contributing

We welcome contributions! Here's how to get started:

### Quick Start
```bash
# Fork and clone
git clone https://github.com/yourusername/kuro.git
cd kuro

# Install dependencies
go mod download

# Run tests
make test

# Start development
make dev
```

### Development Guidelines

1. **Code Style**: We use `gofmt` and `golangci-lint`
2. **Testing**: Add tests for new features
3. **Documentation**: Update README and docs/
4. **Commit Messages**: Use [Conventional Commits](https://conventionalcommits.org/)

### Areas We Need Help

- 🌐 **New Protocol Support** (gRPC, UDP, MQTT)
- 🎨 **Web UI Improvements** 
- 📚 **Documentation & Examples**
- 🐛 **Bug Fixes & Performance**
- 🌍 **Internationalization**

### Making Your First Contribution

1. Check [Issues](https://github.com/usekuro/kuro/issues) for `good first issue` labels
2. Join our [Discord](https://discord.gg/usekuro) for questions
3. Read our [Contributing Guide](CONTRIBUTING.md)

## 📊 Roadmap

### v1.1 (Current)
- [x] HTTP, TCP, WebSocket, SFTP protocols
- [x] Template engine with custom functions
- [x] Web management interface
- [x] Docker support
- [x] Comprehensive testing

### v1.2 (Next Quarter)
- [ ] gRPC protocol support
- [ ] Plugin system for custom protocols
- [ ] Performance improvements
- [ ] Advanced debugging tools

### v2.0 (Future)
- [ ] Distributed mock orchestration
- [ ] Cloud deployment integrations
- [ ] Visual mock designer
- [ ] AI-powered mock generation

## 🙏 Acknowledgments

- Inspired by [Bruno](https://usebruno.com) for its focus on simplicity
- Go Templates for powerful templating capabilities
- The amazing Go community for excellent libraries
- All our [contributors](https://github.com/usekuro/kuro/graphs/contributors)

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=usekuro/kuro&type=Date)](https://star-history.com/#usekuro/kuro&Date)

---

<div align="center">

**UseKuro** - Mock any protocol like a master 🐱

[Website](https://usekuro.com) • [Documentation](https://docs.usekuro.com) • [Examples](examples/) • [Discord](https://discord.gg/usekuro)

**Made with ❤️ by the UseKuro community**

</div>