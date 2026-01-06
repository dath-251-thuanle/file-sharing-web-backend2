# Deployment Guide

## 1. Yêu cầu hệ thống

### Bắt buộc
- Docker Engine (>= 20.x)
- Docker Compose v2 (`docker compose`)
- Git

### Chỉ cần cho development (local)
- Go >= 1.25 (để chạy server trực tiếp, không bắt buộc nếu chỉ dùng Docker)
- Make (khuyến nghị, không bắt buộc)

---

## 2. Chuẩn bị .env

Clone reposistory
```bash
git clone <repo-url>
cd file-sharing-web-backend2
```

### 2.1. Local development

Tạo file `.env` từ mẫu:

```bash
cp .env.example .env
```

Mẫu env dành cho local development:

```
SERVER_PORT=8080
API_PORT=8080

DB_NAME=dev
DB_USER=dev
DB_PASSWORD=dev

GIN_MODE=debug
CORS_ALLOWED_ORIGINS=*
JWT_SECRET_KEY=dev-secret
```

### 2.2. Production
Tạo file .env trên server:

```
SERVER_PORT=8080
NGINX_PORT=80

DB_NAME=prod
DB_USER=prod_user
DB_PASSWORD=STRONG_PASSWORD

GIN_MODE=release
CORS_ALLOWED_ORIGINS=https://yourdomain.com
JWT_SECRET_KEY=SUPER_STRONG_SECRET
```

## 3. Chạy local (development)
### 3.1. Chạy bằng Docker

```bash
docker compose -f docker-compose.yml -f docker-compose.local.yml up -d --build
```

API được expose trực tiếp qua http://localhost:8080

### 3.2. Chạy local giống production (có nginx)
Yêu cầu: đăng nhập ghcr.io (xem phần bên [dưới](#42-đăng-nhập-vào-ghcrio))
```bash
docker compose up -d
```
nginx chạy ở http://localhost

### 3.3 Chạy bằng Makefile
```bash
make docker-up
```


## 4. Deploy lên production
### 4.1. Chuẩn bị server
Cài Docker + Docker Compose

### 4.2. Đăng nhập vào ghcr.io:
- Truy cập tại [đây](https://github.com/settings/tokens) để tạo PAT (Personal Access Token)
- Chọn Generate new token (classic)
- Tick các quyền: repo và read:packages
- Sau khi generate token, copy và lưu lại
- Sau đó đăng nhập ghcr.io

```bash
echo <YOUR_GITHUB_PAT> | docker login ghcr.io -u <GITHUB_USERNAME> --password-stdin
```

### 4.3. Deploy
Chạy duy nhất một lệnh:

```bash
docker compose up -d
```

## 5. Quản lý container
Xem logs API
```bash
docker compose logs -f api
```
Stop services (giữ dữ liệu)
```bash
docker compose down
```
Reset database (⚠️ chỉ dùng cho local)
```bash
docker compose down -v
docker compose up -d
```