# IAM API Server 配置

## 目录结构

| 文件 | 作用 |
|------|------|
| `config.go` | 定义配置结构体 `AppConfig`，包含 Server、JWT、MySQL、Redis、gRPC、Log 六个子配置模块 |
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
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
  black_list_ttl: 1h

jwt:
  secret_key: "your-secret-key"
  access_token_expire_time: 1h
  refresh_token_expire_time: 168h
  user_id_key: "user_id"
  token_lookup: "header:Authorization"
  issuer: "iam-apiserver"
  skip_paths:
    - "/health"
    - "/api/v1/auth/login"
    - "/api/v1/auth/register"

mysql:
  host: "127.0.0.1"
  port: 3306
  username: "root"
  password: "your-password"
  database_name: "iam"
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 1h
  max_retries: 3

redis:
  host: "127.0.0.1"
  port: 6379
  password: ""
  database_name: 0
  health_check_interval: 10s
  pool_size: 10
  min_idle_conns: 5
  max_idle_conns: 10
  conn_max_idle_time: 5m
  conn_max_lifetime: 1h
  dial_timeout: 5s
  read_timeout: 3s
  write_timeout: 3s
  pool_timeout: 4s

log:
  level: "info"
  format: "console"
  output-paths:
    - stdout
  error-output-paths:
    - stderr
  disable-caller: false
  disable-stacktrace: false
  enable-color: false
  development: false
  name: ""
```

### 方式二：环境变量（推荐用于生产环境）

所有配置项均可通过 `IAM_` 前缀的环境变量覆盖，支持按配置层级自动映射（Viper 自动将 `server.host` 映射到 `IAM_SERVER_HOST`）：

**Windows (cmd)**
```cmd
set IAM_SERVER_PORT=8080
set IAM_JWT_SECRET_KEY=my-production-secret-key
set IAM_MYSQL_PASSWORD=my-db-password
set IAM_REDIS_PASSWORD=my-redis-password
set IAM_LOG_LEVEL=debug
set IAM_LOG_FORMAT=json
```

**Linux/macOS**
```bash
export IAM_SERVER_PORT=8080
export IAM_JWT_SECRET_KEY=my-production-secret-key
export IAM_MYSQL_PASSWORD=my-db-password
export IAM_REDIS_PASSWORD=my-redis-password
export IAM_LOG_LEVEL=debug
export IAM_LOG_FORMAT=json
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
| `read_timeout` | `IAM_SERVER_READ_TIMEOUT` | `30s` | 读取超时 |
| `write_timeout` | `IAM_SERVER_WRITE_TIMEOUT` | `30s` | 写入超时 |
| `idle_timeout` | `IAM_SERVER_IDLE_TIMEOUT` | `120s` | 空闲超时 |
| `black_list_ttl` | `IAM_SERVER_BLACK_LIST_TTL` | `1h` | 黑名单过期时间 |
| `grace_timeout` | `IAM_SERVER_GRACE_TIMEOUT` | `10s` | 优雅关闭超时 |

### JWT 配置（`jwt`）

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|---------|--------|------|
| `secret_key` | `IAM_JWT_SECRET_KEY` | （无默认值） | JWT 签名密钥，**生产环境必须修改** |
| `access_token_expire_time` | `IAM_ACCESS_TOKEN_EXPIRE_TIME` | `1h` | Access Token 过期时间 |
| `refresh_token_expire_time` | `IAM_REFRESH_TOKEN_EXPIRE_TIME` | `168h` | Refresh Token 过期时间 |
| `user_id_key` | `IAM_JWT_USER_ID_KEY` | `user_id` | 用户ID在Token中的键名 |
| `token_lookup` | `IAM_JWT_TOKEN_LOOKUP` | `header:Authorization` | Token 提取方式 |
| `issuer` | `IAM_JWT_ISSUER` | `iam-apiserver` | JWT 签发者 |
| `skip_paths` | `IAM_JWT_SKIP_PATHS` | `["/health","/api/v1/auth/login","/api/v1/auth/register"]` | JWT 豁免路径列表 |

