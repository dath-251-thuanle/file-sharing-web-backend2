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
| 1234567    | Đậu Minh Khôi        | NHÓM A, Class Diagram |
| 2311888    | Cao Vũ Hoàng Long    | NHÓM B |
| 2311906    | Nguyễn Hoàng Long    | NHÓM B |
| 2312955    | Đặng Hải Sơn         | NHÓM B, Use Case diagram  |

***NHÓM A: DATABASE DESIGN, API (Admin, System Management, File Management)**

***NHÓM B: API (Authentication, User Management, Statistics & Analytics)**

## Cấu trúc thư mục
```bash
/
├── cdm/
│   └── server/
│       ├── .env
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
│   │   │   ├── auth_dto.go
│   │   │   ├── file_dto.go
│   │   │   └── user_dto.go
│   │   ├── handlers/
│   │   │   ├── admin_handler.gp
│   │   │   ├── auth_handler.go
│   │   │   ├── file_handler.go
│   │   │   └── user_handler.go
│   │   └── routes/
│   │       ├── auth_routes.go
│   │       ├── router.go
│   │       └── user_routes.go
│   ├── app/
│   │   ├── app.go
│   │   ├── auth_module.go
│   │   └── user_module.go
│   ├── domain/
│   │   ├── auth.go
│   │   └── user.go
│   ├── infrastructure/
│   │   ├── database/
│   │   │   ├── connection.go
│   │   │   └── init.sql
│   │   └── jwt/
│   │       ├── interface.go
│   │       └── jwt.go
│   ├── middleware/
│   │   └── auth_middleware.go
│   ├── repository/
│   │   ├── auth_repository.go
│   │   ├── interface.go
│   │   └── user_repository.go
│   └── service/
│       ├── auth_service.go
│       ├── interface.go
│       └── user_service.go
├── pkg/
│   ├── utils/
│   │   ├── convert.go
│   │   ├── helper.go
│   │   └── response.go
│   └── validation/
│       ├── custom_validation.go
│       └── validation.go
├── test/
│   ├── auth_test.go
│   └── file_test.go
├── .DS_Store
├── .env
├── Makefile
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

Setup docker:
```
docker run --name postgres-db -e POSTGRES_PASSWORD=postgres -p 5435:5432 -d postgres
```

Tạo user và database:
```
docker exec -it postgres-db psql -U postgres

# Đã vào shell của postgres

CREATE USER haixon WITH PASSWORD "123456";

CREATE DATABASE "file-sharing";

\c "file-sharing";

GRANT ALL PRIVILEGES ON DATABASE "file-sharing" TO haixon;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO haixon;

exit;
```

Init và chạy server:
```
make            # Chạy các lênh SQL trong init.sql để tạo bảng (chỉ dùng 1 lần cho 1 docker)

make server     # Chạy server
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
