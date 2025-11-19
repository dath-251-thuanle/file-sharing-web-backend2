# API list documentation (file-sharing-web-backend)

## Tổng quan
Hệ thống chia sẻ file qua web có các tính năng:
- Upload file và tạo link chia sẻ với người khác.

- Thiết lập quyền truy cập cho file (`public`, `password-protected`, `private`).

- Thiết lập thời gian hiệu lực cho file (`from`, `to`).

- Bảo vệ file bằng mật khẩu.

- Chia sẻ file với danh sách người dụng cụ thể.

- Tự động xóa file đã hết hạn.

## Danh sách các APIs
### 1. Authentication & User Management
|API                         |Mô tả                                       |
|:---------------------------|:-------------------------------------------|
|`POST /api/auth/register`   |Tạo tài khoản mới (không bắt buộc để upload)|
|`POST /api/auth/login`      |Đăng nhập để lấy JWT token                  |
|`POST /api/auth/login/totp` |Xác thực mã TOTP (6 chữ số)                 |
|`POST /api/auth/totp/setup` |Bật hoặc reset TOTP                         |
|`POST /api/auth/totp/verify`|Xác minh mã TOTP                            |
|`POST /api/auth/logout`     |Đăng xuất (client tự xóa token)             |

### 2. File Management
|API                                  |Mô tả                                                 |
|:------------------------------------|:-----------------------------------------------------|
|`POST /api/files/upload`             |Upload file mới và tạo link để chia sẻ                |
|`GET /api/files/:shareToken`         |Lấy thông tin file (sử dụng share token)              |
|`GET /api/files/:shareToken/download`|Tải file về                                           |
|`DELETE /api/files/:id`              |Xóa file (chỉ owner), anonymous uploader không thể xóa|
|`GET /api/files/my`                  |Lấy danh sách file của user đã đăng nhập              |

### 3. Admin / System Management
|API                      |Mô tả                     |
|:------------------------|:-------------------------|
|`POST /api/admin/cleanup`            |Xóa file hết hạn (Cron job hoặc Admin endpoint)       |
|`GET /api/admin/policy`  |Lấy cấu hình hệ thống     |
|`PATCH /api/admin/policy`|Cập nhật cấu hình hệ thống|