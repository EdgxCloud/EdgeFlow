# Cloud Storage & File Sharing Nodes

**Category**: Storage & External Services
**Status**: NEW - Just Implemented
**Total Nodes**: 3 (Google Drive, AWS S3, FTP)

---

## 📦 Implemented Storage Nodes

### 1. Google Drive Node ✅ NEW

**Location**: `pkg/nodes/storage/google_drive.go`

**Description**: Complete Google Drive integration for file storage and sharing

**Operations Supported**:
- `upload` - Upload files to Google Drive
- `download` - Download files from Google Drive
- `list` - List files and folders
- `delete` - Delete files
- `share` - Create shareable links with permissions
- `createFolder` - Create folders
- `getMetadata` - Get file metadata

**Configuration**:
```json
{
  "credentials": "Google OAuth2 JSON credentials",
  "token": "OAuth2 access token (JSON)",
  "folderId": "Optional default folder ID"
}
```

**Use Cases**:
- Backup IoT sensor data to cloud
- Share device logs with team
- Store configuration files
- Archive time-series data
- Generate shareable reports

**Example Usage**:
```json
// Upload a file
{
  "operation": "upload",
  "filePath": "/data/sensor_readings.csv",
  "fileName": "readings_2026-01-22.csv",
  "parentId": "folder_id_here"
}

// Create shareable link
{
  "operation": "share",
  "fileId": "file_id_here",
  "type": "anyone",
  "role": "reader"
}

// List files in folder
{
  "operation": "list",
  "query": "'folder_id' in parents",
  "pageSize": 50
}
```

**Authentication**:
1. Create Google Cloud Project
2. Enable Drive API
3. Create OAuth2 credentials
4. Get access token (refresh token for long-term use)

**Status**: Production ready

---

### 2. AWS S3 Node ✅ NEW

**Location**: `pkg/nodes/storage/aws_s3.go`

**Description**: Amazon S3 object storage integration

**Operations Supported**:
- `upload` - Upload files to S3 bucket
- `download` - Download files from S3
- `list` - List objects in bucket
- `delete` - Delete objects
- `getSignedUrl` - Generate presigned URLs for sharing
- `copy` - Copy objects within S3
- `getMetadata` - Get object metadata

**Configuration**:
```json
{
  "region": "us-east-1",
  "accessKey": "AWS Access Key ID",
  "secretKey": "AWS Secret Access Key",
  "bucket": "my-bucket-name",
  "prefix": "optional/path/prefix"
}
```

**Use Cases**:
- Archive large sensor datasets
- Store images from cameras
- Backup system logs
- Distribute firmware updates
- Host static files

**Example Usage**:
```json
// Upload file
{
  "operation": "upload",
  "filePath": "/logs/system.log",
  "key": "logs/2026/01/system.log",
  "contentType": "text/plain",
  "acl": "private"
}

// Generate presigned URL (for sharing)
{
  "operation": "getSignedUrl",
  "key": "data/report.pdf",
  "expirationMinutes": 60
}

// List objects
{
  "operation": "list",
  "prefix": "sensors/temperature/",
  "maxKeys": 100
}
```

**Features**:
- Server-side encryption support
- ACL control (private, public-read, etc.)
- Presigned URLs with expiration
- Prefix-based organization
- Large file support

**Status**: Production ready

---

### 3. FTP Node ✅ NEW

**Location**: `pkg/nodes/storage/ftp.go`

**Description**: Traditional FTP file transfer

**Operations Supported**:
- `upload` - Upload files to FTP server
- `download` - Download files from FTP server
- `list` - List directory contents
- `delete` - Delete files
- `mkdir` - Create directories
- `rename` - Rename files/directories

**Configuration**:
```json
{
  "host": "ftp.example.com",
  "port": 21,
  "username": "ftpuser",
  "password": "ftppass"
}
```

**Use Cases**:
- Legacy system integration
- File transfer to/from NAS
- Backup to FTP server
- Integration with older equipment
- Simple file sharing

**Example Usage**:
```json
// Upload file
{
  "operation": "upload",
  "localPath": "/data/readings.csv",
  "remotePath": "/uploads/readings.csv"
}

// Download file
{
  "operation": "download",
  "remotePath": "/config/device.conf",
  "localPath": "/etc/device.conf"
}

// List directory
{
  "operation": "list",
  "path": "/uploads"
}
```

**Status**: Production ready

---

## 🔄 Planned Storage Nodes

### 4. Dropbox Node ⏳ PLANNED

**Description**: Dropbox file storage and sharing

**Priority**: HIGH
**Effort**: 2 days

**Operations**:
- Upload/download files
- Create shared links
- List files/folders
- Move/copy files
- Delete files

**Dependencies**: Dropbox SDK for Go

---

### 5. Microsoft OneDrive Node ⏳ PLANNED

**Description**: OneDrive cloud storage integration

**Priority**: HIGH
**Effort**: 2 days

**Operations**:
- Upload/download files
- Share files with permissions
- List files/folders
- Create folders
- Manage sharing links

**Dependencies**: Microsoft Graph API

---

### 6. SFTP Node ⏳ PLANNED

**Description**: Secure FTP with SSH encryption

**Priority**: MEDIUM
**Effort**: 1 day

**Operations**:
- Upload/download files (encrypted)
- List directories
- SSH key authentication
- Set file permissions
- Manage symlinks

**Dependencies**: SSH library for Go

---

### 7. Box Node ⏳ PLANNED

**Description**: Box.com enterprise file sharing

