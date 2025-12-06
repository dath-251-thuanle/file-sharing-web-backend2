
<h1 id="file-sharing-system-api">File Sharing System API v1.0.0</h1>


Hệ thống chia sẻ file tạm thời với các tính năng:
- Upload file và tạo link chia sẻ
- Thiết lập quyền truy cập (public/password-protected/private)
- Thời gian hiệu lực linh hoạt (from/to)
- Bảo vệ bằng mật khẩu hoặc TOTP
- Chia sẻ với danh sách người dùng cụ thể
- Tự động xóa file hết hạn

Base URLs:

* <a href="http://localhost:8080">http://localhost:8080</a>

* <a href="https://api.filesharing.com">https://api.filesharing.com</a>

Email: <a href="mailto:support@filesharing.com">API Support</a> 
License: <a href="https://opensource.org/licenses/MIT">MIT</a>

# Authentication

- HTTP Authentication, scheme: bearer JWT token từ /auth/login

* API Key (CronSecret)
    - Parameter Name: **X-Cron-Secret**, in: header. Secret key cho cron job (lưu trong env, không commit vào repo).
- Nên thiết lập cơ chế rotation cố định (ví dụ đổi secret 30/60 ngày một lần hoặc khi có sự cố).
- Việc rotation gồm: tạo secret mới, cập nhật vào store an toàn (secret manager/CI), redeploy cron job, vô hiệu hóa secret cũ.
- Ghi log thời điểm tạo/thu hồi secret để phục vụ audit.

<h1 id="file-sharing-system-api-authentication">Authentication</h1>

User authentication and authorization

## post__auth_register

`POST /auth/register`

*Đăng ký tài khoản mới*

Tạo tài khoản người dùng mới (không bắt buộc để upload file).

**Lưu ý:**
- User tự đăng ký sẽ luôn có `role = user` (mặc định)
- Admin account được tạo qua script/seed data hoặc endpoint riêng (chỉ admin gọi được)

**Lưu ý về TOTP/2FA:**
- Không thể bật TOTP ngay khi đăng ký
- Sau khi đăng ký, user cần đăng nhập để lấy Bearer token
- Sau đó gọi `/auth/totp/setup` và `/auth/totp/verify` để bật 2FA

> Body parameter

```json
{
    "username": "nam123",
    "email": "<nam@example.com>",
    "password": "123456"
}
```
| Field        | Type    | Required | Description                             |
| :------------| :-------| :--------| :---------------------------------------|
| `username`   | string  | yes      | Tên người dùng (unique)                 |
| `email`      | string  | yes      | Email (unique)                          |
| `password`   | string  | yes      | Mật khẩu (tối thiểu 6 ký tự)            |

**Response (200 OK)**
```json
{
    "message": "User registered successfully",
    "userId": "550e8400-e29b-41d4-a716-446655440000"
}
```
## `POST /api/auth/login`

Đăng nhập để lấy JWT token.

**Request Body**
```json
{
    "email": "<nam@example.com>",
    "password": "123456"
}
```
**Response (200 OK - không bật TOTP)**
```json
{
    "accessToken": "eyJhbGciOi...",
    "user": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "username": "nam123",
        "email": "<nam@example.com>"
    }
}
```
**Response (200 OK - có TOTP)**
```json
{
    "cid": "431ba686-c69d....",
    "message": "TOTP verification required",
    "requireTOTP": true
}
```
## `POST /api/auth/login/totp`

> Body parameter

```json
{
    "cid": "a3c34041-7e57e5...",
    "code": "123456"
}
```
**Response (200 OK)**
```json
{
    "accessToken": "eyJhbGciOi...",
    "user": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "username": "nam123",
        "email": "<nam@example.com>"
    }
}
```
## `POST /api/auth/totp/setup`

Bật hoặc reset TOTP.

**Headers**
- Authorization: Bearer `<token>`

**Response (200 OK)**
```json
{
    "message": "TOTP secret generated",
    "totpSetup": {
        "secret": "NB2W45DFOIZA====",
        "qrCode": "data:image/png;base64,..."
    }
}
```

<h3 id="post__auth_login-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[LoginRequest](#schemaloginrequest)|true|none|

> Example responses

> Đăng nhập thành công hoặc yêu cầu TOTP

```json
{
  "accessToken": "eyJhbGciOi...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "nam123",
    "email": "nam@example.com"
  }
}
```

```json
{
  "requireTOTP": true,
  "message": "TOTP verification required",
  "cid": "8d4f3bb1-2f52-4a76-b951-7c21ef991abc"
}
```

> Sai email hoặc password

```json
{
  "error": "Unauthorized",
  "message": "Invalid email or password"
}
```

<h3 id="post__auth_login-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Đăng nhập thành công hoặc yêu cầu TOTP|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|Sai email hoặc password|[Error](#schemaerror)|

<h3 id="post__auth_login-responseschema">Response Schema</h3>

#### Enumerated Values

|Property|Value|
|---|---|
|role|user|
|role|admin|

<aside class="success">
This operation does not require authentication
</aside>

## post__auth_login_totp

`POST /auth/login/totp`

*Xác thực TOTP để hoàn tất đăng nhập*

Gọi endpoint này sau khi `/auth/login` trả về `requireTOTP: true`.
Xác thực mã TOTP 6 chữ số để lấy access token.

**Không cần Bearer token** vì đây là bước hoàn tất đăng nhập - user chưa có token.
Backend xác minh mã TOTP đúng thì mới cấp access token lần đầu.

> Body parameter

```json
{
  "cid": "8d4f3bb1-2f52-4a76-b951-7c21ef991abc",
  "code": "123456"
}
```

<h3 id="post__auth_login_totp-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[TOTPVerifyRequest](#schematotpverifyrequest)|true|none|

> Example responses

> 200 Response

```json
{
  "accessToken": "eyJhbGciOi...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "nam123",
    "email": "nam@example.com",
    "role": "user",
    "totpEnabled": true
  }
}
```

> Mã TOTP không hợp lệ hoặc phiên đăng nhập đã hết hạn

```json
{
  "error": "Unauthorized",
  "message": "Invalid or expired TOTP code"
}
```

```json
{
  "error": "Unauthorized",
  "message": "Login session expired. Please restart the login flow."
}
```

<h3 id="post__auth_login_totp-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Xác thực thành công|[LoginResponse](#schemaloginresponse)|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|Mã TOTP không hợp lệ hoặc phiên đăng nhập đã hết hạn|[Error](#schemaerror)|

<aside class="success">
This operation does not require authentication
</aside>

## post__auth_totp_setup

