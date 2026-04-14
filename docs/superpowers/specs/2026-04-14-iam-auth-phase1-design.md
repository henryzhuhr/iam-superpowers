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
│   └── common/                 # 共享基础设施
│       ├── database/           # PostgreSQL 连接
│       ├── redis/              # Redis 连接
│       ├── jwt/                # JWT 签发/验证
│       ├── email/              # 邮件发送（验证码）
│       └── errors/             # 统一错误处理
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

### 8.3 部署

- `docker-compose.yml`：PostgreSQL + Redis
- `Makefile`：构建、测试、迁移、运行

---

## 9. Phase 2+ 预留（不做实现）

以下能力在后续阶段添加，Phase 1 仅做结构预留：

- OAuth 2.0 / OIDC 第三方登录（Google, GitHub）
- SAML / OIDC 企业 SSO
- 2FA（TOTP、短信）
- 细粒度权限（permission）管理
- Webhook 事件通知
- 自定义域名路由
- 审计日志
- 国际化（i18n）
