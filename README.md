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
| Student ID | Full Name            | Role                        
|:----------:|:--------------------:|:-------:|
| 1234567    | Đậu Minh Khôi        |         |
| 1234567    | Đậu Minh Khôi        |         | 
| 1234567    | Đậu Minh Khôi        |         |
| 1234567    | Đậu Minh Khôi        |         |
| 1234567    | Đậu Minh Khôi        |         |
| 1234567    | Đậu Minh Khôi        |         |

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

## Hướng dẫn cài đặt

## Workflow
**1. Clone repository**
```bash
git clone <repo-url>
```

**2. Create a new branch**
```bash
git checkout -b feature
```
**3. Adding your changes**

**4. Commit & Push your branch**
```bash
git add .
git commit -m "Your message"
git push origin feature/your-feuture
```

**5. Create your new pull request**