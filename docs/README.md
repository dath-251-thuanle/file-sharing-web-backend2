# API Hệ thống Chia sẻ File (File Sharing System API)

Một hệ thống chia sẻ file tạm thời bảo mật, cho phép người dùng upload file với thời gian hiệu lực tùy chỉnh, kiểm soát quyền truy cập và các tính năng bảo mật nâng cao như TOTP (xác thực 2 bước).

**Phiên bản:** 1.0.0  
**Base URL (Dev):** `http://localhost:8080`  
**Base URL (Prod):** `https://api.filesharing.com`

---

## 🚀 Các Tính Năng Chính

* **Upload Linh hoạt:** Hỗ trợ upload ẩn danh (public) hoặc upload có xác thực (private/quản lý file).
* **Các Lớp Bảo Mật:**
    * **Bảo vệ bằng Mật khẩu:** Mã hóa file với mật khẩu (tối thiểu 8 ký tự).
    * **Kiểm soát Truy cập (Access Control):** Giới hạn người download theo danh sách email cụ thể (Whitelist).
    * **Xác thực Hai yếu tố (TOTP):** Bảo vệ tài khoản người dùng với 2FA.
* **Quản lý Thời hạn:** Thiết lập ngày bắt đầu (`from`) và kết thúc (`to`) cho file. File sẽ tự động được dọn dẹp sau khi hết hạn.
* **Lịch sử Download:** Chủ sở hữu file (Owner) có thể xem nhật ký chi tiết (ai đã tải và tải khi nào).
* **Quản trị (Admin):** API dành cho admin để cấu hình chính sách hệ thống (system policy) và chạy các tác vụ dọn dẹp.

---

## 🔐 Xác thực (Authentication)

API sử dụng **Bearer Token (JWT)** để xác thực.

### Luồng Đăng nhập (Login Flow - Hỗ trợ 2FA)

1.  **Đăng nhập cơ bản:** Gửi request `POST` tới `/auth/login`.
    * *Nếu chưa bật 2FA:* Trả về ngay `accessToken`.
    * *Nếu đã bật 2FA:* Trả về `requireTOTP: true` và một `cid` (Challenge ID) của phiên đăng nhập.
2.  **Xác minh TOTP:** Gửi request `POST` tới `/auth/login/totp` kèm theo `cid` và mã 6 số (`code`) từ ứng dụng authenticator để nhận `accessToken`.

### Thiết lập 2FA
Để bật tính năng bảo mật 2 lớp cho tài khoản:
1.  Gọi `/auth/totp/setup` (cần Token đăng nhập) để nhận Secret Key và mã QR.
2.  Quét mã QR bằng ứng dụng xác thực (Google Authenticator, Authy...).
3.  Gọi `/auth/totp/verify` để kích hoạt.

---

## 🛠 Ví dụ Sử dụng API

### 1. Đăng ký tài khoản (User Registration)
Tạo một tài khoản người dùng tiêu chuẩn.

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "nam123",
    "email": "nam@example.com",
    "password": "SafePassword123!"
  }'
```

### 2\. Đăng nhập (Login)

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "nam@example.com",
    "password": "SafePassword123!"
  }'
```

### 3\. Upload File (Bảo mật)

Upload một file ở chế độ **riêng tư (private)**, **có mật khẩu**, và chỉ hiệu lực trong **3 ngày**.

*Yêu cầu phải có `Authorization` header.*

```bash
curl -X POST http://localhost:8080/files/upload \
  -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>" \
  -F "file=@./contract.pdf" \
  -F "isPublic=false" \
  -F "password=SecretPass123" \
  -F "availableFrom=2025-11-20T08:00:00Z" \
  -F "availableTo=2025-11-23T08:00:00Z" \
  -F "sharedWith=partner@example.com"
```

### 4\. Download File có Bảo vệ

Để tải file, người dùng phải vượt qua các lớp kiểm tra theo thứ tự:

1.  **Kiểm tra Thời gian** (File còn hiệu lực không?)
2.  **Kiểm tra Whitelist** (Email của bạn có nằm trong danh sách cho phép không?)
3.  **Kiểm tra Mật khẩu** (Header `X-File-Password`)

<!-- end list -->

```bash
curl -X GET http://localhost:8080/files/{shareToken}/download \
  -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>" \
  -H "X-File-Password: SecretPass123" \
  --output downloaded_contract.pdf
```

-----

## 📚 Tài liệu Tham khảo API (API Reference)

### Xác thực (`/auth`)

| Phương thức | Endpoint | Mô tả |
| :--- | :--- | :--- |
| `POST` | `/auth/register` | Đăng ký người dùng mới. |
| `POST` | `/auth/login` | Đăng nhập (trả về Token hoặc yêu cầu TOTP). |
| `POST` | `/auth/login/totp` | Hoàn tất đăng nhập nếu bật 2FA. |
| `POST` | `/auth/totp/setup` | Tạo secret/QR code cho 2FA. |
| `POST` | `/auth/totp/verify` | Kích hoạt 2FA. |
| `GET` | `/user` | Lấy thông tin profile hiện tại. |

### Quản lý File (`/files`)

| Phương thức | Endpoint | Mô tả |
| :--- | :--- | :--- |
| `POST` | `/files/upload` | Upload file (Token là tùy chọn nếu upload public). |
| `GET` | `/files/my` | Danh sách file của người dùng hiện tại. |
| `GET` | `/files/info/{id}` | Lấy metadata chi tiết (Chỉ Owner). |
| `DELETE` | `/files/info/{id}` | Xóa file (Chỉ Owner). |
| `GET` | `/files/{shareToken}` | Lấy thông tin file công khai. |
| `GET` | `/files/{shareToken}/download`| Tải nội dung file (binary). |
| `GET` | `/files/stats/{id}` | Xem thống kê lượt tải. |
| `GET` | `/files/download-history/{id}`| Xem nhật ký tải chi tiết. |

### Quản trị (`/admin`)

| Phương thức | Endpoint | Mô tả |
| :--- | :--- | :--- |
| `POST` | `/admin/cleanup` | Chạy lệnh xóa file hết hạn (Cần `X-Cron-Secret` hoặc Admin Token). |
| `GET` | `/admin/policy` | Xem cấu hình hệ thống (Giới hạn dung lượng, số ngày max). |
| `PATCH` | `/admin/policy` | Cập nhật cấu hình hệ thống. |

-----

## ⚠️ Mã Lỗi (Error Codes)

| Mã (Code) | Ý nghĩa | Nguyên nhân thường gặp |
| :--- | :--- | :--- |
| `400` | Bad Request | Thiếu file, mật khẩu quá yếu, hoặc khoảng thời gian không hợp lệ. |
| `401` | Unauthorized | Thiếu hoặc sai Bearer Token / Mã TOTP. |
| `403` | Forbidden | Sai mật khẩu file (`X-File-Password`), không nằm trong whitelist, hoặc không phải owner. |
| `410` | Gone | File đã hết hạn (Expired). |
| `423` | Locked | File đã upload nhưng chưa đến giờ hiệu lực (`availableFrom` ở tương lai). |
| `429` | Too Many Requests | Bị giới hạn tần suất gọi API (Rate limiting). |

## 📦 Chính sách Hệ thống (Mặc định)

  * **Dung lượng file tối đa:** 50MB (Admin có thể điều chỉnh)
  * **Thời hạn mặc định:** 7 Ngày
  * **Thời hạn tối đa:** 30 Ngày
  * **Chính sách mật khẩu:** Tối thiểu 8 ký tự
