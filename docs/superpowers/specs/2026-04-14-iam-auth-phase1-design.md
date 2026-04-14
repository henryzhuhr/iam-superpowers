# IAM 认证服务 - Phase 1 设计文档

> **阶段**：Phase 1（增强版核心认证）  
> **架构**：DDD 四层模块化单体  
> **日期**：2026-04-14

---

## 1. 概述

构建一个安全、可扩展的 IAM（身份和访问管理）认证服务的第一阶段。Phase 1 聚焦于核心认证能力：邮箱注册/登录、密码管理、JWT 双 Token 认证、基础用户资料、邮箱验证、简单租户模型和基础 RBAC 角色。

### 1.1 核心决策
| 决策项 | 选择 |
|--------|------|
| 开发模式 | 分阶段迭代 |
| 架构风格 | DDD 四层模块化单体（domain → repository → service → handler） |
| 技术栈 | Golang + Gin + PostgreSQL + Redis |
| 认证策略 | 双 Token（短时效 JWT Access Token + 长时效 Redis Refresh Token） |
| 数据库迁移 | golang-migrate |
| 密码存储 | bcrypt (cost factor 12) |

### 1.2 Phase 1 范围

- 邮箱注册 + 邮箱验证码验证
- 邮箱密码登录
- 双 Token 认证（Access Token 15min + Refresh Token 7days）
- 登出（注销 Refresh Token）
- 密码找回（邮件链接/验证码）
- 密码重置
- 密码修改
- 基础用户资料（头像、姓名、联系方式）
- 租户模型（tenant_id 字段，租户创建与管理）
- 基础 RBAC 角色（角色定义、用户-角色绑定）
- Web 管理控制台（Vue3 + TypeScript）

---

## 2. 目录结构

```
iam-superpowers/
├── cmd/
│   └── server/                 # 主入口 main.go
├── internal/
│   ├── auth/                   # Auth 领域
│   │   ├── domain/             # User 实体、Password 值对象
│   │   ├── repository/         # UserRepository 接口
│   │   ├── service/            # AuthService（注册/登录/刷新Token）
│   │   └── handler/            # HTTP 路由（Gin）
│   ├── user/                   # User 领域
│   │   ├── domain/
│   │   ├── repository/
│   │   └── service/
│   ├── tenant/                 # Tenant 领域
│   │   ├── domain/
│   │   ├── repository/
│   │   └── service/
│   ├── role/                   # Role 领域
│   │   ├── domain/
│   │   ├── repository/
│   │   └── service/
│   ├── audit/                  # Audit 领域（操作日志）
│   │   ├── domain/
│   │   ├── repository/
│   │   └── service/
│   └── common/                 # 共享基础设施
│       ├── database/           # PostgreSQL 连接
│       ├── redis/              # Redis 连接
│       ├── jwt/                # JWT 签发/验证
│       ├── email/              # 邮件发送（验证码）
│       └── errors/             # 统一错误处理
├── web/                      # 前端管理控制台（Vue3 + TypeScript）
│   ├── src/
│   │   ├── views/            # 页面组件
│   │   ├── components/       # 可复用组件
│   │   ├── router/           # 路由配置
│   │   ├── store/            # Pinia 状态管理
│   │   ├── api/              # API 调用封装
│   │   └── types/            # TypeScript 类型定义
│   ├── package.json
│   └── vite.config.ts
├── migrations/                 # golang-migrate SQL 文件
├── configs/                    # 配置文件
├── tests/e2e/                  # Python 端到端测试
├── docker-compose.yml
└── Makefile
```

---

## 3. 领域模型

### 3.1 User（用户）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| tenant_id | UUID | 所属租户 ID（外键） |
| email | VARCHAR(255) | 邮箱（唯一索引，按 tenant_id + email 联合唯一） |
| password_hash | VARCHAR(255) | bcrypt 哈希 |
| name | VARCHAR(100) | 用户姓名 |
| avatar_url | VARCHAR(500) | 头像 URL |
| status | ENUM | active / inactive / locked |
| email_verified | BOOLEAN | 是否已验证邮箱 |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

**领域行为**：
- `Create(email, password)`: 注册用户，生成 password_hash 和 email_verification_code
- `VerifyPassword(password)`: 校验密码
- `ChangePassword(oldPassword, newPassword)`: 修改密码
- `VerifyEmail(code)`: 验证邮箱

### 3.2 Tenant（租户）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| name | VARCHAR(100) | 租户名称 |
| unique_code | VARCHAR(50) | 租户唯一标识（唯一索引） |
| custom_domain | VARCHAR(255) | 自定义域名（可选） |
| status | ENUM | active / inactive |
| created_at | TIMESTAMP | 创建时间 |

