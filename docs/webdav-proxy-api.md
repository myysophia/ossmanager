# WebDAV Proxy RESTful API

This document describes the RESTful WebDAV proxy endpoints that wrap WebDAV operations and return JSON responses.

## Overview

The WebDAV proxy API provides RESTful endpoints that internally use WebDAV operations while maintaining consistent RBAC checks and returning JSON responses. This allows clients to interact with WebDAV storage using familiar REST patterns.

## Base URL

All WebDAV proxy endpoints are under:
```
/api/v1/webdav/objects
```

## Authentication

The API supports two authentication methods:

### 1. JWT Token Authentication
```http
Authorization: Bearer <jwt_token>
```

### 2. WebDAV Token Authentication
```http
X-WebDAV-Token: <webdav_token>
```

Alternative query parameter:
```
?webdav_token=<webdav_token>
```

## Endpoints

### 1. List Directory (PROPFIND)

**Endpoint:** `GET /api/v1/webdav/objects/{bucket}?path={directory_path}`

Maps to WebDAV PROPFIND operation.

**Parameters:**
- `bucket` (path): Target bucket name
- `path` (query): Directory path to list (default: `/`)

**Response:**
```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "items": [
      {
        "name": "file.txt",
        "path": "/documents/file.txt",
        "isDir": false,
        "size": 1024,
        "mtime": "2024-01-15T10:30:00Z",
        "mimeType": "text/plain"
      },
      {
        "name": "subfolder",
        "path": "/documents/subfolder/",
        "isDir": true,
        "size": 0,
        "mtime": "2024-01-15T09:00:00Z"
      }
    ],
    "path": "/documents",
    "total": 2
  }
}
```

**Example:**
```bash
curl -H "Authorization: Bearer <token>" \
  "https://api.example.com/api/v1/webdav/objects/my-bucket?path=/documents"
```

### 2. Upload File (PUT)

**Endpoint:** `POST /api/v1/webdav/objects/{bucket}/file?path={file_path}`

Maps to WebDAV PUT operation.

**Parameters:**
- `bucket` (path): Target bucket name  
- `path` (query): File path for upload

**Request Body:** File content (binary)

**Response:**
```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "success": true,
    "message": "file uploaded successfully",
    "path": "/documents/uploaded-file.txt"
  }
}
```

**Example:**
```bash
curl -X POST \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/octet-stream" \
  --data-binary @local-file.txt \
  "https://api.example.com/api/v1/webdav/objects/my-bucket/file?path=/documents/uploaded-file.txt"
```

### 3. Delete File/Folder (DELETE)

**Endpoint:** `DELETE /api/v1/webdav/objects/{bucket}?path={target_path}`

Maps to WebDAV DELETE operation.

**Parameters:**
- `bucket` (path): Target bucket name
- `path` (query): Path of file or folder to delete

**Response:**
```json
{
  "code": 200,
  "message": "操作成功", 
  "data": {
    "success": true,
    "message": "file or directory deleted successfully",
    "path": "/documents/file-to-delete.txt"
  }
}
```

**Example:**
```bash
curl -X DELETE \
  -H "Authorization: Bearer <token>" \
  "https://api.example.com/api/v1/webdav/objects/my-bucket?path=/documents/file-to-delete.txt"
```

### 4. Rename/Move (MOVE)

**Endpoint:** `PATCH /api/v1/webdav/objects/{bucket}/rename`

Maps to WebDAV MOVE operation.

**Parameters:**
- `bucket` (path): Target bucket name

**Request Body:**
```json
{
  "oldPath": "/documents/old-name.txt",
  "newPath": "/documents/new-name.txt"
}
```

**Response:**
```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "success": true,
    "message": "renamed from /documents/old-name.txt to /documents/new-name.txt successfully",
    "path": "/documents/new-name.txt"
  }
}
```

**Example:**
```bash
curl -X PATCH \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"oldPath":"/documents/old-name.txt","newPath":"/documents/new-name.txt"}' \
  "https://api.example.com/api/v1/webdav/objects/my-bucket/rename"
```

### 5. Create Directory (MKCOL)

**Endpoint:** `POST /api/v1/webdav/objects/{bucket}/mkdir`

Maps to WebDAV MKCOL operation.

**Parameters:**
- `bucket` (path): Target bucket name

**Request Body:**
```json
{
  "path": "/documents/new-folder"
}
```

**Response:**
```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "success": true,
    "message": "directory created successfully",
    "path": "/documents/new-folder"
  }
}
```

**Example:**
```bash
curl -X POST \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"path":"/documents/new-folder"}' \
  "https://api.example.com/api/v1/webdav/objects/my-bucket/mkdir"
```

## Error Responses

All endpoints return consistent error responses:

```json
{
  "code": 400,
  "message": "invalid bucket name"
}
```

Common error codes:
- `400` - Invalid parameters or request body
- `401` - Authentication required or invalid token
- `403` - Access denied to bucket or insufficient permissions  
- `404` - Resource not found
- `500` - Internal server error

## Security

### Path Validation
All paths are validated using WebDAV path validation to prevent:
- Directory traversal attacks (`../`)
- Invalid characters in bucket names
- Malformed paths

### Permission Checks
The API enforces the same RBAC permissions as the native WebDAV interface:
- **Read operations** (GET): Require `webdav:read` permission
- **Write operations** (POST, PATCH): Require `webdav:write` permission  
- **Delete operations** (DELETE): Require `webdav:delete` permission

### Bucket Access
Users must have explicit access to the target bucket through role-based permissions.

## Integration with WebDAV

The proxy API internally uses the existing WebDAV filesystem implementation:
- Preserves all WebDAV semantics and behavior
- Maintains transaction consistency
- Uses the same storage backend
- Applies identical permission checks
- Supports the same file operations

This ensures that operations performed through the REST API are fully compatible with native WebDAV clients.