`POST /auth/totp/setup`

*Thiết lập TOTP*

Bật hoặc reset TOTP (2FA) cho user.

**Yêu cầu:** User phải đăng nhập trước (cần Bearer token).

**Luồng sử dụng:**
1. User đăng nhập bình thường để lấy Bearer token
2. Gọi endpoint này để nhận TOTP secret và QR code
3. Quét QR code bằng app authenticator (Google Authenticator, Authy,...)
4. Gọi `/auth/totp/verify` với mã 6 số để kích hoạt 2FA

> Example responses

> TOTP secret được tạo

```json
{
  "message": "TOTP secret generated",
  "totpSetup": {
    "secret": "NB2W45DFOIZA====",
    "qrCode": "data:image/png;base64,iVBORw0KGgo..."
  }
}
```

> Thiếu hoặc sai Bearer token

```json
{
  "error": "Unauthorized",
  "message": "Bearer token is required"
}
```

```json
{
  "error": "Unauthorized",
  "message": "Invalid or expired access token"
}
```

<h3 id="post__auth_totp_setup-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|TOTP secret được tạo|[TOTPSetupResponse](#schematotpsetupresponse)|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|Thiếu hoặc sai Bearer token|[Error](#schemaerror)|

<aside class="warning">
To perform this operation, you must be authenticated by means of one of the following methods:
BearerAuth
</aside>

## post__auth_totp_verify

`POST /auth/totp/verify`

*Xác minh mã TOTP để kích hoạt 2FA*

Xác minh mã TOTP 6 chữ số để kích hoạt 2FA cho tài khoản.

**Yêu cầu:** User phải đăng nhập và đã gọi `/auth/totp/setup` trước đó (cần Bearer token).

**Sau khi verify thành công:**
- Tài khoản được đánh dấu `totpEnabled = true`
- Các lần đăng nhập sau sẽ yêu cầu mã TOTP

> Body parameter

```json
{
  "code": "123456"
}
```

<h3 id="post__auth_totp_verify-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|object|true|none|
|» code|body|string|true|Mã TOTP 6 chữ số từ app authenticator|

> Example responses

> 200 Response

```json
{
  "message": "TOTP verified successfully",
  "totpEnabled": true
}
```

> Mã TOTP không hợp lệ hoặc đã hết hạn

```json
{
  "error": "Invalid TOTP code",
  "message": "The provided code is incorrect or expired"
}
```

> Thiếu hoặc sai Bearer token

```json
{
  "error": "Unauthorized",
  "message": "Bearer token is required"
}
```

```json
{
  "error": "Unauthorized",
  "message": "Invalid or expired access token"
}
```

<h3 id="post__auth_totp_verify-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|TOTP được xác minh thành công, 2FA đã được kích hoạt|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|Mã TOTP không hợp lệ hoặc đã hết hạn|[Error](#schemaerror)|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|Thiếu hoặc sai Bearer token|[Error](#schemaerror)|

<h3 id="post__auth_totp_verify-responseschema">Response Schema</h3>

Status Code **200**

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|» message|string|false|none|none|
|» totpEnabled|boolean|false|none|none|

<aside class="warning">
To perform this operation, you must be authenticated by means of one of the following methods:
BearerAuth
</aside>

## post__auth_logout

`POST /auth/logout`

*Đăng xuất*

Đăng xuất (client tự xóa token)

> Example responses

> 200 Response

```json
{
  "message": "User logged out"
}
```

<h3 id="post__auth_logout-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Đăng xuất thành công|Inline|

<h3 id="post__auth_logout-responseschema">Response Schema</h3>

Status Code **200**

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|» message|string|false|none|none|

<aside class="warning">
To perform this operation, you must be authenticated by means of one of the following methods:
BearerAuth
</aside>

## get__user

`GET /user`

*Lấy thông tin profile của user hiện tại*

Trả về thông tin profile của user hiện tại (id, username, email, role, totpEnabled, TOTP status, ...).

> Example responses

> Thông tin user

```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "nam123",
    "email": "nam@example.com",
    "role": "user",
    "totpEnabled": true
  }
}
```

> 401 Response

```json
{
  "error": "Unauthorized",
  "message": "Invalid or missing authentication token"
}
```

<h3 id="get__user-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Thông tin user|[UserProfileResponse](#schemauserprofileresponse)|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|Unauthorized - Missing or invalid token|[Error](#schemaerror)|

<aside class="warning">
To perform this operation, you must be authenticated by means of one of the following methods:
BearerAuth
</aside>

<h1 id="file-sharing-system-api-files">Files</h1>

File management operations

## post__files_upload

`POST /files/upload`

*Upload file*

Upload file mới và tạo share link.
Authorization header là optional - nếu không có thì anonymous upload (chỉ upload được file public).

**Lưu ý:**
- Các cấu hình private (`isPublic = false`, dùng whitelist `sharedWith`, password nâng cao) yêu cầu Bearer token để hệ thống gắn owner.
- Anonymous upload luôn `isPublic = true` và không thể chỉnh sửa/xóa file sau khi upload.
- Thời gian hiệu lực được validate theo system policy: `availableFrom` phải nhỏ hơn hoặc bằng `availableTo`, `availableTo` không được nằm trong quá khứ và tổng thời gian hiệu lực không vượt quá `maxValidityDays`. Nếu bỏ trống, backend tự áp dụng logic mặc định (FROM+TO/Chỉ FROM/Chỉ TO/Không có) như mô tả ở phần Validity Period.

> Body parameter

```yaml
file: string
isPublic: true
password: stringst
availableFrom: 2025-11-10T00:00:00Z
availableTo: 2025-11-17T00:00:00Z
sharedWith:
  - user1@example.com
  - user2@example.com

```

<h3 id="post__files_upload-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[FileUploadRequest](#schemafileuploadrequest)|true|none|

> Example responses

> Upload thành công

```json
{
  "success": true,
  "message": "File uploaded successfully",
  "file": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "fileName": "report.pdf",
    "shareToken": "a1b2c3d4e5f6g7h8"
  },
  "totpSetup": {
    "secret": "NB2W45DFOIZA====",
    "qrCode": "data:image/png;base64,iVBORw0KGgo..."
  }
}
```

> Thiếu file hoặc vi phạm validation (password quá ngắn, cấu hình thời gian không hợp lệ, ... )

```json
{
  "error": "Validation error",
  "message": "File is required"
}
```

```json
{
  "error": "Validation error",
  "message": "Password must have at least 8 characters"
}
```

```json
{
  "error": "Validation error",
  "message": "availableFrom must be before availableTo and within allowed policy window"
}
```