### 3.3 Role（角色）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| tenant_id | UUID | 所属租户 ID（外键） |
| name | VARCHAR(50) | 角色名称 |
| description | VARCHAR(200) | 角色描述 |
| is_system | BOOLEAN | 是否系统预置 |
| created_at | TIMESTAMP | 创建时间 |

### 3.4 UserRole（用户-角色关联）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| user_id | UUID | 用户 ID（外键） |
| role_id | UUID | 角色 ID（外键） |
| tenant_id | UUID | 租户 ID（冗余字段，用于查询优化） |
| created_at | TIMESTAMP | 创建时间 |

### 3.5 AuditLog（操作日志）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| tenant_id | UUID | 租户 ID |
| user_id | UUID | 操作者用户 ID |
| action | VARCHAR(100) | 操作类型（如 user.create, role.assign） |
| target_type | VARCHAR(50) | 目标类型（如 user, role, tenant） |
| target_id | UUID | 目标 ID |
| details | JSONB | 操作详情 |
| ip_address | VARCHAR(45) | 操作者 IP |
| created_at | TIMESTAMP | 操作时间 |

---

## 4. API 端点

### 4.1 认证模块

| 端点 | 方法 | 说明 | 认证要求 |
|------|------|------|----------|
| `/api/v1/auth/register` | POST | 邮箱注册 | 无 |
| `/api/v1/auth/login` | POST | 邮箱密码登录 | 无 |
| `/api/v1/auth/refresh` | POST | 刷新 Access Token | Refresh Token |
| `/api/v1/auth/logout` | POST | 登出 | JWT |
| `/api/v1/auth/verify-email` | POST | 邮箱验证码验证 | 无 |
| `/api/v1/auth/forgot-password` | POST | 发送重置密码邮件 | 无 |
| `/api/v1/auth/reset-password` | POST | 重置密码 | 重置 Token |

### 4.2 用户模块

| 端点 | 方法 | 说明 | 认证要求 |
|------|------|------|----------|
| `/api/v1/users/me` | GET | 获取当前用户信息 | JWT |
| `/api/v1/users/me` | PUT | 更新当前用户信息 | JWT |
| `/api/v1/users/me/password` | PUT | 修改密码 | JWT |

### 4.3 管理控制台 API（管理员专用）

| 端点 | 方法 | 说明 | 认证要求 |
|------|------|------|----------|
| `/api/v1/admin/users` | GET | 用户列表（分页、搜索、过滤） | JWT + Admin Role |
| `/api/v1/admin/users/:id` | GET | 用户详情 | JWT + Admin Role |
| `/api/v1/admin/users/:id` | PUT | 编辑用户信息 | JWT + Admin Role |
| `/api/v1/admin/users/:id/status` | PUT | 禁用/启用用户 | JWT + Admin Role |
| `/api/v1/admin/users/:id/reset-password` | POST | 管理员重置密码 | JWT + Admin Role |
| `/api/v1/admin/users/:id/roles` | GET | 获取用户角色 | JWT + Admin Role |
| `/api/v1/admin/users/:id/roles` | PUT | 分配用户角色 | JWT + Admin Role |
| `/api/v1/admin/tenants` | GET | 租户列表 | JWT + Admin Role |
| `/api/v1/admin/tenants` | POST | 创建租户 | JWT + Admin Role |
| `/api/v1/admin/tenants/:id` | GET | 租户详情 | JWT + Admin Role |
| `/api/v1/admin/roles` | GET | 角色列表 | JWT + Admin Role |
| `/api/v1/admin/roles` | POST | 创建角色 | JWT + Admin Role |
| `/api/v1/admin/audit-logs` | GET | 操作日志列表 | JWT + Admin Role |

---

## 5. 关键数据流

### 5.1 注册流程

```
Client → POST /auth/register(email, password, tenant_code)
  → handler 校验输入格式
  → service 根据 tenant_code 查找租户
  → service 检查 email 是否已存在（同一租户内唯一）
  → domain: User.Create() 生成 password_hash(bcrypt)
  → 生成 6 位数字验证码存 Redis, key: email_verify:{email}, TTL 5min
  → repository: Save User
  → email service 发送验证码邮件
  → 返回 201 { user_id, message }
```

### 5.2 登录流程（双 Token）

