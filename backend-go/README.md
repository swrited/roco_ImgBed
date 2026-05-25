# 星诺图床 — 后端

基于 Go + Gin 框架构建的高性能图床 API 服务。

## 技术栈

- **语言**: Go 1.21+
- **Web 框架**: Gin
- **ORM**: GORM
- **数据库**: SQLite (内置) / MySQL (生产)
- **认证**: JWT + API Key
- **密码哈希**: bcrypt
- **图片处理**: Go 标准库 imaging

## 目录结构

```
internal/
├── config/           # 配置加载 & 数据库初始化
├── handler/          # HTTP 请求处理器
│   └── admin/        # 管理员接口
├── middleware/       # 认证、日志、CORS 等中间件
├── model/            # GORM 数据模型 & DTO
├── router/           # 路由定义
└── service/          # 业务逻辑
    ├── image/        # 图片处理 (缩略图、格式转换等)
    └── storage/      # 存储策略适配器
```

## 快速开始

```bash
# 下载依赖
go mod download

# 运行 (开发模式)
go run cmd/server/main.go

# 构建二进制文件
go build -o server cmd/server/main.go
```

服务默认监听 `:8000`。

## 配置

通过环境变量或 `.env` 文件配置：

```bash
cp .env.example .env
# 编辑 .env
```

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DB_CONNECTION` | `sqlite` | `sqlite` 或 `mysql` |
| `DB_HOST` | `127.0.0.1` | MySQL 主机 |
| `DB_PORT` | `3306` | MySQL 端口 |
| `DB_USERNAME` | `root` | MySQL 用户名 |
| `DB_PASSWORD` | - | MySQL 密码 |
| `DB_DATABASE` | `lskypro` | 数据库名 |
| `APP_PORT` | `8000` | 服务端口 |
| `APP_URL` | `http://localhost:8000` | 应用公网地址 |
| `JWT_SECRET` | `lskypro-secret-change-me` | **生产环境必须修改** |

## Makefile 命令

```bash
make build          # 构建二进制
make run            # 开发模式运行
make test           # 运行测试
make test-coverage  # 生成测试覆盖率报告
make lint           # 代码检查
make fmt            # 格式化代码
make clean          # 清理构建产物
make docker         # 构建 Docker 镜像
```

## API 概览

| 端点 | 说明 | 认证 |
|------|------|------|
| `POST /api/v1/tokens` | 用户登录 | 无 |
| `POST /api/v1/register` | 用户注册 | 无 |
| `POST /api/v1/upload` | 上传图片 | 可选 |
| `GET /api/v1/images` | 图片列表 | Bearer / API Key |
| `GET /api/v1/albums` | 相册列表 | Bearer / API Key |
| `GET /api/v1/gallery` | 公开画廊 | 可选 |
| `GET /api/v1/settings/public` | 公开设置 | 无 |
| `GET /api/v1/admin/console` | 控制台数据 | Admin |
| `GET /api/v1/admin/users` | 用户管理 | Admin |
| `GET /api/v1/admin/settings` | 系统设置 | Admin |

完整 API 文档请参阅前端项目的 API 文档页面。

## Docker

```bash
# 构建镜像
docker build -t sino-imgbed-backend .

# 运行容器
docker run -p 8000:8000 --env-file .env sino-imgbed-backend
```
