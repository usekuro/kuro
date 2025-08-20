# UseKuro Mock Deletion Guide

## 🎯 Overview

This guide demonstrates how mock deletion works in UseKuro, with clear distinctions between example mocks (protected) and user mocks (deletable).

## 🛡️ Mock Types & Deletion Rules

### 📚 **Example Mocks** (Protected)
- **Source**: Built-in templates and examples
- **Location**: "📚 Example Mocks" section
- **Badge**: 🔵 Blue "Example" badge
- **Deletion**: ❌ **NOT ALLOWED** (403 Forbidden)
- **Purpose**: Permanent templates for reference
- **Controls**: Only 📤 Export button available

### 👤 **User Mocks** (Deletable)
- **Source**: Created by users
- **Location**: "👤 Your Mocks" section  
- **Badge**: 🟢 Green "User" badge
- **Deletion**: ✅ **ALLOWED** (Individual deletion)
- **Purpose**: User's working mocks
- **Controls**: ✏️ Edit, 📤 Export, 🗑️ **Delete** buttons

## 🔄 Individual Mock Deletion Process

### ✅ Deleting User Mocks (Step by Step)

```
1. Navigate to User Workspace tab
2. Locate "👤 Your Mocks" section
3. Find the mock you want to delete
4. Click the 🗑️ (trash) button
5. Confirm deletion in popup
6. Mock is removed immediately
```

**API Call**: `DELETE /api/user/{userID}/mocks/{mockID}`
**Response**: `200 OK - "User mock deleted successfully"`

### ❌ Attempting to Delete Example Mocks

```
1. Navigate to User Workspace tab
2. Locate "📚 Example Mocks" section
3. Notice: NO 🗑️ button available
4. Only 📤 Export button shown
5. Examples cannot be deleted via UI
```

**If attempted via API**: `DELETE /api/user/default/mocks/{exampleID}`
**Response**: `403 Forbidden - "Cannot delete example mocks. Only user-created mocks can be deleted."`

## 🎨 Visual Interface Guide

### Example Mocks Section
```
📚 Example Mocks (Read-only templates)
┌─────────────────────────────────────────────────────────┐
│ [HTTP] Sample API Server        :8080  [Example]        │
│ Basic REST API for testing               [📤 Export]    │
│                                                         │
│ [SFTP] File Server             :2222  [Example]        │
│ SFTP server with authentication          [📤 Export]    │
└─────────────────────────────────────────────────────────┘
```

### User Mocks Section
```
👤 Your Mocks (2 total)
┌─────────────────────────────────────────────────────────┐
│ [HTTP] My API Mock             :8081  [User]            │
│ My development API              [✏️] [📤] [🗑️]        │
│                                                         │
│ [SFTP] Project Files           :2223  [User]            │
│ File server for my project      [✏️] [📤] [🗑️]        │
└─────────────────────────────────────────────────────────┘
```

## 🧪 Testing Individual Mock Deletion

### Test Case 1: Delete User Mock ✅
```bash
# Create a test user mock
curl -X POST http://localhost:8080/api/user/testuser/mocks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Mock",
    "protocol": "http", 
    "port": 8082,
    "description": "Mock for deletion test"
  }'

# Response: 201 Created
# Mock ID: user_testuser_1234567890

# Delete the user mock
curl -X DELETE http://localhost:8080/api/user/testuser/mocks/user_testuser_1234567890

# Response: 200 OK
# {"success": true, "message": "User mock deleted successfully"}
```

### Test Case 2: Attempt to Delete Example Mock ❌
```bash
# Try to delete an example mock
curl -X DELETE http://localhost:8080/api/user/default/mocks/example_http_basic

# Response: 403 Forbidden
# {"error": "Cannot delete example mocks. Only user-created mocks can be deleted."}
```

## 📋 Deletion Scenarios by Workspace

### Scenario 1: Default Workspace
```
Workspace: default
├── 📚 Example Mocks (3 items)
│   ├── HTTP API Example        ❌ Cannot delete
│   ├── SFTP Server Example     ❌ Cannot delete  
│   └── TCP Echo Example        ❌ Cannot delete
└── 👤 Your Mocks (0 items)
    └── (No user mocks yet)     ℹ️ Nothing to delete
```

