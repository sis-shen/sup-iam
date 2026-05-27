# IAM API Server 配置

## 目录结构

| 文件 | 作用 |
|------|------|
| `config.go` | 定义配置结构体 `AppConfig`，包含 Server、JWT、MySQL、Redis 四个子配置模块 |
| `loader.go` | 配置加载引擎，负责读取配置文件、环境变量、设置默认值及配置验证 |
| `config.yaml` | 默认 YAML 配置文件（开发环境），含所有配置项的默认值和注释说明 |

---

## 配置框架

使用 **[spf13/viper](https://github.com/spf13/viper)** v1.21.0 作为配置管理库。Viper 支持多层级配置源，按优先级从低到高排列：

1. **默认值** — Go 代码中硬编码的默认值（`setDefaults` 函数）
2. **配置文件** — YAML 格式的配置文件
3. **环境变量** — 以 `IAM_` 为前缀的环境变量（优先级最高）

---

## 配置方式

### 方式一：修改配置文件

直接编辑 `config.yaml`，修改对应配置项即可：

```yaml
server:
  port: 8888
  mode: "release"

jwt:
  secret_key: "your-secret-key"
  access_token_expire_time: 60  # 单位：分钟
  refresh_token_expire_time: 7  # 单位：天

mysql:
  host: "127.0.0.1"
  port: 3306
  username: "root"
  password: "your-password"
  database_name: "iam"

redis:
  host: "127.0.0.1"
  port: 6379
  password: ""
  database_name: 0
```

### 方式二：环境变量（推荐用于生产环境）

所有配置项均可通过 `IAM_` 前缀的环境变量覆盖，支持按配置层级自动映射（Viper 自动将 `server.host` 映射到 `IAM_SERVER_HOST`）：

**Windows (cmd)**
```cmd
set IAM_SERVER_PORT=8080
set IAM_JWT_SECRET_KEY=my-production-secret-key
set IAM_MYSQL_PASSWORD=my-db-password
set IAM_REDIS_PASSWORD=my-redis-password
```

**Linux/macOS**
```bash
export IAM_SERVER_PORT=8080
export IAM_JWT_SECRET_KEY=my-production-secret-key
export IAM_MYSQL_PASSWORD=my-db-password
export IAM_REDIS_PASSWORD=my-redis-password
```

### 方式三：指定自定义配置文件路径

```bash
export IAM_CONFIG_FILE=/etc/iam/custom-config.yaml
```

### 方式四：按环境自动加载

设置环境变量 `IAM_ENV` 或 `GO_ENV` 后，加载器会自动搜索对应环境的配置文件：

| 环境变量值 | 加载的文件 |
|------------|-----------|
| `production` / `prod` | `config.prod.yaml` |
| `development` / `dev` | `config.dev.yaml` |
| `test` | `config.test.yaml` |
| 未设置或其他值 | `config.yaml` |

配置文件搜索路径顺序：
1. `.`（当前工作目录）
2. `config/`
3. `../config/`
4. `/etc/iam/`

---

## 配置项完整说明

### 服务器配置（`server`）

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|---------|--------|------|
| `host` | `IAM_SERVER_HOST` | `0.0.0.0` | 监听地址 |
| `port` | `IAM_SERVER_PORT` | `8888` | 监听端口 |
| `mode` | `IAM_SERVER_MODE` | `debug` | 运行模式：`debug` / `release` / `test` |
| `read_timeout` | `IAM_SERVER_READ_TIMEOUT` | `30` | 读取超时（秒） |
| `write_timeout` | `IAM_SERVER_WRITE_TIMEOUT` | `30` | 写入超时（秒） |

### JWT 配置（`jwt`）

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|---------|--------|------|
| `secret_key` | `IAM_JWT_SECRET_KEY` | （无默认值） | JWT 签名密钥，**生产环境必须修改** |
| `access_token_expire_time` | `IAM_ACCESS_TOKEN_EXPIRE_TIME` | `60` | Access Token 过期时间（分钟） |
| `refresh_token_expire_time` | `IAM_REFRESH_TOKEN_EXPIRE_TIME` | `7` | Refresh Token 过期时间（天） |

### MySQL 配置（`mysql`）

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|---------|--------|------|
| `host` | `IAM_MYSQL_HOST` | `127.0.0.1` | 数据库地址 |
| `port` | `IAM_MYSQL_PORT` | `3306` | 数据库端口 |
| `username` | `IAM_MYSQL_USERNAME` | `root` | 数据库用户名 |
| `password` | `IAM_MYSQL_PASSWORD` | （空） | 数据库密码，**生产环境必须通过环境变量注入** |
| `database_name` | `IAM_MYSQL_DATABASE_NAME` | `iam` | 数据库名称 |

### Redis 配置（`redis`）

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|---------|--------|------|
| `host` | `IAM_REDIS_HOST` | `127.0.0.1` | Redis 地址 |
| `port` | `IAM_REDIS_PORT` | `6379` | Redis 端口 |
| `password` | `IAM_REDIS_PASSWORD` | （空） | Redis 密码 |
| `database_name` | `IAM_REDIS_DATABASE_NAME` | `0` | Redis 数据库编号 |

---

## 配置验证规则

加载配置时，`validateConfig` 函数会执行以下校验：

- `jwt.secret_key` — **不能为空**（生产环境必须设置）
- `mysql.password` — **不能为空**（生产环境必须设置）
- `server.port`、`mysql.port`、`redis.port` — 必须在 1～65535 范围内
- `server.mode` — 必须是 `debug`、`release` 或 `test` 之一

---

## 配置加载流程

```
Load(configPath)
  ├── 设置默认值 (setDefaults)
  ├── 加载配置文件 (LoadConfigFile)
  │     ├── 自动检测路径 (autoConfigFilePath)
  │     │     ├── 优先使用 IAM_CONFIG_FILE 环境变量
  │     │     ├── 根据 IAM_ENV / GO_ENV 选择配置文件名
  │     │     └── 按路径顺序搜索配置文件
  │     └── 解析 YAML 到 Viper
  ├── 绑定环境变量 (IAM_ 前缀)
  ├── 解组到 AppConfig 结构体 (Unmarshal)
  └── 配置验证 (validateConfig)
```