**Priority**: MEDIUM
**Effort**: 2 days

**Operations**:
- Upload/download files
- Share files
- Version management
- Permissions control

---

### 8. WebDAV Node ⏳ PLANNED

**Description**: WebDAV protocol support

**Priority**: LOW
**Effort**: 2 days

**Use Cases**:
- OwnCloud/Nextcloud integration
- Various cloud storage services

---

## 📊 Storage Nodes Summary

### Current Status:
```
Implemented: 3/8 (38%)
- ✅ Google Drive
- ✅ AWS S3
- ✅ FTP
- ⏳ Dropbox
- ⏳ OneDrive
- ⏳ SFTP
- ⏳ Box
- ⏳ WebDAV
```

### By Priority:
- **HIGH**: Dropbox, OneDrive (4 days total)
- **MEDIUM**: SFTP, Box (3 days total)
- **LOW**: WebDAV (2 days)

**Total estimated effort for remaining**: ~9 days

---

## 🎯 Use Case Examples

### IoT Data Backup Workflow:
```
Schedule (daily)
  → Read sensor data from SQLite
  → Function (format as CSV)
  → AWS S3 (upload with timestamp)
  → Email (send confirmation)
```

### Cloud Report Generation:
```
InfluxDB (query last 24h data)
  → Function (generate PDF report)
  → Google Drive (upload)
  → Google Drive (share with team)
  → Telegram (send link)
```

### Multi-Cloud Backup:
```
File In (read backup file)
  → AWS S3 (primary backup)
  → Google Drive (secondary backup)
  → FTP (tertiary backup to NAS)
  → Debug (log results)
```

### Firmware Distribution:
```
HTTP In (webhook on new firmware)
  → AWS S3 (upload firmware)
  → AWS S3 (getSignedUrl)
  → MQTT Out (broadcast download link to devices)
```

---

## 🔐 Security Best Practices

### 1. Credential Management:
- ✅ Never hardcode credentials
- ✅ Use environment variables
- ✅ Store OAuth tokens securely
- ✅ Rotate access keys regularly
- ✅ Use IAM roles when possible (AWS)

### 2. Access Control:
- ✅ Minimum necessary permissions
- ✅ Use presigned URLs for temporary access
- ✅ Set appropriate ACLs (S3)
- ✅ Implement sharing expiration
- ✅ Audit file access logs

### 3. Data Protection:
- ✅ Enable encryption at rest (S3)
- ✅ Use HTTPS/TLS for all transfers
- ✅ Validate file checksums
- ✅ Implement virus scanning
- ✅ Backup critical data to multiple locations

---

## 📦 Dependencies Added

```go
require (
    golang.org/x/oauth2            // OAuth2 authentication
    google.golang.org/api/drive/v3 // Google Drive API
    github.com/aws/aws-sdk-go/aws  // AWS SDK
    github.com/aws/aws-sdk-go/service/s3 // S3 service
    github.com/jlaffaye/ftp        // FTP client
)
```

---

## 🚀 Quick Start

### Google Drive Setup:

1. **Create Google Cloud Project**
   - Go to console.cloud.google.com
   - Create new project
   - Enable Drive API

2. **Create OAuth2 Credentials**
   - Create OAuth 2.0 Client ID
   - Download JSON credentials

3. **Get Access Token**
   - Use OAuth2 flow to get token
   - Store refresh token for long-term use

4. **Configure Node**:
   ```json
   {
     "credentials": "<paste credentials JSON>",
     "token": "<paste token JSON>"
   }
   ```

### AWS S3 Setup:

1. **Create IAM User**
   - AWS Console → IAM
   - Create user with programmatic access

2. **Attach S3 Policy**
   - Attach `AmazonS3FullAccess` or custom policy

3. **Get Credentials**
   - Save Access Key ID and Secret Access Key

4. **Configure Node**:
   ```json
   {
     "region": "us-east-1",
     "accessKey": "AKIAIOSFODNN7EXAMPLE",
     "secretKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
     "bucket": "my-iot-data"
   }
   ```

### FTP Setup:

1. **Configure FTP Server**
   - Ensure FTP service is running
   - Create user account

2. **Configure Node**:
   ```json
   {
     "host": "192.168.1.100",
     "port": 21,
     "username": "ftpuser",
     "password": "secure_password"
   }
   ```

---

## 📝 Common Issues & Solutions

### Google Drive:

**Issue**: "Invalid credentials"
**Solution**: Ensure credentials JSON is valid and Drive API is enabled

**Issue**: "Token expired"
**Solution**: Refresh the OAuth2 token using refresh token

### AWS S3:

**Issue**: "Access Denied"
**Solution**: Check IAM permissions and bucket policy

**Issue**: "Bucket not found"
**Solution**: Verify bucket name and region

### FTP:

**Issue**: "Connection timeout"
**Solution**: Check firewall rules, ensure port 21 is open

**Issue**: "Login failed"
**Solution**: Verify username and password

---

## 🎓 Advanced Features

### Google Drive:
- Shared drive support
- File versioning
- Team folder management
- Real-time collaboration

### AWS S3:
- Server-side encryption (SSE-S3, SSE-KMS)
- Lifecycle policies
- Versioning
- Cross-region replication
- S3 Transfer Acceleration

### FTP:
- Passive/Active mode
- Custom timeout settings
- Binary/ASCII transfer modes
- Resume interrupted transfers

---

**Last Updated**: 2026-01-22
**Status**: 3 nodes production ready, 5 planned
**Next**: Implement Dropbox and OneDrive nodes