```
Client → POST /auth/login(email, password)
  → handler 校验输入
  → service 根据 email 查找用户
  → service 校验密码（bcrypt Compare）
  → 若密码错误次数 >= 5，锁定账户
  → 生成 JWT Access Token (15min 过期，HS256 签名)
    Payload: sub(user_id), tid(tenant_id), roles[], iat, exp, jti
  → 生成 Refresh Token (UUID)
  → 存储 Refresh Token 到 Redis, key: refresh:{user_id}:{token}, TTL 7天
  → 返回 { access_token, refresh_token, expires_in }
```

### 5.3 Token 刷新流程

```
Client → POST /auth/refresh(refresh_token)
  → service 从 Redis 查找 refresh:{user_id}:{token}
  → 若不存在或过期，返回 401
  → 删除旧 Refresh Token（rotation）
  → 生成新 JWT Access Token
  → 生成新 Refresh Token 存 Redis
  → 返回 { access_token, refresh_token, expires_in }
```

### 5.4 登出流程

```
Client → POST /auth/logout (携带 JWT + Refresh Token)
  → handler 验证 JWT
  → service 删除 Redis 中的 Refresh Token
  → 可选：将当前 JWT 的 jti 加入黑名单（短期 TTL）
  → 返回 200
```

---

## 6. 技术实现

### 6.1 JWT 设计

- **算法**：HS256（Phase 1），后续可升级 RS256
- **Access Token**：15 分钟过期
- **Payload**：`sub`(user_id), `tid`(tenant_id), `roles`[], `iat`, `exp`, `jti`(JWT ID)
- **Refresh Token**：UUID，存 Redis，key 格式 `refresh:{user_id}:{token}`，TTL 7 天
- **黑名单**：Redis key `jwt_blacklist:{jti}`，TTL 同 Access Token 过期时间

### 6.2 密码安全

- 存储：`bcrypt`，cost factor = 12
- 强度校验：最少 8 位，包含大小写字母 + 数字
- 暴力破解防护：同一账户连续失败 5 次后锁定

### 6.3 邮箱验证

- 注册时生成 6 位数字验证码
- 存 Redis，key 格式 `email_verify:{email}`，TTL 5 分钟
- 验证通过后删除 Redis 记录，更新用户 `email_verified = true`

### 6.4 错误处理

统一 JSON 错误格式：
```json
{
  "code": "ERROR_CODE",
  "message": "用户可读信息",
  "details": {}
}
```

HTTP 状态码约定：
- `200/201`: 成功
- `400`: 参数错误
- `401`: 未认证
- `403`: 权限不足
- `404`: 资源不存在
- `409`: 冲突（如邮箱已注册）
- `429`: 限流
- `500`: 内部错误

### 6.5 中间件

| 中间件 | 功能 |
|--------|------|
| JWT Auth | 从 `Authorization: Bearer <token>` 提取并验证 JWT，注入 context |
| Tenant Resolver | 从 JWT 中获取 `tid`，注入到 context |
| Rate Limiter | 基于 IP 的滑动窗口限流（Redis） |
| CORS | 支持前端跨域 |
| Recovery | 统一 recover panic，返回 500 |

### 6.6 配置管理

- 使用 `viper` 管理配置
- 支持环境变量 + `.env` 文件（开发环境）
- 关键配置项：`DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`, `REDIS_HOST`, `REDIS_PORT`, `JWT_SECRET`, `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_FROM`

---

## 7. 数据库迁移

使用 `golang-migrate` 管理：

- 迁移文件放在 `migrations/` 目录
- 命名格式：`{version}_{name}.up.sql` / `{version}_{name}.down.sql`
- Phase 1 初始迁移：
  - `001_create_tenants.up.sql`
  - `002_create_users.up.sql`
  - `003_create_roles.up.sql`
  - `004_create_user_roles.up.sql`
  - `005_create_indexes.up.sql`

---

## 8. 测试策略

### 8.1 单元测试

- Go 标准 `testing` 包 + `testify` 断言
- 领域层（domain）重点测试业务逻辑
- Service 层使用 mock repository 测试

### 8.2 端到端测试

- Python + pytest + uv 包管理
- 通过 HTTP API 调用测试完整流程
- 测试用例覆盖：注册、登录、Token 刷新、登出、密码管理、邮箱验证

### 8.3 开发环境管理

**基础设施（docker-compose.yml）**：

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: iam_dev
      POSTGRES_USER: iam_user
      POSTGRES_PASSWORD: iam_pass
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  maildev:                    # 开发用 SMTP 服务器
    image: maildev/maildev
    ports:
      - "1080:1080"           # Web UI 查看邮件
      - "1025:1025"           # SMTP 端口
