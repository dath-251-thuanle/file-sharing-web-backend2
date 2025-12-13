# 📁 File Sharing Web Backend

Hệ thống chia sẻ file tạm thời qua web, được xây dựng bằng **Golang** với framework **Gin** và database **PostgreSQL**.

## 📋 Mục lục
- [Tính năng](#-tính-năng)
- [Tech Stack](#-tech-stack)
- [Cấu trúc dự án](#-cấu-trúc-dự-án)
- [Cài đặt và Chạy](#-cài-đặt-và-chạy)
- [API Documentation](#-api-documentation)
- [Makefile Commands](#-makefile-commands)
- [Thành viên nhóm](#-thành-viên-nhóm)
- [Report đồ án](#-report-đồ-án)

---

## ✨ Tính năng

- **Upload & Share**: Upload file và tạo link chia sẻ với share token
- **Thời gian hiệu lực**: Thiết lập `availableFrom` và `availableTo` cho file
- **Bảo mật đa lớp**:
  - Password protection
  - Whitelist người dùng (sharedWith)
  - TOTP/2FA cho tài khoản
- **File preview**: Xem trước file trực tiếp trong browser
- **Thống kê download**: Theo dõi lịch sử tải về chi tiết
- **Anonymous upload**: Hỗ trợ upload không cần đăng nhập

---

## 🛠 Tech Stack

| Component | Technology |
|-----------|------------|
| **Language** | Go 1.25+ |
| **Framework** | Gin |
| **Database** | PostgreSQL 17 |
| **Authentication** | JWT |
| **2FA** | TOTP (Google Authenticator) |
| **Storage** | Local filesystem |
| **Container** | Docker & Docker Compose |

---

## 📂 Cấu trúc dự án

```
file-sharing-web-backend/
├── cmd/server/           # Entry point
│   └── main.go
├── config/               # Configuration
├── docs/                 # Documentation
│   ├── API_docs.md
│   └── openapi.yaml
├── internal/
│   ├── api/              # API layer
│   │   ├── dto/          # Data Transfer Objects
│   │   ├── handlers/     # Request handlers
│   │   └── routes/       # Route definitions
│   ├── app/              # Application modules
│   ├── domain/           # Domain models
│   ├── infrastructure/   # External services
│   │   ├── database/     # DB connection & schema
│   │   ├── jwt/          # JWT service
│   │   └── storage/      # File storage
│   ├── middleware/       # Auth & Admin middleware
│   ├── repository/       # Data access layer
│   └── service/          # Business logic
├── pkg/                  # Shared packages
│   ├── utils/
│   └── validation/
├── test/                 # Tests
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── go.mod
```

---

## 🚀 Cài đặt và Chạy

### Yêu cầu
- Docker & Docker Compose
- Go 1.25+ (nếu chạy local)

### Sử dụng Docker (Recommended)

```bash
# 1. Clone repository
git clone <repo-url>
cd file-sharing-web-backend

# 2. Tạo file .env
cp example.env .env
# Chỉnh sửa các thông số trong .env

# 3. Khởi chạy
docker compose up -d

# Database tự động được tạo từ init.sql
# Server chạy tại http://localhost:8080
```

### Chạy Local (Development)

```bash
# 1. Đảm bảo PostgreSQL đang chạy
docker compose up -d db

# 2. Chạy server
make server
```

### Reset Database

```bash
make docker-reset
```

---

## 📖 API Documentation

Chi tiết về tất cả endpoints có trong:
- **[API_docs.md](docs/API_docs.md)** - Tài liệu tổng quan
- **[openapi.yaml](docs/openapi.yaml)** - OpenAPI 3.0 specification

### Quick Overview

| Category | Endpoints |
|----------|-----------|
| **Auth** | `POST /auth/register`, `/auth/login`, `/auth/logout`, `/auth/totp/*` |
| **User** | `GET /user` |
| **Files** | `POST /files/upload`, `GET /files/my`, `GET /files/available`, `GET /files/{shareToken}/download`, `GET /files/{shareToken}/preview` |
| **Admin** | `POST /admin/cleanup`, `GET/PATCH /admin/policy` |

### Base URL
- Development: `http://localhost:8080`
- Production: `https://api.filesharing-hcmut.com`

---

## 🔧 Makefile Commands

```bash
make server        # Chạy server development
make docker-reset  # Reset database (xóa data + khởi động lại)
make docker-logs   # Xem logs API
make test          # Chạy tests
make clean         # Xóa build artifacts
make deps          # Tải dependencies
```

---

## 👥 Thành viên nhóm

| MSSV | Họ tên | Công việc |
|------|--------|-----------|
| 2311159 | Lê Thanh Huy | Nhóm A |
| 2311681 | Nguyễn Đình Khôi | Nhóm A |
| 2311659 | Đậu Minh Khôi | Nhóm A, Class Diagram |
| 2311888 | Cao Vũ Hoàng Long | Nhóm B |
| 2311906 | Nguyễn Hoàng Long | Nhóm B |
| 2312955 | Đặng Hải Sơn | Nhóm B, Use Case Diagram |

**Nhóm A:** Database Design, API (Admin, System Management, File Management)

**Nhóm B:** API (Authentication, User Management, Statistics & Analytics)

---

## 📄 Report đồ án

👉 [Xem Report tại đây](report/Report_DACNPM.pdf)

---

## 📝 License

MIT License
