# IAM Auth Server 配置

## 目录结构

| 文件 | 作用 |
|------|------|
| `config.go` | 定义配置结构体 `Config`，包含 Server、gRPC、Redis、Log、Analytics、Cache 六个子配置模块 |
| `loader.go` | 配置加载引擎，负责读取配置文件、环境变量、设置默认值及配置验证 |
| `config.yaml` | 默认 YAML 配置文件（开发环境），含所有配置项的默认值和注释说明 |

---

## 配置框架

使用 **[spf13/viper](https://github.com/spf13/viper)** v1.21.0 作为配置管理库。Viper 支持多层级配置源，按优先级从低到高排列：

1. **默认值** — Go 代码中硬编码的默认值（`NewConfig` 及各子模块 `NewXxxOptions` 函数）
2. **配置文件** — YAML 格式的配置文件
3. **环境变量** — 以 `IAM_` 为前缀的环境变量（优先级最高）

---

## 配置方式

### 方式一：修改配置文件

直接编辑 `config.yaml`，修改对应配置项即可：

```yaml
server:
  host: "0.0.0.0"
  port: 8889
  mode: "debug"
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
  grace_timeout: 10s
  enable_redis_sink: false
  redis_key_prefix: "iam:log:"
  sink_level: ""

grpc:
  host: "127.0.0.1"
  port: 9090
  etcd_server_discovery: false
  service_name: ""

redis:
  host: "127.0.0.1"
  port: 6379
  addrs: []
  username: ""
  password: ""
  database_name: 0
  health_check_interval: 10s
  enable_cluster: false
  master_name: ""
  use_ssl: false
  ssl_insecure_skip_verify: false
  pool_size: 10
  max_active_conns: 100
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

analytics:
  enable: true
  enable_detail_recording: true
  analytics_key_name: "iam-system-analytics"
  pool_size: 10
  record_buffer_size: 1024
  flush_interval: 10s
  storage_expiration: 24h

cache:
  num_counters: 0
  max_cost: 0
  buffer_items: 0
```

### 方式二：环境变量（推荐用于生产环境）

所有配置项均可通过 `IAM_` 前缀的环境变量覆盖，支持按配置层级自动映射（Viper 自动将 `server.host` 映射到 `IAM_SERVER_HOST`）：

**Windows (cmd)**
```cmd
set IAM_SERVER_PORT=8080
set IAM_GRPC_HOST=192.168.1.100
```

**Linux/macOS**
```bash
export IAM_SERVER_PORT=8080
export IAM_GRPC_HOST=192.168.1.100
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
| `port` | `IAM_SERVER_PORT` | `8080` | 监听端口 |
| `mode` | `IAM_SERVER_MODE` | `debug` | 运行模式：`debug` / `release` / `test` |
| `read_timeout` | `IAM_SERVER_READ_TIMEOUT` | `15s` | 读取超时 |
| `write_timeout` | `IAM_SERVER_WRITE_TIMEOUT` | `15s` | 写入超时 |
| `idle_timeout` | `IAM_SERVER_IDLE_TIMEOUT` | `60s` | 空闲超时 |
| `grace_timeout` | `IAM_SERVER_GRACE_TIMEOUT` | `60s` | 优雅关闭超时 |
| `enable_redis_sink` | `IAM_SERVER_ENABLE_REDIS_SINK` | `false` | 日志是否同步写入 Redis |
| `redis_key_prefix` | `IAM_SERVER_REDIS_KEY_PREFIX` | `iam-auth` | 日志在 Redis 中的 key 前缀 |
| `sink_level` | `IAM_SERVER_SINK_LEVEL` | `info` | 日志同步级别（为空则同步所有级别） |
| `load_cache_ttl` | 无 | 0（默认不设置） | 缓存加载 TTL |

### gRPC 配置（`grpc`）

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|---------|--------|------|
| `host` | `IAM_GRPC_HOST` | `0.0.0.0` | IAM API Server gRPC 地址 |
| `port` | `IAM_GRPC_PORT` | `8080` | IAM API Server gRPC 端口 |
| `etcd_server_discovery` | `IAM_GRPC_ETCD_SERVER_DISCOVERY` | `false` | 是否启用 etcd 服务发现 |
| `service_name` | `IAM_GRPC_SERVICE_NAME` | `""` | gRPC 服务名称（etcd 服务发现时使用） |

### Redis 配置（`redis`）

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|---------|--------|------|
| `host` | `IAM_REDIS_HOST` | `127.0.0.1` | Redis 地址 |
| `port` | `IAM_REDIS_PORT` | `6379` | Redis 端口 |
| `addrs` | `IAM_REDIS_ADDRS` | `[]` | Redis 集群节点地址列表 |
| `username` | `IAM_REDIS_USERNAME` | `""` | Redis 用户名 |
| `password` | `IAM_REDIS_PASSWORD` | `""` | Redis 密码 |
| `database_name` | `IAM_REDIS_DATABASE_NAME` | `0` | Redis 数据库编号 |
| `health_check_interval` | `IAM_REDIS_HEALTH_CHECK_INTERVAL` | `5s` | 健康检查间隔 |
| `enable_cluster` | `IAM_REDIS_ENABLE_CLUSTER` | `false` | 是否启用集群模式 |
| `master_name` | `IAM_REDIS_MASTER_NAME` | `""` | Sentinel master 名称 |
| `use_ssl` | `IAM_REDIS_USE_SSL` | `false` | 是否启用 SSL 连接 |
| `ssl_insecure_skip_verify` | `IAM_REDIS_SSL_INSECURE` | `false` | 是否跳过 SSL 证书验证 |
| `pool_size` | `IAM_REDIS_POOL_SIZE` | `10` | 连接池大小 |
| `max_active_conns` | `IAM_REDIS_MAX_ACTIVE_CONNS` | `100` | 最大活跃连接数 |
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

### 分析记录配置（`analytics`）

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|---------|--------|------|
| `enable` | 无 | `true` | 是否启用分析记录功能 |
| `enable_detail_recording` | 无 | `true` | 是否启用详细记录 |
| `analytics_key_name` | 无 | `iam-system-analytics` | 分析记录在 Redis 中的 key 名称 |
| `pool_size` | 无 | `10` | 分析记录工作池大小 |
| `record_buffer_size` | 无 | `1024` | 记录缓冲区大小 |
| `flush_interval` | 无 | `10s` | 刷新间隔（1~1000s） |
| `storage_expiration` | 无 | `24h` | 存储过期时间 |

### 本地缓存配置（`cache`）

用于 ristretto 本地缓存配置，缓存 Secret 和 Policy 数据。

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|---------|--------|------|
| `num_counters` | 无 | `0` | 计数器数量（预估唯一 key 数量的 10 倍） |
| `max_cost` | 无 | `0` | 缓存最大成本（内存预算） |
| `buffer_items` | 无 | `0` | 写入缓冲区大小 |

---

## 配置验证规则

加载配置时，`validateConfig` 函数会执行以下校验：

- `server.port`、`grpc.port` — 必须在 1～65535 范围内
- `server.mode` — 必须是 `debug`、`release` 或 `test` 之一

---

## 配置加载流程

```
Load(configPath)
  ├── 设置默认值 (NewConfig)
  ├── 加载配置文件 (LoadConfigFile)
  │     ├── 自动检测路径 (autoConfigFilePath)
  │     │     ├── 优先使用 IAM_CONFIG_FILE 环境变量
  │     │     ├── 根据 IAM_ENV / GO_ENV 选择配置文件名
  │     │     └── 按路径顺序搜索配置文件
  │     └── 解析 YAML 到 Viper
  ├── 绑定环境变量 (IAM_ 前缀)
  │     ├── AutomaticEnv 自动映射
  │     └── LoadEnvVars 显式绑定重要变量
  ├── 解组到 Config 结构体 (Unmarshal)
  └── 配置验证 (validateConfig)
```
