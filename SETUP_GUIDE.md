# UseKuro Setup and Implementation Guide

## 🚀 Quick Start

This guide walks you through setting up UseKuro with the new mock definition generation and workspace management features.

## 📋 Prerequisites

- Go 1.19 or later
- Git
- curl (for testing)
- expect (optional, for SFTP testing)

## 🔧 Installation

### 1. Clone and Build

```bash
# Clone the repository
git clone <repository-url>
cd usekuro

# Build the project
go mod tidy
go build -o bin/usekuro cmd/usekuro/main.go
```

### 2. Verify Installation

```bash
# Check that the binary was created
ls -la bin/usekuro

# Should show the executable file
```

## 🏃‍♂️ Running UseKuro

### Start the Server

```bash
# Start UseKuro web server
./bin/usekuro web

# Server will start on http://localhost:8080
```

### Access the Dashboard

Open your browser and navigate to: http://localhost:8080

## ✨ New Features Overview

### 1. Automatic Mock Definition Generation

**Before**: Creating mocks from the frontend resulted in "Mock definition not found" error.

**Now**: Mock definitions are automatically generated based on protocol type.

### 2. Workspace Management

**New Feature**: Create, switch, and delete user workspaces to organize your mocks.

### 3. Enhanced Protocol Support

All protocols now have working default configurations:
- **HTTP/HTTPS**: REST API endpoints
- **SFTP**: File server with authentication
- **TCP**: Echo server
- **WebSocket**: Ping-pong server

## 📖 Usage Guide

### Creating Your First Mock

1. **Open Dashboard**: Navigate to http://localhost:8080
2. **Go to User Workspace**: Click the "User Workspace" tab
3. **Review Examples**: See the "📚 Example Mocks" section (read-only templates)
4. **Create Your Mock**: Click "➕ New Mock" in the "👤 Your Mocks" section
5. **Fill Details**:
   - Name: "My API Mock"
   - Protocol: "http"
   - Port: 8081
   - Description: "Test API for development"
6. **Save**: Click "Create"
7. **Start Mock**: Go to "Mock Services" tab and click "Start"

✅ **No more "Mock definition not found" error!**
✅ **Example mocks preserved as templates!**

### Managing Workspaces

1. **Open Workspace Manager**: In "User Workspace" tab, click "🏢 Workspaces"
2. **Create New Workspace**:
   - Enter name: "project-alpha"
   - Click "Create"
3. **Switch Workspace**: Click "Switch" next to any workspace
4. **Start Fresh**: New workspace shows zero user mocks but keeps example templates
5. **Delete Workspace**: Click "Delete" (only deletes YOUR mocks, preserves examples)

💡 **Example mocks are always available as read-only templates in every workspace**

### Testing SFTP Functionality

UseKuro now includes a built-in SFTP server:

```bash
# Connect to SFTP server
sftp -P 2222 usekuro@localhost

# Default credentials:
# Username: usekuro
# Password: kuro123

# Try these commands:
ls                    # List files
get welcome.txt       # Download file
put local_file.txt    # Upload file
quit                  # Exit
```

## 🧪 Testing Your Implementation

Run the comprehensive test suite:

```bash
# Make test script executable
chmod +x test_implementation.sh

# Run all tests
./test_implementation.sh
```

The test will verify:
- ✅ Mock creation without errors
- ✅ Workspace management
- ✅ Example mock protection
- ✅ User/Example mock separation
- ✅ SFTP functionality
- ✅ API endpoints
- ✅ Protocol-specific configurations

## 📁 Generated Mock Examples

### HTTP Mock
```yaml
protocol: http
port: 8080
meta:
  name: "My HTTP Mock"
  description: "Test HTTP service"
routes:
  - path: /
    method: GET
    response:
      status: 200
      headers:
        Content-Type: application/json
      body: |
        {"message": "Hello from My HTTP Mock", "status": "ok"}
```

### SFTP Mock
```yaml
protocol: sftp
port: 2222
meta:
  name: "My SFTP Mock"
  description: "Test SFTP service"
sftpAuth:
  username: "usekuro"
  password: "kuro123"
files:
  - path: "/welcome.txt"
    content: "Welcome to My SFTP Mock SFTP server!"
  - path: "/data/sample.json"
    content: '{"server": "My SFTP Mock", "protocol": "sftp"}'
```

### TCP Mock
```yaml
protocol: tcp
port: 9999
meta:
  name: "My TCP Mock"
  description: "Test TCP service"
onMessage:
  match: ".*"
  conditions:
    - if: "HELLO"
      respond: "HELLO from My TCP Mock"
  else: "Echo: {{message}}"
```

### WebSocket Mock
```yaml
protocol: websocket
port: 8088
meta:
  name: "My WebSocket Mock"
  description: "Test WebSocket service"
onMessage:
  match: ".*"
  conditions:
    - if: "ping"
      respond: "pong from My WebSocket Mock"
  else: "Echo: {{message}}"
```