### gRPC 配置（`grpc`）

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|---------|--------|------|
| `host` | `IAM_GRPC_HOST` | `0.0.0.0` | 监听地址 |
| `port` | `IAM_GRPC_PORT` | `9090` | 监听端口 |
| `etcd_server_discovery` | `IAM_GRPC_ETCD_SERVER_DISCOVERY` | `false` | etcd 服务发现开关 |
| `etcd_host` | `IAM_GRPC_ETCD_HOST` | `127.0.0.1` | etcd 服务地址 |
| `etcd_port` | `IAM_GRPC_ETCD_PORT` | `2379` | etcd 服务端口 |
| `service_name` | `IAM_GRPC_SERVICE_NAME` | `""` | gRPC 服务名称（开启 etcd 服务发现时需要配置） |
| `lease_ttl` | `IAM_GRPC_LEASE_TTL` | `10s` | etcd 租约过期时间 |
| `service_address` | `IAM_GRPC_SERVICE_ADDRESS` | `""` | gRPC 服务注册地址（为空时自动获取本机 IP+grpc端口） |

### MySQL 配置（`mysql`）

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|---------|--------|------|
| `host` | `IAM_MYSQL_HOST` | `127.0.0.1` | 数据库地址 |
| `port` | `IAM_MYSQL_PORT` | `3306` | 数据库端口 |
| `username` | `IAM_MYSQL_USERNAME` | `root` | 数据库用户名 |
| `password` | `IAM_MYSQL_PASSWORD` | （空） | 数据库密码，**生产环境必须通过环境变量注入** |
| `database_name` | `IAM_MYSQL_DATABASE_NAME` | `iam` | 数据库名称 |
| `max_idle_conns` | `IAM_MYSQL_MAX_IDLE_CONNS` | `10` | 最大空闲连接数 |
| `max_open_conns` | `IAM_MYSQL_MAX_OPEN_CONNS` | `100` | 最大打开连接数 |
| `conn_max_lifetime` | `IAM_MYSQL_CONN_MAX_LIFETIME` | `1h` | 连接最大存活时间 |
| `max_retries` | `IAM_MYSQL_MAX_RETRIES` | `3` | 最大重试次数 |

### Redis 配置（`redis`）

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|---------|--------|------|
| `host` | `IAM_REDIS_HOST` | `127.0.0.1` | Redis 地址 |
| `port` | `IAM_REDIS_PORT` | `6379` | Redis 端口 |
| `password` | `IAM_REDIS_PASSWORD` | （空） | Redis 密码 |
| `database_name` | `IAM_REDIS_DATABASE_NAME` | `0` | Redis 数据库编号 |
| `health_check_interval` | `IAM_REDIS_HEALTH_CHECK_INTERVAL` | `10s` | 健康检查间隔 |
| `pool_size` | `IAM_REDIS_POOL_SIZE` | `10` | 连接池大小 |
| `min_idle_conns` | `IAM_REDIS_MIN_IDLE_CONNS` | `5` | 最小空闲连接数 |
| `max_idle_conns` | `IAM_REDIS_MAX_IDLE_CONNS` | `10` | 最大空闲连接数 |
| `conn_max_idle_time` | `IAM_REDIS_CONN_MAX_IDLE_TIME` | `5m` | 连接最大空闲时间 |
| `conn_max_lifetime` | `IAM_REDIS_CONN_MAX_LIFETIME` | `1h` | 连接最大存活时间 |
| `dial_timeout` | `IAM_REDIS_DIAL_TIMEOUT` | `5s` | 拨号超时 |
| `read_timeout` | `IAM_REDIS_READ_TIMEOUT` | `3s` | 读取超时 |
| `write_timeout` | `IAM_REDIS_WRITE_TIMEOUT` | `3s` | 写入超时 |
| `pool_timeout` | `IAM_REDIS_POOL_TIMEOUT` | `4s` | 连接池获取超时 |

### 日志配置（`log`）

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|---------|--------|------|
| `level` | `IAM_LOG_LEVEL` | `info` | 日志级别：`debug` / `info` / `warn` / `error` / `panic` / `fatal` |
| `format` | `IAM_LOG_FORMAT` | `console` | 输出格式：`console` / `json` |
| `output-paths` | `IAM_LOG_OUTPUT_PATHS` | `["stdout"]` | 日志输出路径 |
| `error-output-paths` | `IAM_LOG_ERROR_OUTPUT_PATHS` | `["stderr"]` | 错误日志输出路径 |
| `disable-caller` | `IAM_LOG_DISABLE_CALLER` | `false` | 是否禁用 caller 信息 |
| `disable-stacktrace` | `IAM_LOG_DISABLE_STACKTRACE` | `false` | 是否禁用堆栈跟踪 |
| `enable-color` | `IAM_LOG_ENABLE_COLOR` | `false` | 是否启用颜色输出（仅 console 格式有效） |
| `development` | `IAM_LOG_DEVELOPMENT` | `false` | 是否为开发模式 |
| `name` | `IAM_LOG_NAME` | `""` | 日志记录器名称 |

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
