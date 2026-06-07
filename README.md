# sup-iam

sup-iam 是一个用 Go 语言编写的云原生身份识别与访问管理系统（IAM），采用**控制面与数据面分离**架构，为第三方用户及其程序化客户端提供统一的身份认证、密钥管理和基于 ACL 的访问控制能力。

## 架构设计

### 设计原则

- **控制面与数据面分离**：管理操作与鉴权决策解耦，互不影响
- **无状态鉴权**：数据面鉴权服务不保存用户登录状态，水平扩展无限制
- **最小权限原则**：默认拒绝所有访问，仅显式授权放行
- **可审计**：所有敏感操作和鉴权决策均记录审计日志

### 系统架构

```
                          ┌──────────────────┐
                          │   Load Balance   │
                          │   (LVS / Nginx)  │
                          └────────┬─────────┘
                                   │
          ┌────────────────────────┼────────────────────────┐
          ▼                        ▼                        ▼
┌──────────────────┐   ┌──────────────────┐   ┌──────────────────┐
│  iam-api-server  │   │ iam-auth-server  │   │    iam-pump      │
│   (控制面核心)     │   │  (数据面鉴权)     │   │  (数据采集)       │
│  端口: 8080      │   │  端口: 8443      │   │  端口: 7070      │
│  gRPC: 8081      │   │                  │   │                  │
└────────┬─────────┘   └────────┬─────────┘   └────────┬─────────┘
         │                      │                      │
         ▼                      ▼                      ▼
    ┌──────────┐          ┌──────────┐           ┌──────────┐
    │  MySQL   │          │  Redis   │           │ MongoDB  │
    │ (元数据)  │          │ (缓存/   │           │ (鉴权日志) │
    └──────────┘          │  日志)   │           └──────────┘
                          └──────────┘
```

## 核心服务

### iam-api-server（控制面核心服务）

管理 User/Secret/Policy 元数据，提供 RESTful API 和 gRPC 查询接口。

- 控制面用户认证（登录/注册/JWT Token 管理）
- User CRUD
- Secret/AK/SK CRUD 及密钥轮换
- Policy CRUD（Casbin DSL 策略定义）
- Secret-Policy 多对多绑定管理
- 审计日志查询

架构分层：Model → Controller → Service → Repository (MySQL)

### iam-auth-server（数据面鉴权服务）

负责请求签名校验和策略评估，是鉴权决策的执行端。

- 请求签名校验（HMAC-SHA256）
- AccessKey/SecretKey 验证
- Casbin ACL 模型策略评估
- 可解释的鉴权决策返回
- 全量本地缓存（Ristretto）消除网络开销
- 异步鉴权日志写入 Redis

**设计约束**：无状态、不保存用户登录状态、全量本地缓存保证最终一致性

### iam-pump（数据采集服务）

将鉴权决策日志从 Redis 缓冲层搬运到 MongoDB 持久化存储。

- Redis List 消费 → MongoDB 持久化
- 插件化 Pump 架构（内置 MongoPump）
- 多副本部署通过 Redlock 分布式锁防重复消费
- TTL 兜底防 OOM

## 鉴权流程

```
Client                     Resource Server           iam-auth-server
  │                              │                         │
  │  ① 使用 SK 生成 HMAC-SHA256 签名  │                         │
  │──────────────────────────────>│                         │
  │                              │  ② 携带 AK + 签名发起鉴权   │
  │                              │────────────────────────>│
  │                              │                         │
  │                              │  ③ 从本地缓存查找 SK       │
  │                              │  ④ 重新计算签名并比对      │
  │                              │  ⑤ Casbin ACL 策略评估    │
  │                              │  ⑥ 返回 Allow/Deny 决策   │
  │                              │<────────────────────────│
  │  ⑦ 根据决策放行或拒绝请求        │                         │
  │<──────────────────────────────│                         │
```

- **签名算法**：HMAC-SHA256
- **防重放**：请求时间戳校验（默认 ±300 秒误差，可配置）
- **策略模型**：Casbin ACL (sub, obj, act)
- **日志**：异步写入 → Redis 缓冲 → iam-pump 搬运 → MongoDB

## 技术栈