> Thiếu hoặc sai Bearer token (khi upload với quyền riêng tư)

```json
{
  "error": "Unauthorized",
  "message": "Bearer token is required for authenticated uploads"
}
```

```json
{
  "error": "Unauthorized",
  "message": "Private uploads (isPublic=false/sharedWith) require authentication"
}
```

> File quá lớn

```json
{
  "error": "Payload too large",
  "message": "File size exceeds the system limit"
}
```

<h3 id="post__files_upload-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|201|[Created](https://tools.ietf.org/html/rfc7231#section-6.3.2)|Upload thành công|[FileUploadResponse](#schemafileuploadresponse)|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|Thiếu file hoặc vi phạm validation (password quá ngắn, cấu hình thời gian không hợp lệ, ... )|[Error](#schemaerror)|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|Thiếu hoặc sai Bearer token (khi upload với quyền riêng tư)|[Error](#schemaerror)|
|413|[Payload Too Large](https://tools.ietf.org/html/rfc7231#section-6.5.11)|File quá lớn|[Error](#schemaerror)|

<aside class="warning">
To perform this operation, you must be authenticated by means of one of the following methods:
BearerAuth, None
</aside>

## get__files_my

`GET /files/my`

*Lấy danh sách file của user hiện tại*

Trả về danh sách file do user hiện tại upload (đã đăng nhập).

**Bao gồm:**
- Danh sách file kèm metadata (status, thời gian hiệu lực, bảo mật, ...)
- Pagination (`page`, `limit`) và tổng số file
- Summary (đếm active/pending/expired/deleted)

<h3 id="get__files_my-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|status|query|string|false|Lọc theo trạng thái file|
|page|query|integer|false|Số trang|
|limit|query|integer|false|Số file mỗi trang|
|sortBy|query|string|false|Sắp xếp theo|
|order|query|string|false|Thứ tự sắp xếp|

#### Enumerated Values

|Parameter|Value|
|---|---|
|status|active|
|status|expired|
|status|pending|
|status|deleted|
|status|all|
|sortBy|createdAt|
|sortBy|fileName|
|order|asc|
|order|desc|

> Example responses

> Danh sách file của user hiện tại

```json
{
  "files": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "fileName": "document.pdf",
      "status": "active",
      "createdAt": "2025-11-19T10:00:00Z"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "fileName": "old-file.pdf",
      "status": "expired",
      "createdAt": "2025-11-10T10:00:00Z"
    }
  ],
  "pagination": {
    "currentPage": 1,
    "totalPages": 3,
    "totalFiles": 42,
    "limit": 20
  },
  "summary": {
    "activeFiles": 28,
    "pendingFiles": 5,
    "expiredFiles": 9,
    "deletedFiles": 0
  }
}
```

> 401 Response

```json
{
  "error": "Unauthorized",
  "message": "Invalid or missing authentication token"
}
```

<h3 id="get__files_my-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Danh sách file của user hiện tại|[UserFilesResponse](#schemauserfilesresponse)|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|Unauthorized - Missing or invalid token|[Error](#schemaerror)|

<aside class="warning">
To perform this operation, you must be authenticated by means of one of the following methods:
BearerAuth
</aside>

## get__files_info_{id}

`GET /files/info/{id}`

*Lấy thông tin file theo ID*

Lấy thông tin chi tiết của file theo UUID (chỉ owner hoặc admin).

**Khác với `/files/{shareToken}`:**
- Endpoint này dùng UUID (file ID), yêu cầu authentication
- Chỉ owner hoặc admin mới truy cập được
- Trả về đầy đủ metadata bao gồm cả thông tin nhạy cảm (không public)

**Use case:** Owner xem thông tin file của mình trong dashboard

<h3 id="get__files_info_{id}-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|id|path|string(uuid)|true|File UUID|

> Example responses

> Thông tin file chi tiết

```json
{
  "file": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "fileName": "contract.pdf",
    "fileSize": 2048576,
    "mimeType": "application/pdf",
    "shareToken": "a1b2c3d4e5f6g7h8",
    "shareLink": "https://example.com/f/a1b2c3d4e5f6g7h8",
    "isPublic": false,
    "hasPassword": true,
    "totpEnabled": true,
    "availableFrom": "2025-11-10T00:00:00Z",
    "availableTo": "2025-11-17T00:00:00Z",
    "status": "active",
    "hoursRemaining": 120.5,
    "sharedWith": [
      "user1@example.com",
      "user2@example.com"
    ],
    "owner": {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "username": "nam123",
      "email": "nam@example.com",
      "role": "user"
    },
    "createdAt": "2025-11-10T10:00:00Z"
  }
}
```

> 401 Response

```json
{
  "error": "Unauthorized",
  "message": "Invalid or missing authentication token"
}
```

> Không phải owner hoặc admin

```json
{
  "error": "Forbidden",
  "message": "You don't have permission to access this file"
}
```

> Không tìm thấy file với ID

```json
{
  "error": "Not found",
  "message": "File not found"
}
```

<h3 id="get__files_info_{id}-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Thông tin file chi tiết|[FileInfoResponse](#schemafileinforesponse)|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|Unauthorized - Missing or invalid token|[Error](#schemaerror)|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|Không phải owner hoặc admin|[Error](#schemaerror)|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|Không tìm thấy file với ID|[Error](#schemaerror)|

<aside class="warning">
To perform this operation, you must be authenticated by means of one of the following methods:
BearerAuth
</aside>

## delete__files_info_{id}

`DELETE /files/info/{id}`

*Xóa file*

Xóa file theo UUID (chỉ owner hoặc admin).

**Quyền truy cập:**
- Chỉ owner của file hoặc admin mới có thể xóa
- Anonymous uploads (file không có owner) **KHÔNG THỂ** xóa file sau khi upload

<h3 id="delete__files_info_{id}-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|id|path|string(uuid)|true|File UUID|

> Example responses

> Xóa thành công

```json
{
  "message": "File deleted successfully",
  "fileId": "550e8400-e29b-41d4-a716-446655440000"
}
```

> 401 Response

```json
{
  "error": "Unauthorized",
  "message": "Invalid or missing authentication token"
}
```

> Không phải owner hoặc anonymous upload

```json
{
  "error": "Forbidden",
  "message": "You don't have permission to delete this file"
}
```

```json
{
  "error": "Forbidden",
  "message": "Anonymous uploads cannot be deleted"
}
```

> Không tìm thấy file với id

```json
{
  "error": "Not found",
  "message": "File not found"
}
```

<h3 id="delete__files_info_{id}-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Xóa thành công|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|Unauthorized - Missing or invalid token|[Error](#schemaerror)|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|Không phải owner hoặc anonymous upload|[Error](#schemaerror)|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|Không tìm thấy file với id|[Error](#schemaerror)|

<h3 id="delete__files_info_{id}-responseschema">Response Schema</h3>

Status Code **200**

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|» message|string|false|none|none|
|» fileId|string(uuid)|false|none|none|

<aside class="warning">
To perform this operation, you must be authenticated by means of one of the following methods:
BearerAuth
</aside>

## get__files_stats_{id}

`GET /files/stats/{id}`

*Lấy thống kê download của file*

Lấy statistics của file từ bảng `file_statistics`.

**Yêu cầu:** Chỉ owner hoặc admin mới xem được

**Dữ liệu trả về:**
- `downloadCount`: Tổng số lượt download
- `uniqueDownloaders`: Số người download khác nhau (authenticated users)
- `lastDownloadedAt`: Lần download gần nhất

**Lưu ý:** Anonymous uploads không có statistics

<h3 id="get__files_stats_{id}-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|id|path|string(uuid)|true|File UUID|

> Example responses

> Statistics của file

```json
{
  "fileId": "550e8400-e29b-41d4-a716-446655440000",
  "fileName": "presentation.pdf",
  "statistics": {
    "downloadCount": 45,
    "uniqueDownloaders": 12,
    "lastDownloadedAt": "2025-11-19T14:30:00Z",
    "createdAt": "2025-11-10T10:00:00Z"
  }
}
```

```json
{
  "fileId": "550e8400-e29b-41d4-a716-446655440000",
  "fileName": "document.docx",
  "statistics": {
    "downloadCount": 0,
    "uniqueDownloaders": 0,
    "lastDownloadedAt": null,
    "createdAt": "2025-11-18T08:00:00Z"
  }
}
```

> 401 Response

```json
{
  "error": "Unauthorized",
  "message": "Invalid or missing authentication token"
}
```

> Không phải owner hoặc admin

```json
{
  "error": "Forbidden",
  "message": "You don't have permission to view statistics for this file"
}
```

> File không tồn tại hoặc là anonymous upload

```json
{
  "error": "Not found",
  "message": "File not found or statistics not available (anonymous upload)"
}
```

<h3 id="get__files_stats_{id}-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Statistics của file|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|Unauthorized - Missing or invalid token|[Error](#schemaerror)|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|Không phải owner hoặc admin|[Error](#schemaerror)|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|File không tồn tại hoặc là anonymous upload|[Error](#schemaerror)|

<h3 id="get__files_stats_{id}-responseschema">Response Schema</h3>

Status Code **200**

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|» fileId|string(uuid)|false|none|none|
|» fileName|string|false|none|none|
|» statistics|object|false|none|none|
|»» downloadCount|integer|false|none|Tổng số lượt download|
|»» uniqueDownloaders|integer|false|none|Số người download khác nhau (không tính anonymous)|
|»» lastDownloadedAt|string(date-time)|false|none|Thời điểm download gần nhất|
|»» createdAt|string(date-time)|false|none|none|

<aside class="warning">
To perform this operation, you must be authenticated by means of one of the following methods:
BearerAuth
</aside>

## get__files_download-history_{id}

`GET /files/download-history/{id}`

*Lấy lịch sử download chi tiết*

Lấy download history từ bảng `download_history` (tương tự browser download history).

**Yêu cầu:** Chỉ owner hoặc admin mới xem được

**Thông tin mỗi download:**
- Người download (username/email hoặc "Anonymous")
- Thời điểm download
- Trạng thái (completed/interrupted)

**Quyền riêng tư:**
- Với anonymous download, hệ thống chỉ ghi nhận một bản ghi mang nhãn "Anonymous" cùng timestamp và trạng thái; **không log IP/User-Agent hoặc fingerprint**.
- Thông tin cá nhân (username/email) chỉ xuất hiện khi người tải đăng nhập và đồng ý với điều khoản sử dụng.

**Pagination:** Hỗ trợ phân trang với `page` và `limit`

<h3 id="get__files_download-history_{id}-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|id|path|string(uuid)|true|File UUID|
|page|query|integer|false|Số trang|
|limit|query|integer|false|Số record mỗi trang|

> Example responses

> Lịch sử download

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
    },
    {
      "id": "650e8400-e29b-41d4-a716-446655440003",
      "downloader": {
        "username": "nguyenvana",
        "email": "nguyenvana@example.com"
      },
      "downloadedAt": "2025-11-18T16:45:00Z",
      "downloadCompleted": false
    }
  ],
  "pagination": {
    "currentPage": 1,
    "totalPages": 1,
    "totalRecords": 3,
    "limit": 50
  }
}
```

> 401 Response

```json
{
  "error": "Unauthorized",
  "message": "Invalid or missing authentication token"
}
```

> Không phải owner hoặc admin

```json
{
  "error": "Forbidden",
  "message": "You don't have permission to view download history for this file"
}
```

> File không tồn tại

```json
{
  "error": "Not found",
  "message": "File not found"
}
```

<h3 id="get__files_download-history_{id}-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Lịch sử download|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|Unauthorized - Missing or invalid token|[Error](#schemaerror)|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|Không phải owner hoặc admin|[Error](#schemaerror)|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|File không tồn tại|[Error](#schemaerror)|

<h3 id="get__files_download-history_{id}-responseschema">Response Schema</h3>

Status Code **200**

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|» fileId|string(uuid)|false|none|none|
|» fileName|string|false|none|none|
|» history|[object]|false|none|none|
|»» id|string(uuid)|false|none|none|
|»» downloader|object|false|none|Null nếu là anonymous download (chỉ ghi nhận "Anonymous" + timestamp, không lưu IP/User-Agent)|
|»»» username|string¦null|false|none|none|
|»»» email|string¦null|false|none|none|
|»» downloadedAt|string(date-time)|false|none|none|
|»» downloadCompleted|boolean|false|none|false nếu download bị gián đoạn|
|» pagination|object|false|none|none|
|»» currentPage|integer|false|none|none|
|»» totalPages|integer|false|none|none|
|»» totalRecords|integer|false|none|none|
|»» limit|integer|false|none|none|

<aside class="warning">
To perform this operation, you must be authenticated by means of one of the following methods:
BearerAuth
</aside>

## get__files_{shareToken}

`GET /files/{shareToken}`

*Lấy thông tin file*

Lấy metadata cơ bản của file qua share token (public endpoint, không cần authentication).

**Response chỉ trả về thông tin cơ bản:**
- `id`, `fileName`, `shareToken`, `status`, `isPublic`, `hasPassword`

<h3 id="get__files_{sharetoken}-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|shareToken|path|string|true|Share token của file|

> Example responses

> Thông tin file

```json
{
  "file": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "fileName": "contract.pdf",
    "shareToken": "a1b2c3d4e5f6g7h8",
    "status": "active",
    "isPublic": false,
    "hasPassword": true
  }
}
```

```json
{
  "file": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "fileName": "document.pdf",
    "shareToken": "b2c3d4e5f6g7h8i9",
    "status": "active",
    "isPublic": true,
    "hasPassword": false
  }
}
```

> Không tìm thấy file với shareToken

```json
{
  "error": "Not found",
  "message": "File not found"
}
```

> File đã hết hạn

```json
{
  "error": "File expired",
  "message": "File has expired"
}
```

<h3 id="get__files_{sharetoken}-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Thông tin file|[FileInfoResponse](#schemafileinforesponse)|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|Không tìm thấy file với shareToken|[Error](#schemaerror)|
|410|[Gone](https://tools.ietf.org/html/rfc7231#section-6.5.9)|File đã hết hạn|[Error](#schemaerror)|

<aside class="success">
This operation does not require authentication
</aside>

## get__files_{shareToken}_download

`GET /files/{shareToken}/download`

*Tải file về*

Download file. File có thể có nhiều lớp bảo mật cùng lúc (password + whitelist).

**Thứ tự kiểm tra bảo mật (theo best practice):**
1. **File status** - Kiểm tra file còn hiệu lực (expired/pending) → 410/423
2. **Whitelist** - Nếu file có `sharedWith` list → yêu cầu Bearer token, verify user email ∈ whitelist → 403 nếu không có quyền
3. **Password** - Nếu file có password → yêu cầu header `X-File-Password` → 403 nếu sai/thiếu

**Lưu ý:** Tất cả các lớp bảo mật phải pass thì mới được download. Bất kỳ lớp nào fail sẽ trả error tương ứng.

**Owner preview trong giai đoạn pending:**
- Nếu requester là owner (JWT hợp lệ, `sub` trùng `file.owner.id`), hệ thống cho phép bypass trạng thái `pending` để chủ file có thể kiểm thử trước khi tới giờ.
- Tất cả user khác vẫn nhận 423 "File not yet available" cho tới khi `availableFrom` đến.

**Notification khi file chuyển trạng thái:**
- Khuyến nghị cấu hình cronjob/background task gửi email/SMS/webhook khi file chuyển từ `pending` sang `active` để owner và whitelist nhận thông báo.
- Endpoint này không tự gửi notification; cần bật tính năng ở backend hoặc job scheduler (xem mục Appendices/Future Enhancements).

<h3 id="get__files_{sharetoken}_download-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|shareToken|path|string|true|Share token của file|
|X-File-Password|header|string|false|Mật khẩu bảo vệ (nếu file có password)|

> Example responses

> 200 Response

> Thiếu Bearer token khi file có whitelist (sharedWith)

```json
{
  "error": "Unauthorized",
  "message": "This file requires authentication. Please provide a Bearer token"
}
```

> Sai password, thiếu password, hoặc không có quyền truy cập

```json
{
  "error": "Incorrect password",
  "message": "The file password is incorrect"
}
```

```json
{
  "error": "Password required",
  "message": "This file is password-protected. Please provide the password parameter"
}
```

```json
{
  "error": "Access denied",
  "message": "You are not allowed to download this file. Your email is not in the shared list"
}
```

> Không tìm thấy file

```json
{
  "error": "Not found",
  "message": "File not found"
}
```

> File đã hết hạn

```json
{
  "error": "File expired",
  "expiredAt": "2025-11-19T16:20:18.619Z"
}
```

> File chưa đến thời gian hiệu lực

```json
{
  "error": "File not yet available",
  "availableFrom": "2025-11-20T10:00:00Z",
  "hoursUntilAvailable": 6
}
```

<h3 id="get__files_{sharetoken}_download-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|File binary|string|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|Thiếu Bearer token khi file có whitelist (sharedWith)|[Error](#schemaerror)|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|Sai password, thiếu password, hoặc không có quyền truy cập|[Error](#schemaerror)|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|Không tìm thấy file|[Error](#schemaerror)|
|410|[Gone](https://tools.ietf.org/html/rfc7231#section-6.5.9)|File đã hết hạn|Inline|
|423|[Locked](https://tools.ietf.org/html/rfc2518#section-10.4)|File chưa đến thời gian hiệu lực|Inline|

<h3 id="get__files_{sharetoken}_download-responseschema">Response Schema</h3>

Status Code **410**

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|» error|string|false|none|none|
|» expiredAt|string(date-time)|false|none|none|

Status Code **423**

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|» error|string|false|none|none|
|» availableFrom|string(date-time)|false|none|none|
|» hoursUntilAvailable|number|false|none|none|

### Response Headers

|Status|Header|Type|Format|Description|
|---|---|---|---|---|
|200|Content-Disposition|string||none|
|200|Content-Length|integer||none|

<aside class="warning">
To perform this operation, you must be authenticated by means of one of the following methods:
BearerAuth, None
</aside>

<h1 id="file-sharing-system-api-admin">Admin</h1>

Admin-only operations

## post__admin_cleanup

`POST /admin/cleanup`

*Xóa file hết hạn*

Xóa file hết hạn (Cron job hoặc Admin endpoint).
Yêu cầu X-Cron-Secret header hoặc admin token.

**Security best practices:**
- `X-Cron-Secret` phải được lưu trong môi trường bảo mật, hỗ trợ rotation định kỳ (ví dụ đổi secret hàng tuần/tháng).
- Nên triển khai rate limiting/IP allowlist để tránh bị spam cleanup (DOS).
- Mọi request tới endpoint này cần được log đầy đủ (timestamp, source, kết quả) để audit.
- Khuyến nghị triển khai song song xác thực bằng JWT admin đối với các tác vụ thủ công; secret chủ yếu dành cho job nội bộ.

> Example responses

> Cleanup hoàn tất

```json
{
  "message": "Expired files removed",
  "deletedFiles": 12,
  "timestamp": "2025-11-19T10:00:00Z"
}
```

> Thiếu hoặc sai Bearer token / X-Cron-Secret

```json
{
  "error": "Unauthorized",
  "message": "X-Cron-Secret header is required"
}
```

```json
{
  "error": "Unauthorized",
  "message": "Invalid or expired admin token"
}
```

> Không phải admin hoặc secret sai

```json
{
  "error": "Forbidden",
  "message": "You don't have permission to perform cleanup"
}
```

```json
{
  "error": "Forbidden",
  "message": "Invalid cron secret"
}
```

> Vượt quá giới hạn gọi cleanup

```json
{
  "error": "Too many requests",
  "message": "Cleanup endpoint is rate limited. Please try again later."
}
```

<h3 id="post__admin_cleanup-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Cleanup hoàn tất|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|Thiếu hoặc sai Bearer token / X-Cron-Secret|[Error](#schemaerror)|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|Không phải admin hoặc secret sai|[Error](#schemaerror)|
|429|[Too Many Requests](https://tools.ietf.org/html/rfc6585#section-4)|Vượt quá giới hạn gọi cleanup|[Error](#schemaerror)|

<h3 id="post__admin_cleanup-responseschema">Response Schema</h3>

Status Code **200**

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|» message|string|false|none|none|
|» deletedFiles|integer|false|none|none|
|» timestamp|string(date-time)|false|none|none|

<aside class="warning">
To perform this operation, you must be authenticated by means of one of the following methods:
BearerAuth, CronSecret
</aside>

## get__admin_policy

`GET /admin/policy`

*Lấy cấu hình hệ thống*

Lấy system policy (chỉ admin)

> Example responses

> System policy

```json
{
  "id": 1,
  "maxFileSizeMB": 50,
  "minValidityHours": 1,
  "maxValidityDays": 30,
  "defaultValidityDays": 7,
  "requirePasswordMinLength": 8
}
```

> Thiếu hoặc sai Bearer token

```json
{
  "error": "Unauthorized",
  "message": "Bearer token is required"
}
```

> Không phải admin

```json
{
  "error": "Forbidden",
  "message": "You don't have permission to access this resource"
}
```

<h3 id="get__admin_policy-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|System policy|[SystemPolicy](#schemasystempolicy)|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|Thiếu hoặc sai Bearer token|[Error](#schemaerror)|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|Không phải admin|[Error](#schemaerror)|

<aside class="warning">
To perform this operation, you must be authenticated by means of one of the following methods:
BearerAuth
</aside>

## patch__admin_policy

`PATCH /admin/policy`

*Cập nhật cấu hình hệ thống*

Cập nhật system policy (chỉ admin)

> Body parameter

```json
{
  "maxFileSizeMB": 100,
  "minValidityHours": 1,
  "maxValidityDays": 14,
  "defaultValidityDays": 5,
  "requirePasswordMinLength": 8
}
```

<h3 id="patch__admin_policy-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[SystemPolicyUpdate](#schemasystempolicyupdate)|true|none|

> Example responses

> Cập nhật thành công

```json
{
  "message": "Policy updated",
  "policy": {
    "id": 1,
    "maxFileSizeMB": 100,
    "minValidityHours": 1,
    "maxValidityDays": 14,
    "defaultValidityDays": 5,
    "requirePasswordMinLength": 8
  }
}
```

> Payload không hợp lệ (giá trị < minimum)

```json
{
  "error": "Validation error",
  "message": "maxValidityDays must be greater than or equal to minValidityHours"
}
```

> Thiếu hoặc sai Bearer token

```json
{
  "error": "Unauthorized",
  "message": "Bearer token is required"
}
```

> Không phải admin

```json
{
  "error": "Forbidden",
  "message": "You don't have permission to access this resource"
}
```

<h3 id="patch__admin_policy-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Cập nhật thành công|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|Payload không hợp lệ (giá trị < minimum)|[Error](#schemaerror)|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|Thiếu hoặc sai Bearer token|[Error](#schemaerror)|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|Không phải admin|[Error](#schemaerror)|

<h3 id="patch__admin_policy-responseschema">Response Schema</h3>

Status Code **200**

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|» message|string|false|none|none|
|» policy|[SystemPolicy](#schemasystempolicy)|false|none|none|
|»» id|integer|false|none|none|
|»» maxFileSizeMB|integer|false|none|none|
|»» minValidityHours|integer|false|none|none|
|»» maxValidityDays|integer|false|none|none|
|»» defaultValidityDays|integer|false|none|none|
|»» requirePasswordMinLength|integer|false|none|none|

<aside class="warning">
To perform this operation, you must be authenticated by means of one of the following methods:
BearerAuth
</aside>

# Schemas

<h2 id="tocS_RegisterRequest">RegisterRequest</h2>
<!-- backwards compatibility -->
<a id="schemaregisterrequest"></a>
<a id="schema_RegisterRequest"></a>
<a id="tocSregisterrequest"></a>
<a id="tocsregisterrequest"></a>

```json
{
  "username": "nam123",
  "email": "nam@example.com",
  "password": "Passw0rd"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|username|string|true|none|Tên người dùng (unique)|
|email|string(email)|true|none|Email (unique)|
|password|string(password)|true|none|Mật khẩu (tối thiểu 8 ký tự)|

<h2 id="tocS_RegisterResponse">RegisterResponse</h2>
<!-- backwards compatibility -->
<a id="schemaregisterresponse"></a>
<a id="schema_RegisterResponse"></a>
<a id="tocSregisterresponse"></a>
<a id="tocsregisterresponse"></a>

```json
{
  "message": "User registered successfully",
  "userId": "550e8400-e29b-41d4-a716-446655440000"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|message|string|false|none|none|
|userId|string(uuid)|false|none|none|

<h2 id="tocS_LoginRequest">LoginRequest</h2>
<!-- backwards compatibility -->
<a id="schemaloginrequest"></a>
<a id="schema_LoginRequest"></a>
<a id="tocSloginrequest"></a>
<a id="tocsloginrequest"></a>

```json
{
  "email": "nam@example.com",
  "password": "Passw0rd"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|email|string(email)|true|none|none|
|password|string(password)|true|none|none|

<h2 id="tocS_LoginResponse">LoginResponse</h2>
<!-- backwards compatibility -->
<a id="schemaloginresponse"></a>
<a id="schema_LoginResponse"></a>
<a id="tocSloginresponse"></a>
<a id="tocsloginresponse"></a>

```json
{
  "accessToken": "eyJhbGciOi...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "nam123",
    "email": "nam@example.com",
    "role": "user",
    "totpEnabled": true
  }
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|accessToken|string|false|none|none|
|user|[User](#schemauser)|false|none|none|

<h2 id="tocS_TOTPRequiredResponse">TOTPRequiredResponse</h2>
<!-- backwards compatibility -->
<a id="schematotprequiredresponse"></a>
<a id="schema_TOTPRequiredResponse"></a>
<a id="tocStotprequiredresponse"></a>
<a id="tocstotprequiredresponse"></a>

```json
{
  "requireTOTP": true,
  "message": "TOTP verification required",
  "cid": "8d4f3bb1-2f52-4a76-b951-7c21ef991abc"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|requireTOTP|boolean|false|none|none|
|message|string|false|none|none|
|cid|string|false|none|Challenge ID cho phiên đăng nhập TOTP|

<h2 id="tocS_TOTPVerifyRequest">TOTPVerifyRequest</h2>
<!-- backwards compatibility -->
<a id="schematotpverifyrequest"></a>
<a id="schema_TOTPVerifyRequest"></a>
<a id="tocStotpverifyrequest"></a>
<a id="tocstotpverifyrequest"></a>

```json
{
  "cid": "8d4f3bb1-2f52-4a76-b951-7c21ef991abc",
  "code": "123456"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|cid|string|false|none|Challenge ID nhận từ bước /auth/login|
|code|string|true|none|none|

<h2 id="tocS_TOTPSetupResponse">TOTPSetupResponse</h2>
<!-- backwards compatibility -->
<a id="schematotpsetupresponse"></a>
<a id="schema_TOTPSetupResponse"></a>
<a id="tocStotpsetupresponse"></a>
<a id="tocstotpsetupresponse"></a>

```json
{
  "message": "TOTP secret generated",
  "totpSetup": {
    "secret": "NB2W45DFOIZA====",
    "qrCode": "data:image/png;base64,..."
  }
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|message|string|false|none|none|
|totpSetup|object|false|none|none|
|» secret|string|false|none|none|
|» qrCode|string|false|none|none|

<h2 id="tocS_FileUploadRequest">FileUploadRequest</h2>
<!-- backwards compatibility -->
<a id="schemafileuploadrequest"></a>
<a id="schema_FileUploadRequest"></a>
<a id="tocSfileuploadrequest"></a>
<a id="tocsfileuploadrequest"></a>

```json
{
  "file": "string",
  "isPublic": true,
  "password": "stringst",
  "availableFrom": "2025-11-10T00:00:00Z",
  "availableTo": "2025-11-17T00:00:00Z",
  "sharedWith": [
    "user1@example.com",
    "user2@example.com"
  ]
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|file|string(binary)|true|none|File cần upload|
|isPublic|boolean|false|none|File public hay private (upload ẩn danh luôn = true)|
|password|string|false|none|Mật khẩu bảo vệ (min 8 ký tự)|
|availableFrom|string(date-time)|false|none|Thời điểm bắt đầu hiệu lực (không được lớn hơn `availableTo` và không vượt quá khoảng thời gian cho phép trong `system_policy.maxValidityDays`).<br><br>**Mặc định:** Nếu cả `availableFrom` và `availableTo` đều null, hệ thống tự động set:<br>- `availableFrom` = thời điểm hiện tại<br>- `availableTo` = thời điểm hiện tại + 7 ngày (default_validity_days)|
|availableTo|string(date-time)|false|none|Thời điểm kết thúc hiệu lực (không nằm trong quá khứ, phải lớn hơn `availableFrom` và nhỏ hơn giới hạn `maxValidityDays`).<br><br>**Mặc định:** Nếu cả `availableFrom` và `availableTo` đều null, hệ thống tự động set:<br>- `availableFrom` = thời điểm hiện tại<br>- `availableTo` = thời điểm hiện tại + 7 ngày (default_validity_days)|
|sharedWith|[string]|false|none|Danh sách email được phép tải (yêu cầu authenticated upload)|

<h2 id="tocS_FileUploadResponse">FileUploadResponse</h2>
<!-- backwards compatibility -->
<a id="schemafileuploadresponse"></a>
<a id="schema_FileUploadResponse"></a>
<a id="tocSfileuploadresponse"></a>
<a id="tocsfileuploadresponse"></a>

```json
{
  "success": true,
  "file": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "fileName": "document.pdf",
    "fileSize": 2048576,
    "mimeType": "application/pdf",
    "shareToken": "a1b2c3d4e5f6g7h8",
    "shareLink": "https://example.com/f/a1b2c3d4e5f6g7h8",
    "isPublic": false,
    "hasPassword": true,
    "availableFrom": "2025-11-10T00:00:00Z",
    "availableTo": "2025-11-17T00:00:00Z",
    "validityDays": 7,
    "status": "active",
    "hoursRemaining": 120.5,
    "sharedWith": [
      "user1@example.com",
      "user2@example.com"
    ],
    "owner": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "nam123",
      "email": "nam@example.com",
      "role": "user",
      "totpEnabled": true
    },
    "createdAt": "2025-11-04T12:00:00Z"
  },
  "message": "File uploaded successfully"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|success|boolean|false|none|none|
|file|[File](#schemafile)|false|none|none|
|message|string|false|none|none|

<h2 id="tocS_File">File</h2>
<!-- backwards compatibility -->
<a id="schemafile"></a>
<a id="schema_File"></a>
<a id="tocSfile"></a>
<a id="tocsfile"></a>

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "fileName": "document.pdf",
  "fileSize": 2048576,
  "mimeType": "application/pdf",
  "shareToken": "a1b2c3d4e5f6g7h8",
  "shareLink": "https://example.com/f/a1b2c3d4e5f6g7h8",
  "isPublic": false,
  "hasPassword": true,
  "availableFrom": "2025-11-10T00:00:00Z",
  "availableTo": "2025-11-17T00:00:00Z",
  "validityDays": 7,
  "status": "active",
  "hoursRemaining": 120.5,
  "sharedWith": [
    "user1@example.com",
    "user2@example.com"
  ],
  "owner": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "nam123",
    "email": "nam@example.com",
    "role": "user",
    "totpEnabled": true
  },
  "createdAt": "2025-11-04T12:00:00Z"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|id|string(uuid)|false|none|none|
|fileName|string|false|none|none|
|fileSize|integer|false|none|Kích thước file (bytes)|
|mimeType|string|false|none|none|
|shareToken|string|false|none|none|
|shareLink|string(uri)|false|none|none|
|isPublic|boolean|false|none|none|
|hasPassword|boolean|false|none|none|
|availableFrom|string(date-time)|false|none|none|
|availableTo|string(date-time)|false|none|none|
|validityDays|integer|false|none|none|
|status|string|false|none|- pending: Chưa đến availableFrom<br>- active: Trong thời gian hiệu lực<br>- expired: Đã hết hạn|
|hoursRemaining|number|false|none|Số giờ còn lại đến hết hạn|
|sharedWith|[string]|false|none|none|
|owner|[User](#schemauser)|false|none|none|
|createdAt|string(date-time)|false|none|none|

#### Enumerated Values

|Property|Value|
|---|---|
|status|pending|
|status|active|
|status|expired|

<h2 id="tocS_FileInfoResponse">FileInfoResponse</h2>
<!-- backwards compatibility -->
<a id="schemafileinforesponse"></a>
<a id="schema_FileInfoResponse"></a>
<a id="tocSfileinforesponse"></a>
<a id="tocsfileinforesponse"></a>

```json
{
  "file": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "fileName": "document.pdf",
    "fileSize": 2048576,
    "mimeType": "application/pdf",
    "shareToken": "a1b2c3d4e5f6g7h8",
    "shareLink": "https://example.com/f/a1b2c3d4e5f6g7h8",
    "isPublic": false,
    "hasPassword": true,
    "availableFrom": "2025-11-10T00:00:00Z",
    "availableTo": "2025-11-17T00:00:00Z",
    "validityDays": 7,
    "status": "active",
    "hoursRemaining": 120.5,
    "sharedWith": [
      "user1@example.com",
      "user2@example.com"
    ],
    "owner": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "nam123",
      "email": "nam@example.com",
      "role": "user",
      "totpEnabled": true
    },
    "createdAt": "2025-11-04T12:00:00Z"
  }
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|file|[File](#schemafile)|false|none|none|

<h2 id="tocS_UserProfileResponse">UserProfileResponse</h2>
<!-- backwards compatibility -->
<a id="schemauserprofileresponse"></a>
<a id="schema_UserProfileResponse"></a>
<a id="tocSuserprofileresponse"></a>
<a id="tocsuserprofileresponse"></a>

```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "nam123",
    "email": "nam@example.com",
    "role": "user",
    "totpEnabled": true
  }
}

```

Response cho GET /user - thông tin profile

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|user|[User](#schemauser)|false|none|none|

<h2 id="tocS_UserFilesResponse">UserFilesResponse</h2>
<!-- backwards compatibility -->
<a id="schemauserfilesresponse"></a>
<a id="schema_UserFilesResponse"></a>
<a id="tocSuserfilesresponse"></a>
<a id="tocsuserfilesresponse"></a>

```json
{
  "files": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "fileName": "document.pdf",
      "fileSize": 2048576,
      "mimeType": "application/pdf",
      "shareToken": "a1b2c3d4e5f6g7h8",
      "shareLink": "https://example.com/f/a1b2c3d4e5f6g7h8",
      "isPublic": false,
      "hasPassword": true,
      "availableFrom": "2025-11-10T00:00:00Z",
      "availableTo": "2025-11-17T00:00:00Z",
      "validityDays": 7,
      "status": "active",
      "hoursRemaining": 120.5,
      "sharedWith": [
        "user1@example.com",
        "user2@example.com"
      ],
      "owner": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "username": "nam123",
        "email": "nam@example.com",
        "role": "user",
        "totpEnabled": true
      },
      "createdAt": "2025-11-04T12:00:00Z"
    }
  ],
  "pagination": {
    "currentPage": 1,
    "totalPages": 3,
    "totalFiles": 42,
    "limit": 20
  },
  "summary": {
    "activeFiles": 28,
    "pendingFiles": 5,
    "expiredFiles": 9,
    "deletedFiles": 0
  }
}