```

**环境变量（.env.example）**：

```env
# 数据库
DB_HOST=localhost
DB_PORT=5432
DB_NAME=iam_dev
DB_USER=iam_user
DB_PASSWORD=iam_pass
DB_SSLMODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET=change-me-in-production
JWT_ACCESS_TOKEN_TTL=900      # 15 分钟（秒）
JWT_REFRESH_TOKEN_TTL=604800  # 7 天（秒）

# SMTP（开发环境用 MailDev）
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM=noreply@iam.local
SMTP_USE_TLS=false

# 服务
SERVER_PORT=8080
```

开发者首次设置流程：
```bash
cp .env.example .env          # 复制配置模板
make up                       # 启动 PostgreSQL + Redis + MailDev
make migrate-up               # 执行数据库迁移
make run                      # 启动 Go 后端服务（localhost:8080）
make run-web                  # 启动前端开发服务器（localhost:5173）
```

**Makefile 命令**：

| 命令 | 功能 |
|------|------|
| `make up` | 启动 docker-compose 依赖服务 |
| `make down` | 停止依赖服务 |
| `make migrate-up` | 执行数据库向上迁移 |
| `make migrate-down` | 回滚最后一次迁移 |
| `make migrate-create <name>` | 创建新的迁移文件 |
| `make run` | 启动 Go 后端（热重载：`air`） |
| `make run-web` | 启动 Vue 前端开发服务器 |
| `make test` | 运行 Go 单元测试 |
| `make test-e2e` | 运行 Python 端到端测试 |
| `make build` | 构建 Go 二进制 |
| `make build-web` | 构建 Vue 前端静态文件 |

**数据库访问方式**：

- Go 服务通过 `database/sql` + `pgx` 驱动连接 PostgreSQL
- 连接池配置：最大连接数 25，空闲超时 5 分钟
- Redis 通过 `go-redis` 客户端连接
- 连接信息从 `viper` 读取环境变量
- 本地开发时 Go 进程直连 Docker 暴露的端口（localhost:5432, localhost:6379）

**热重载**：

- Go 后端使用 `air` 工具实现代码变更自动重启
- Vue 前端使用 Vite HMR（热模块替换）

---

## 9. 管理控制台（Web 前端）

### 9.1 技术栈

- **框架**：Vue 3 + TypeScript
- **构建工具**：Vite
- **UI 组件库**：Element Plus（或 Ant Design Vue）
- **状态管理**：Pinia
- **路由**：Vue Router
- **HTTP 客户端**：Axios

### 9.2 页面结构

| 页面 | 路由 | 说明 |
|------|------|------|
| 登录页 | `/login` | 管理员登录 |
| 仪表盘 | `/dashboard` | 概览统计（用户数、租户数、最近活动） |
| 用户管理 | `/users` | 用户列表（搜索、过滤、分页）、用户详情编辑 |
| 租户管理 | `/tenants` | 租户列表、创建租户、租户详情 |
| 角色管理 | `/roles` | 角色列表、创建角色、角色分配 |
| 操作日志 | `/audit-logs` | 操作日志列表（时间范围筛选） |

### 9.3 功能清单

**用户管理**：
- 用户列表（支持按邮箱搜索、按状态过滤、分页）
- 查看用户详情（基本信息、角色、状态）
- 编辑用户信息（姓名、头像、联系方式）
- 禁用/启用用户
- 管理员重置用户密码
- 分配/移除用户角色

**租户管理**：
- 租户列表（搜索、分页）
- 创建新租户（名称、unique_code、自定义域名）
- 查看租户详情

**角色管理**：
- 角色列表（按租户过滤）
- 创建自定义角色（名称、描述）
- 查看角色详情

**操作日志**：
- 日志列表（时间范围筛选、按操作者搜索）
- 查看操作详情

### 9.4 前端认证流程

- 登录页输入邮箱密码 → 调用 `/auth/login` → 获取 Access Token + Refresh Token
- Access Token 存 localStorage/sessionStorage
- Refresh Token 存 httpOnly Cookie 或 localStorage
- Axios 拦截器自动附加 `Authorization: Bearer <token>`
- 401 时自动调用 `/auth/refresh` 刷新 Token
- 刷新失败则跳转登录页

---

## 10. Phase 2+ 预留（不做实现）

以下能力在后续阶段添加，Phase 1 仅做结构预留：

- OAuth 2.0 / OIDC 第三方登录（Google, GitHub）
- SAML / OIDC 企业 SSO
- 2FA（TOTP、短信）
- 细粒度权限（permission）管理
- Webhook 事件通知
- 自定义域名路由
- 国际化（i18n）
- SDK 支持（JS、Python、Go）