| 类别 | 技术 |
|------|------|
| **语言** | Go 1.25+ |
| **Web 框架** | Gin v1.12 |
| **ORM** | GORM v1.31 + MySQL |
| **缓存** | Ristretto（本地内存缓存）+ Redis |
| **策略引擎** | Casbin v2（ACL 模型） |
| **认证** | JWT（控制面）+ HMAC-SHA256（数据面） |
| **日志** | Zap + klog 兼容 |
| **消息** | Redis Pub/Sub（缓存变更通知） |
| **分布式锁** | Redsync |
| **序列化** | Protobuf / Msgpack |
| **服务发现** | etcd（可选） |
| **配置管理** | Viper + Pflag |
| **部署** | Docker + Kubernetes Helm Charts |

## 数据库

| 存储 | 用途 | 核心数据 |
|------|------|---------|
| **MySQL** | 元数据持久化 | users, secrets, policies, bindings, audit 表 |
| **Redis** | 缓存/日志缓冲/分布式锁 | 全量缓存、鉴权日志缓冲、JWT 黑名单、变更通知 Pub/Sub |
| **MongoDB** | 鉴权日志持久化 | capped collection（默认最大 5GB） |

核心实体关系：`User 1:N Secret M:N Policy`（通过 Binding 关联）

## API 概览

### 控制面 API（iam-api-server）`/api/v1/`

| 类别 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 认证 | POST | `/auth/login` | 用户登录 |
| 认证 | POST | `/auth/register` | 用户注册 |
| 认证 | POST | `/auth/logout` | 退出登录 |
| 认证 | POST | `/auth/refresh` | Token 刷新 |
| 用户 | GET/POST | `/users` | 用户列表/创建 |
| 用户 | GET/PUT/DELETE | `/users/{id}` | 用户详情/更新/删除 |
| 密钥 | GET/POST | `/secrets` | 密钥列表/创建 |
| 密钥 | PUT | `/secrets/{id}/rotate` | 密钥轮换 |
| 策略 | GET/POST | `/policies` | 策略列表/创建 |
| 策略 | GET/PUT/DELETE | `/policies/{id}` | 策略详情/更新/删除 |
| 绑定 | GET/POST | `/bindings` | 绑定列表/创建 |
| 审计 | GET | `/audits/policies` | 策略审计 |
| 审计 | GET | `/audits/bindings` | 绑定审计 |

### 数据面 API（iam-auth-server）`/auth/v1/`

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/verify` | 请求鉴权 |

## 快速开始

### 前置要求

- Go 1.25+
- MySQL / Redis / MongoDB
- Docker & Kubernetes（可选，用于容器化部署）

### 本地开发

```bash
# 构建所有服务
make all

# 检查代码质量
make fmt   # 格式化
make vet   # 静态检查
make test  # 运行测试（含竞态检测 + 覆盖率）

# 开发模式运行
make dev-api   # 启动 API Server
make dev-auth  # 启动 Auth Server
```

### Docker 构建

```bash
make docker-build          # 构建所有服务镜像
make docker-push           # 推送至 GHCR
```

### Kubernetes 部署

```bash
# 安装基础设施
make helm-mysql-install    # Bitnami MySQL
make helm-redis-install    # Bitnami Redis Cluster
make helm-mongodb-install  # Bitnami MongoDB

# 配置
make kube-cm-create        # 创建 ConfigMap
make kube-secret-create    # 创建 Secrets

# 部署 IAM
make helm-iam-install      # 安装 IAM Helm Chart
```

## CI/CD

- **Docker CI**（`.github/workflows/docker-ci.yml`）：push 到 main 或 PR 时自动运行测试、构建镜像并推送至 GHCR
- **AI Code Review**（`.github/workflows/ai-review.yaml`）：push 到 main 时自动对 Go 代码变更进行 AI 审查

## 文档

详细的系统设计文档位于 `docs/devel/design/` 目录：

- [系统架构设计说明书](docs/devel/design/系统架构设计说明书.md)
- [软件需求规格说明书](docs/devel/design/软件需求规格说明书.md)
- [API 服务架构设计说明书](docs/devel/design/API服务架构设计说明书.md)
- [鉴权设计说明书](docs/devel/design/鉴权设计说明书.md)
- [数据库设计说明书](docs/devel/design/数据库设计说明书.md)
- [全局配置设计说明书](docs/devel/design/全局配置设计说明书.md)
- [日志封装设计说明书](docs/devel/design/日志封装设计说明书.md)
- [全量缓存设计说明书](docs/devel/design/全量缓存设计说明书.md)
- [数据采集设计说明书](docs/devel/design/数据采集设计说明书.md)

API 规范文档位于 `docs/devel/api/` 目录，包含 OpenAPI/Swagger 定义。

## 许可

本项目基于 [GNU General Public License v3.0](LICENSE) 开源。