```

Response cho GET /files/my - danh sách file của user hiện tại

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|files|[[File](#schemafile)]|false|none|Danh sách file của user (có thể filter theo status)|
|pagination|object|false|none|none|
|» currentPage|integer|false|none|none|
|» totalPages|integer|false|none|none|
|» totalFiles|integer|false|none|none|
|» limit|integer|false|none|none|
|summary|object|false|none|none|
|» activeFiles|integer|false|none|Số file đang active|
|» pendingFiles|integer|false|none|Số file chưa đến thời gian hiệu lực|
|» expiredFiles|integer|false|none|Số file đã hết hạn|
|» deletedFiles|integer|false|none|Số file đã bị xóa|

<h2 id="tocS_SystemPolicy">SystemPolicy</h2>
<!-- backwards compatibility -->
<a id="schemasystempolicy"></a>
<a id="schema_SystemPolicy"></a>
<a id="tocSsystempolicy"></a>
<a id="tocssystempolicy"></a>

```json
{
  "id": 1,
  "maxFileSizeMB": 50,
  "minValidityHours": 1,
  "maxValidityDays": 30,
  "defaultValidityDays": 7,
  "requirePasswordMinLength": 8
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|id|integer|false|none|none|
|maxFileSizeMB|integer|false|none|none|
|minValidityHours|integer|false|none|none|
|maxValidityDays|integer|false|none|none|
|defaultValidityDays|integer|false|none|none|
|requirePasswordMinLength|integer|false|none|none|

<h2 id="tocS_SystemPolicyUpdate">SystemPolicyUpdate</h2>
<!-- backwards compatibility -->
<a id="schemasystempolicyupdate"></a>
<a id="schema_SystemPolicyUpdate"></a>
<a id="tocSsystempolicyupdate"></a>
<a id="tocssystempolicyupdate"></a>

```json
{
  "maxFileSizeMB": 100,
  "minValidityHours": 1,
  "maxValidityDays": 14,
  "defaultValidityDays": 5,
  "requirePasswordMinLength": 8
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|maxFileSizeMB|integer|false|none|none|
|minValidityHours|integer|false|none|none|
|maxValidityDays|integer|false|none|none|
|defaultValidityDays|integer|false|none|none|
|requirePasswordMinLength|integer|false|none|none|

<h2 id="tocS_User">User</h2>
<!-- backwards compatibility -->
<a id="schemauser"></a>
<a id="schema_User"></a>
<a id="tocSuser"></a>
<a id="tocsuser"></a>

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "nam123",
  "email": "nam@example.com",
  "role": "user",
  "totpEnabled": true
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|id|string(uuid)|false|none|none|
|username|string|false|none|none|
|email|string(email)|false|none|none|
|role|string|false|none|none|
|totpEnabled|boolean|false|none|none|

#### Enumerated Values

|Property|Value|
|---|---|
|role|user|
|role|admin|

<h2 id="tocS_Error">Error</h2>
<!-- backwards compatibility -->
<a id="schemaerror"></a>
<a id="schema_Error"></a>
<a id="tocSerror"></a>
<a id="tocserror"></a>

```json
{
  "error": "Error message",
  "message": "Detailed error description",
  "code": "VALIDATION_ERROR"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|error|string|false|none|none|
|message|string|false|none|none|
|code|string|false|none|none|