## 🔧 Configuration Files

### Workspace Structure
```
workspaces/
├── default/              # Protected default workspace
│   ├── mocks/           # Mock definitions (.kuro files)
│   ├── configs/         # Configuration files
│   └── config.json      # Workspace settings
└── {user-workspace}/    # User-created workspaces
    ├── mocks/
    ├── configs/
    ├── uploads/
    ├── exports/
    └── config.json
```

### User Configuration Example
```json
{
  "user_id": "project-alpha",
  "created_at": "2024-01-15T10:30:00Z",
  "settings": {
    "theme": "dark",
    "auto_save": true,
    "auto_backup": true
  }
}
```

## 🐛 Troubleshooting

### Mock Won't Start
**Problem**: Mock shows as created but won't start.

**Solution**: 
1. Check port conflicts in the dashboard
2. Verify protocol is supported
3. Check server logs for detailed errors
4. Ensure you're working with user mocks, not example mocks

### Cannot Delete Mock
**Problem**: Getting "403 Forbidden" when trying to delete a mock.

**Solution**: 
1. You're trying to delete an example mock (protected)
2. Example mocks are read-only templates
3. Only user-created mocks can be deleted
4. Look for mocks with "User" badge instead of "Example" badge

### SFTP Connection Fails
**Problem**: Cannot connect to SFTP server.

**Solutions**:
1. Ensure SFTP mock is running
2. Verify port 2222 is not blocked
3. Use correct credentials: usekuro/kuro123
4. Try: `sftp -o StrictHostKeyChecking=no -P 2222 usekuro@localhost`

### Workspace Creation Fails
**Problem**: Cannot create new workspace.

**Solutions**:
1. Check workspace name is unique
2. Ensure write permissions in workspaces directory
3. Verify disk space is available
4. Remember: deleting workspace only removes YOUR mocks, not examples

### Port Already in Use
**Problem**: "Port X is already in use" error.

**Solutions**:
1. Stop conflicting mock in dashboard
2. Choose different port number
3. Check system for other services using the port: `lsof -i :PORT`

## 📚 API Reference

### Mock Management
```bash
# Create mock
POST /api/user/{userID}/mocks
Content-Type: application/json
{
  "name": "My Mock",
  "protocol": "http",
  "port": 8081,
  "description": "Test mock"
}

# List user mocks
GET /api/user/{userID}/mocks

# Delete mock
DELETE /api/user/{userID}/mocks/{mockID}
```

### Workspace Management
```bash
# List workspaces
GET /api/workspaces

# Create workspace
POST /api/user/{userID}/workspace

# Delete workspace
DELETE /api/user/{userID}/workspace
```

### Server Control
```bash
# Health check
GET /health

# Get public config
GET /api/config

# Get user stats
GET /api/user/{userID}/stats
```

## 🚀 Next Steps

1. **Explore Examples**: Check the `examples/` directory for more complex configurations
2. **Read Documentation**: Review `IMPLEMENTATION_SUMMARY.md` for technical details
3. **Customize Mocks**: Edit generated .kuro files for advanced configurations
4. **Integration**: Use mocks in your development workflow
5. **Contribute**: Submit issues and improvements to the project

## 💡 Tips and Best Practices

### Mock Organization
- Use descriptive names for user mocks
- Group related mocks in dedicated workspaces
- Include meaningful descriptions
- Use consistent port numbering schemes
- Reference example mocks as templates for common patterns
- Keep user mocks separate from examples for clarity

### Development Workflow
1. Create workspace per project
2. Review example mocks for template ideas
3. Create user mocks for all external dependencies
4. Start mocks before running tests
5. Export mock configurations for team sharing
6. Delete workspace when project ends (examples remain available)

### Performance
- Stop unused mocks to free resources
- Use appropriate ports (avoid system reserved ports)
- Monitor server resources with many concurrent mocks

### Security
- Change default SFTP credentials in production
- Use HTTPS for sensitive mock data
- Restrict network access as needed

## 📞 Support

If you encounter issues:

1. **Check Logs**: Server logs provide detailed error information
2. **Run Tests**: Use `./test_implementation.sh` to verify setup
3. **Review Examples**: Compare with working examples in `examples/`
4. **Documentation**: Check `IMPLEMENTATION_SUMMARY.md` for technical details

## 🎉 Success!

You now have a fully functional UseKuro setup with:
- ✅ Automatic mock definition generation
- ✅ Smart workspace management with example preservation
- ✅ Protected example mocks as permanent templates
- ✅ Clear separation between user and example mocks
- ✅ Multi-protocol support
- ✅ SFTP server functionality
- ✅ Intuitive web-based management interface

Happy mocking! 🚀