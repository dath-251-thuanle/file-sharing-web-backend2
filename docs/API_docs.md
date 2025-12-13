# API Documentation
## Table of Contents
- [API Specification](#api-specification)
- [API Overview](#api-overview)
- [Endpoints Summary](#endpoints-summary)
- [Response Codes](#response-codes)
- [Database Tables](#database-tables)
- [Local Storage](#local-storage)
- [TOTP/2FA Flow](#totp2fa-flow)
- [File Statistics & Analytics](#file-statistics--analytics)
- [File Status](#file-status)
- [Validity Period Logic](#validity-period-logic)
- [Security](#security)
- [Download Access Control](#download-access-control)
- [Quick Reference](#quick-reference)
---
## API Specification
Project sử dụng **OpenAPI 3.0.4** để định nghĩa API:
### Xem Documentation
#### Online Tools
1. **Swagger Editor**: https://editor.swagger.io/
   - Copy nội dung `openapi.yaml` vào editor
   - Xem live preview và validate
2. **Postman**:
   - Import file `openapi.yaml`
   - Tự động generate API collection

---
## API Overview
### Base URL
- Development: `http://localhost:8080`
- Production: `https://api.filesharing-hcmut.com`
### Authentication
- Type: Bearer Token (JWT)
- Header: `Authorization: Bearer <token>`
---
## Endpoints Summary
### Authentication
| Method | Endpoint | Mô tả | Auth |
|--------|----------|-------|------|
| `POST` | `/auth/register` | Đăng ký tài khoản mới | ❌ |
| `POST` | `/auth/login` | Đăng nhập (trả về token hoặc yêu cầu TOTP) | ❌ |
| `POST` | `/auth/login/totp` | Xác thực TOTP để hoàn tất đăng nhập | ❌ |
| `POST` | `/auth/totp/setup` | Thiết lập TOTP cho user | ✅ Bearer |
| `POST` | `/auth/totp/verify` | Xác minh mã TOTP để kích hoạt 2FA | ✅ Bearer |
| `POST` | `/auth/logout` | Đăng xuất | ✅ Bearer |
| `GET` | `/user` | Lấy thông tin profile user hiện tại | ✅ Bearer |
### Files
| Method | Endpoint | Mô tả | Auth |
|--------|----------|-------|------|
| `POST` | `/files/upload` | Upload file | Optional |
| `GET` | `/files/my` | Lấy danh sách file do user hiện tại upload | ✅ Bearer |
| `GET` | `/files/available` | Lấy danh sách file có thể tải về | Optional |
| `GET` | `/files/info/{id}` | Lấy thông tin file theo UUID (chỉ owner/admin) | ✅ Bearer |
| `DELETE` | `/files/info/{id}` | Xóa file (chỉ owner/admin) | ✅ Bearer |
| `GET` | `/files/stats/{id}` | Lấy thống kê download của file (chỉ owner/admin) | ✅ Bearer |
| `GET` | `/files/download-history/{id}` | Lấy lịch sử download chi tiết (chỉ owner/admin) | ✅ Bearer |
| `GET` | `/files/{shareToken}` | Lấy thông tin file qua share token (public) | ❌ |
| `GET` | `/files/{shareToken}/download` | Tải file về (hỗ trợ password) | Optional |
| `GET` | `/files/{shareToken}/preview` | Xem trước file trong browser (inline display) | Optional |
### Admin
| Method | Endpoint | Mô tả | Auth |
|--------|----------|-------|------|
| `POST` | `/admin/cleanup` | Xóa file hết hạn | ✅ Admin/Cron |
| `GET` | `/admin/policy` | Lấy cấu hình hệ thống | ✅ Admin |
| `PATCH` | `/admin/policy` | Cập nhật cấu hình | ✅ Admin |
---
## Response Codes
| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Success |
| 201 | Created | Upload thành công |
| 400 | Bad Request | Validation error / Invalid token |
| 401 | Unauthorized | Cần đăng nhập / Token expired |
| 403 | Forbidden | Không có quyền / Wrong password |
| 404 | Not Found | Không tìm thấy resource |
| 409 | Conflict | Email/username đã tồn tại |
| 410 | Gone | File đã hết hạn |
| 413 | Payload Too Large | File quá lớn |
| 423 | Locked | File chưa đến thời gian hiệu lực |
| 429 | Too Many Requests | Vượt quá rate limit (cleanup endpoint) |
---
## Database Tables
Project sử dụng PostgreSQL với các bảng được khởi tạo qua Docker Compose (mount file `init.sql`).
| Table | Description | Key Features |
|-------|-------------|--------------|
| `users` | User accounts | TOTP support (`enableTOTP`, `secretTOTP`), roles (user/admin) |
| `files` | Uploaded files metadata | Share tokens, password, validity period, public/private |
| `filestat` | Aggregated download stats | `download_count`, `user_download_count` |
| `shared` | File sharing relationships | Many-to-many: user_id ↔ file_id |
| `download` | Download history log | Audit trail, user tracking |
| `jwt_blacklist` | Revoked JWT tokens | Token invalidation |
| `usersLoginSession` | TOTP login sessions | Challenge ID (`cid`) for 2FA flow |
**Schema:** Xem `internal/infrastructure/database/init.sql`
### Database Schema Details
```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    role VARCHAR(50) NOT NULL,
    enableTOTP BOOLEAN DEFAULT FALSE,
    secretTOTP VARCHAR(255)
);
-- Files table
CREATE TABLE files (
    id UUID PRIMARY KEY,
    user_id UUID,                    -- NULL cho anonymous upload
    name VARCHAR(255) NOT NULL,
    password VARCHAR(255),           -- Password protection (min 6 chars)
    type TEXT,                       -- MIME type
    size BIGINT,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now(),
    available_from TIMESTAMPTZ,
    available_to TIMESTAMPTZ,
    enable_totp BOOLEAN DEFAULT false,
    share_token TEXT
);
-- File statistics
CREATE TABLE filestat (
    file_id UUID NOT NULL,
    download_count BIGINT DEFAULT 0,
    user_download_count BIGINT DEFAULT 0
);
-- Shared files (whitelist)
CREATE TABLE shared (
    user_id UUID NOT NULL,
    file_id UUID NOT NULL,
    PRIMARY KEY (user_id, file_id)
);
-- Download history
CREATE TABLE download (
    download_id UUID PRIMARY KEY,
    time TIMESTAMPTZ DEFAULT now(),
    user_id UUID,                    -- NULL cho anonymous download
    file_id UUID NOT NULL
);
```
### Stored Procedure
```sql
-- Procedure để ghi nhận download
CREATE PROCEDURE proc_download(f_id UUID, u_id UUID)
-- Tự động:
-- 1. Tăng download_count
-- 2. Tăng user_download_count (nếu user chưa download file này)
-- 3. Ghi log vào bảng download
```
---
## Local Storage
Backend sử dụng **Local File Storage** để lưu trữ file:
| Config | Value |
|--------|-------|
| **Provider** | Local filesystem |
| **Storage Path** | `uploads/` (relative to working directory) |
| **Max File Size** | 50MB (configurable via policy) |
**Lưu ý:**
- File được lưu với tên ngẫu nhiên (`storage_name`) để tránh trùng lặp
- Tên gốc (`name`) được lưu trong database để hiển thị cho user
- Khi xóa file, cả record trong DB và file trên disk đều bị xóa
---
## TOTP/2FA Flow
### User TOTP (2FA for Account Login)
**Luồng bật TOTP:**
```
1. User đăng ký: POST /auth/register
   ↓
2. User đăng nhập lần đầu: POST /auth/login 
   → Nhận accessToken
   ↓
3. User muốn bật 2FA: POST /auth/totp/setup (cần Bearer token)
   → Nhận secret + qrCode (base64 PNG)
   ↓
4. User quét QR code bằng Google Authenticator/Authy
   ↓
5. User xác minh mã: POST /auth/totp/verify (cần Bearer token)
   → Tài khoản được đánh dấu totpEnabled=true
```
**Luồng đăng nhập với TOTP:**
```
1. User nhập email/password: POST /auth/login
   → Response: { requireTOTP: true, cid: "xxx" }
   ↓
2. User nhập mã 6 số từ app: POST /auth/login/totp
   Body: { cid: "xxx", code: "123456" }
   → Nhận accessToken
```
**Bảng liên quan:** `usersLoginSession` lưu `cid` tạm thời cho phiên đăng nhập TOTP
---
## File Statistics & Analytics
### GET /files/stats/{id}
Lấy thống kê download của file (chỉ owner/admin).
**Dữ liệu trả về:**
| Field | Description |
|-------|-------------|
| `downloadCount` | Tổng số lượt download |
| `uniqueDownloaders` | Số người download khác nhau (authenticated users only) |
| `lastDownloadedAt` | Thời điểm download gần nhất |
**Source:** Bảng `filestat`
**Note:** Anonymous uploads không có statistics
### GET /files/download-history/{id}
Lấy lịch sử download chi tiết (chỉ owner/admin).
**Dữ liệu trả về:**
| Field | Description |
|-------|-------------|
| `history[].id` | Download ID |
| `history[].downloader` | User info (null nếu anonymous) |
| `history[].downloadedAt` | Timestamp |
| `history[].downloadCompleted` | Trạng thái hoàn thành |
**Pagination:** `?page=1&limit=50`
**Source:** Bảng `download`
**Privacy:** Anonymous download chỉ ghi nhận timestamp, không log IP/User-Agent
---
## File Status
| Status | Description |
|--------|-------------|
| `pending` | Chưa đến thời gian `availableFrom` (owner có thể preview bằng JWT, người khác nhận 423) |
| `active` | Đang trong thời gian hiệu lực |
| `expired` | Đã hết hạn (`availableTo` đã qua) |
---
## Validity Period Logic
| Input | Result |
|-------|--------|
| FROM + TO | Hiệu lực từ FROM đến TO |
| Chỉ TO | Hiệu lực từ hiện tại đến TO |
| Chỉ FROM | Hiệu lực từ FROM đến FROM + 7 ngày |
| Không có | Hiệu lực từ hiện tại đến +7 ngày |
**Validation bổ sung:**
- `availableFrom` ≤ `availableTo`
- `availableTo` không nằm trong quá khứ tại thời điểm upload
- `availableTo` không vượt quá `maxValidityDays` (30 ngày mặc định)
- Vi phạm → backend trả lỗi `invalidValidityRange`
**System Policy (Default):**
| Policy | Value |
|--------|-------|
| `maxFileSizeMB` | 50 |
| `minValidityHours` | 1 |
| `maxValidityDays` | 30 |
| `defaultValidityDays` | 7 |
| `requirePasswordMinLength` | 6 |
Admin có thể thay đổi qua `PATCH /admin/policy`
---
## Security
### Bearer Token (JWT)
- **Lấy từ:** `POST /auth/login` hoặc `POST /auth/login/totp`
- **Format:** `Authorization: Bearer <token>`
- **Dùng cho:** Tất cả authenticated endpoints
### X-Cron-Secret
- Secret key cho cron job (lưu trong env)
- Dùng cho endpoint `/admin/cleanup`
- Header: `X-Cron-Secret: <secret>`
- Nên rotation định kỳ (30-60 ngày)
### X-File-Password
- Password để download file được bảo vệ
- Header: `X-File-Password: <password>`
- Dùng cho endpoint `/files/{shareToken}/download`
### CORS
```go
AllowOrigins:     []string{"http://localhost:3000"}
AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"}
AllowCredentials: true
```
---
## Download Access Control
Các endpoint tải file hỗ trợ nhiều lớp bảo mật đồng thời. Backend kiểm tra theo thứ tự:
```
1. File status
   ├── Hết hạn → 410 Gone
   └── Chưa đến giờ → 423 Locked
2. Whitelist (sharedWith)
   ├── Thiếu Bearer token → 401 Unauthorized
   └── User không trong whitelist → 403 Forbidden
3. Password
   ├── Thiếu password → 403 Forbidden
   └── Sai password → 403 Forbidden
4. ✅ Success → 200 OK (trả file binary)
```
### /files/{shareToken}/download
| HTTP Code | Case | Description |
|-----------|------|-------------|
| `200` | Success | Trả file binary |
| `401` | `missingAuth` | File private nhưng thiếu Bearer token |
| `403` | `wrongPassword` | Password sai |
| `403` | `missingPassword` | File có password nhưng không gửi |
| `403` | `notWhitelisted` | User không nằm trong danh sách chia sẻ |
| `404` | `notFound` | Share token không tồn tại |
| `410` | `expired` | File đã hết hạn |
| `423` | `pending` | File chưa đến thời gian hiệu lực |
**Owner preview:**
- Chủ file (JWT hợp lệ, `sub` = ownerId) có thể bypass trạng thái `pending` để kiểm thử link
- Người khác vẫn nhận `423` cho tới khi `availableFrom` đến
### /files/{shareToken}/preview
Xem trước file trong browser (inline display) thay vì tải về.
| Aspect | `/download` | `/preview` |
|--------|-------------|------------|
| **Header** | `Content-Disposition: attachment` | `Content-Disposition: inline` |
| **Hành vi** | Tải file về máy | Hiển thị trong browser |
| **Use case** | Download file | Xem PDF, hình ảnh, video trực tiếp |
**MIME types hỗ trợ preview:**
- Images: `image/jpeg`, `image/png`, `image/gif`, `image/webp`
- Documents: `application/pdf`
- Text: `text/plain`, `text/html`, `text/css`, `text/javascript`
- Video: `video/mp4`, `video/webm`
- Audio: `audio/mpeg`, `audio/wav`
**Lưu ý:** Các lớp bảo mật (status, whitelist, password) áp dụng giống endpoint `/download`
---
## Quick Reference
### Common Use Cases
#### 1. Anonymous Upload + Share
```bash
POST /files/upload
Content-Type: multipart/form-data
Body: file=@document.pdf
# Response
{
  "success": true,
  "file": {
    "id": "xxx",
    "shareToken": "a1b2c3d4e5f6g7h8"
  }
}
# Chia sẻ link
→ http://localhost:8080/files/a1b2c3d4e5f6g7h8/download
```
**Lưu ý:** Anonymous chỉ được upload file public, không đặt whitelist và không thể xóa sau khi upload.
#### 2. Upload với Password Protection
```bash
POST /files/upload
Content-Type: multipart/form-data
Body: 
  file=@secret.pdf
  password=secret123
# Người download cần header
X-File-Password: secret123
```
#### 3. Share với Whitelist
```bash
POST /files/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data
Body:
  file=@confidential.pdf
  isPublic=false
  sharedWith=["user1@gmail.com", "user2@gmail.com"]
# Chỉ user1 và user2 có thể download (cần đăng nhập)
```
#### 4. Owner Xem Ai Đã Download File
```bash
# Tổng quan
GET /files/stats/{fileId}
Authorization: Bearer <token>
# Chi tiết từng lượt download
GET /files/download-history/{fileId}?page=1&limit=50
Authorization: Bearer <token>
```
#### 5. Owner Xem Danh Sách File Của Mình
```bash
GET /files/my?status=all&page=1&limit=20&sortBy=createdAt&order=desc
Authorization: Bearer <token>
# Response
{
  "files": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "fileName": "document.pdf",
      "shareToken": "a1b2c3d4e5f6g7h8",
      "status": "active",
      "createdAt": "2025-11-19T10:00:00Z"
    }
  ],
  "pagination": { "currentPage": 1, "totalPages": 3, "totalFiles": 42 },
  "summary": { "activeFiles": 28, "pendingFiles": 5, "expiredFiles": 9 }
}
```
#### 6. Xem Các File Có Thể Tải Về
```bash
# Anonymous - chỉ xem file public
GET /files/available?page=1&limit=10
# Authenticated - xem file public + file được share cho mình
GET /files/available?page=1&limit=10
Authorization: Bearer <token>
```
#### 7. Download File Có Nhiều Lớp Bảo Mật
```bash
# File có: password + whitelist
# 1. Đăng nhập (để pass whitelist check)
POST /auth/login
→ Nhận accessToken
# 2. Download với password
GET /files/{shareToken}/download
Authorization: Bearer <token>
X-File-Password: secret123
```
### Docker Commands
```bash
# Khởi động tất cả services
docker compose up -d
# Xem logs
docker compose logs -f api
# Tạo bảng database (chạy lần đầu hoặc khi cần reset)
docker exec -i postgres-db psql -U <user> -d <dbname> < internal/infrastructure/database/init.sql
# Dừng services
docker compose down
# Dừng và xóa data
docker compose down -v
```
### Makefile Commands
```bash
make server        # Chạy server development
make docker-reset  # Reset database (xóa data + khởi động lại)
make docker-logs   # Xem logs API
make test          # Chạy tests
make clean         # Xóa build artifacts
make deps          # Tải dependencies
```
---
## API Response Examples
### Success Response
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": { ... }
}
```
### Error Response
```json
{
  "error": "Error type",
  "message": "Detailed error description"
}
```
### File Upload Response
```json
{
  "success": true,
  "message": "File uploaded successfully",
  "file": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "fileName": "report.pdf",
    "fileSize": 2048576,
    "mimeType": "application/pdf",
    "shareToken": "a1b2c3d4e5f6g7h8",
    "isPublic": true,
    "hasPassword": false,
    "availableFrom": "2025-11-10T00:00:00Z",
    "availableTo": "2025-11-17T00:00:00Z",
    "status": "active",
    "createdAt": "2025-11-10T10:00:00Z"
  }
}
```
### Login Response (với TOTP)
```json
{
  "requireTOTP": true,
  "message": "TOTP verification required",
  "cid": "8d4f3bb1-2f52-4a76-b951-7c21ef991abc"
}
```
### Download History Response
```json
{
  "fileId": "550e8400-e29b-41d4-a716-446655440000",
  "fileName": "presentation.pdf",
  "history": [
    {
      "id": "650e8400-e29b-41d4-a716-446655440001",
      "downloader": {
        "username": "tranthib",
        "email": "tranthib@example.com"
      },
      "downloadedAt": "2025-11-19T14:30:00Z",
      "downloadCompleted": true
    },
    {
      "id": "650e8400-e29b-41d4-a716-446655440002",
      "downloader": null,
      "downloadedAt": "2025-11-19T10:15:00Z",
      "downloadCompleted": true
    }
  ],
  "pagination": {
    "currentPage": 1,
    "totalPages": 1,
    "totalRecords": 2,
    "limit": 50
  }
}
```
---
> 📖 **Xem thêm:** 
> - [OpenAPI Specification](./openapi.yaml) - Chi tiết đầy đủ về tất cả endpoints và schemas
> - [Database Schema](../internal/infrastructure/database/init.sql) - SQL schema và stored procedures