### Scenario 2: User Workspace with Mixed Content
```
Workspace: project-alpha
├── 📚 Example Mocks (3 items)
│   ├── HTTP API Example        ❌ Cannot delete (protected)
│   ├── SFTP Server Example     ❌ Cannot delete (protected)
│   └── TCP Echo Example        ❌ Cannot delete (protected)
└── 👤 Your Mocks (4 items)
    ├── Auth Service Mock       ✅ Can delete individually
    ├── Database Mock           ✅ Can delete individually
    ├── File Storage Mock       ✅ Can delete individually
    └── Email Service Mock      ✅ Can delete individually
```

### Scenario 3: After Individual Deletions
```
Workspace: project-alpha
├── 📚 Example Mocks (3 items)
│   ├── HTTP API Example        ❌ Still protected
│   ├── SFTP Server Example     ❌ Still protected
│   └── TCP Echo Example        ❌ Still protected
└── 👤 Your Mocks (2 items)     ← Reduced count
    ├── Auth Service Mock       ✅ Can still delete
    └── Email Service Mock      ✅ Can still delete
    # Database Mock - DELETED ✅
    # File Storage Mock - DELETED ✅
```

## 🔍 Identifying Mock Types

### Quick Visual Identification
| Feature | Example Mocks | User Mocks |
|---------|---------------|------------|
| **Section Header** | 📚 Example Mocks | 👤 Your Mocks |
| **Badge Color** | 🔵 Blue "Example" | 🟢 Green "User" |
| **Border** | Blue glow border | Standard border |
| **Edit Button** | ❌ Not available | ✅ Available |
| **Delete Button** | ❌ Not available | ✅ Available |
| **Export Button** | ✅ Available | ✅ Available |

### Technical Identification
```javascript
// Mock ID patterns
Example Mock: "example_http_basic" or "sftp_server_demo"
User Mock:    "user_john_1234567890" or "user_project_1234567890"

// API Response Structure
{
  "mocks": [...],           // User mocks only
  "example_mocks": [...],   // Example mocks only
  "total_user": 2,          // Count of user mocks
  "total_examples": 3       // Count of example mocks
}
```

## ⚡ Quick Actions

### Delete All User Mocks in Workspace
```bash
# Get user mocks
curl http://localhost:8080/api/user/myproject/mocks

# Delete each user mock individually
for mockId in $(curl -s http://localhost:8080/api/user/myproject/mocks | jq -r '.mocks[].id'); do
  curl -X DELETE http://localhost:8080/api/user/myproject/mocks/$mockId
done
```

### Delete Entire User Workspace (All User Mocks)
```bash
# This deletes ALL user mocks in the workspace but preserves examples
curl -X DELETE http://localhost:8080/api/user/myproject/workspace

# Result: 
# ✅ All user mocks in 'myproject' workspace deleted
# ✅ Example mocks preserved and still available
# ✅ User can create new workspace with same name
```

## 🛡️ Safety Features

### Protection Mechanisms
1. **API-Level Protection**: 403 Forbidden for example mock deletion
2. **UI-Level Protection**: No delete button shown for example mocks  
3. **ID-Based Validation**: Only `user_` prefixed mocks can be deleted
4. **Workspace Isolation**: Users can only delete their own mocks
5. **Confirmation Dialogs**: User must confirm individual deletions

### Recovery Options
1. **Individual Recovery**: Re-create deleted user mocks manually
2. **Template Recovery**: Use example mocks as templates
3. **Workspace Recovery**: Create new workspace (examples available)
4. **Backup Strategy**: Export user mocks before deletion

## 📞 Troubleshooting

### "Cannot Delete Mock" Error
**Problem**: Getting 403 Forbidden when trying to delete
**Cause**: Attempting to delete an example mock
**Solution**: Only delete mocks with green "User" badge

### "Mock Not Found" Error  
**Problem**: 404 error when deleting mock
**Cause**: Mock ID doesn't exist or wrong workspace
**Solution**: Verify mock exists in current workspace

### Delete Button Missing
**Problem**: No 🗑️ button visible
**Cause**: Looking at example mocks section
**Solution**: Use "👤 Your Mocks" section for deletable mocks

## ✅ Summary

**Individual Mock Deletion in UseKuro:**

✅ **User Mocks**: Can be deleted individually anytime  
✅ **Granular Control**: Delete specific mocks without affecting others  
✅ **Workspace Isolation**: Only affects current workspace  
✅ **Example Protection**: Templates always preserved  
✅ **UI Clarity**: Visual indicators show what can be deleted  
✅ **API Safety**: Backend enforces deletion rules  

**Perfect for iterative development where you need to clean up specific mocks while keeping others running!** 🚀