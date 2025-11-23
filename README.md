# HỆ THỐNG CHIA SẺ FILE THÔNG QUA WEB (file-sharing-web-backend)
## Mục lục
[1. Tổng quan dự án](#tổng-quan-dự-án)

[2. Danh sách thành viên](#danh-sách-thành-viên)

[3. Cấu trúc thư mục](#cấu-trúc-thư-mục)

[4. Yêu cầu hệ thống](#yêu-cầu-hệ-thống)

[5. Hướng dẫn cài đặt](#hướng-dẫn-cài-đặt)

[6. Workflow](#workflow)

## Tổng quan dự án
Đây là repository chứa mã nguồn **Back-end** cho hệ thống chia sẻ file thông qua web, được xây dựng bằng Golang và sử dụng PostgreSQL.

Tính năng:
- Người dùng có thể upload các file lên hệ thống và chia sẻ chúng với người khác.
- Người dùng có thể thiết lập các thuộc tính sau khi chia sẻ file:
    - Có hiệu lực từ `from` đến `to`.
    - Có cài đặt mật khẩu (`password`)?
    - Có cài đặt `TOTP`?
    - Có thể chia sẻ với danh sách người dùng khác.

## Danh sách thành viên
| MSSV | Họ tên            | Công việc    |
| ----------:|:-------------------- |:------- |
| 2311159    | Lê Thanh Huy         | NHÓM A |
| 2311681    | Nguyễn Đình Khôi     | NHÓM A |
| 2311659    | Đậu Minh Khôi        | NHÓM A, Class Diagram |
| 2311888    | Cao Vũ Hoàng Long    | NHÓM B |
| 2311906    | Nguyễn Hoàng Long    | NHÓM B |
| 2312955    | Đặng Hải Sơn         | NHÓM B, Use Case diagram  |

***NHÓM A: DATABASE DESIGN, API (Admin, System Management, File Management)**

***NHÓM B: API (Authentication, User Management, Statistics & Analytics)**

## Cấu trúc thư mục

```bash
/file-sharing
├── cmd/
│   └── server/
│       └── main.go
├── config/
│   ├── app.yaml
│   └── config.go
├── docs/
│   ├── API_docs.md
│   └── README.md
├── internal/
│   ├── api/
│   │   ├── dto/
│   │   │   ├── admin_dto.go
│   │   │   ├── auth_dto.go
│   │   │   ├── file_dto.go
│   │   │   └── user_dto.go
│   │   ├── handlers/
│   │   │   ├── admin_handler.go
│   │   │   ├── auth_handler.go
│   │   │   ├── file_handler.go
│   │   │   └── user_handler.go
│   │   └── routes/
│   │       ├── admin_routes.go
│   │       ├── auth_routes.go
│   │       └── file_routes.go
│   │       └── router.go
│   │       └── user_routes.go
│   ├── app/
│   │   ├── admin_module.go
│   │   ├── app.go
│   │   ├── auth_module.go
│   │   ├── file_module.go
│   │   └── user_module.go
│   ├── domain/
│   │   ├── auth.go
│   │   ├── file_stat.go
│   │   ├── file.go
│   │   ├── share_with.go
│   │   └── user.go
│   ├── infrastructure/
│   │   ├── database/
│   │   │   ├── connection.go
│   │   │   └── init.sql
│   │   └── jwt/
│   │       ├── interface.go
│   │       └── jwt.go
│   │   └── storage/
│   │       ├── localstorage.go
│   │       └── storage.go
│   ├── middleware/
│   │   └── admin_middleware.go
│   │   └── auth_middleware.go
│   ├── repository/
│   │   ├── auth_repository.go
│   │   ├── file_repository.go
│   │   ├── interface.go
│   │   ├── share_repository.go
│   │   └── user_repository.go
│   └── service/
│       ├── admin_service.go
│       ├── auth_service.go
│       ├── file_service.go
│       ├── interface.go
│       └── user_service.go
├── pkg/
│   ├── utils/
│   │   ├── convert.go
│   │   ├── helper.go
│   │   └── random.go
│   │   └── response.go
│   └── validation/
│       ├── custom_validation.go
│       └── validation.go
├── test/
│   ├── auth_test.go
│   └── file_test.go
├── .age.key.pub
├── dev.enc
├── example.env
├── Makefile
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

## Yêu cầu hệ thống

- Cần có: Docker, PostresSQL, Golang (kèm thư viện Gin)
- Không bắt buộc:
    - Postman: kiểm thử API.

## Hướng dẫn cài đặt

Tạo file .env tại thư mục gốc:
```
cp example.env .env 
# Điền các thông số cấu hình (DB connection, JWT, v.v.).
```

Khởi chạy hệ thống bằng Docker:
```
docker compose up --build -d
```

Tạo bảng trong PostgreSQL:
```
docker exec -i postgres-db psql -U haixon -d file-sharing < internal/infrastructure/database/init.sql
```
Chạy server bằng Makefile (tùy chọn):
```
make server
```

Ở đây có thể dùng Postman hoặc curl để kiểm thử các API.

## Workflow

**1. Fork repository**

**2. Clone repository**
```bash
git clone <repo-url>
```

**3. Thêm các thay đổi**

**4. Commit và Push branch của bạn**
```bash
git add .
git commit -m "Tên commit"
git push origin <nhánh của bạn>
```

**5. Tạo pull request trên trang Github hoặc github-cli**

## 📄 Report đồ án

👉 [Xem Report tại đây](report/Report_DACNPM.pdf)
