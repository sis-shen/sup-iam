# BenchMark 测试设计说明书

**项目名称：** sup-iam 身份识别与访问管理系统

**编写人：** 沈冬法

**日期：** 2026年6月29日

**版本号：** V1.0

---

# 1. 引言

## 1.1 编写目的

`Authorize` 函数作为鉴权服务的核心函数，是业务中 CPU 耗时的主要代码段。本文档定义其 Benchmark 测试方案，通过标准化的性能测试量化各场景下的执行耗时，识别瓶颈并为后续优化提供依据。

## 1.2 适用范围

本文档涵盖 iam-auth-server 内部 `service.AuthCase.Authorize` 函数的性能基准测试，不包含网络 IO（gRPC、HTTP）和外部依赖（Redis、MySQL）的性能测试。

## 1.3 参考文档

- 《全量缓存设计说明书》
- `internal/iam-auth-server/v1/service/auth.go`
- `internal/iam-auth-server/v1/service/auth_test.go`

## 1.4 被测函数签名

```go
func (ac *AuthCase) Authorize(secretID string, username string, path string, method string) (bool, []string, error)
```

---

# 2. 测试环境

## 2.1 硬件环境

| 项目 | 配置 |
|------|------|
| CPU | AMD Ryzen 9 7940H (16 逻辑核) |
| 内存 | 32 GB |
| 操作系统 | Windows 11 / Linux (WSL2 / Docker) |

## 2.2 软件环境

| 项目 | 版本 |
|------|------|
| Go | 1.22+ |
| Casbin | v2.135.0 |
| Ristretto | v0.2.0 |
| 测试框架 | Go testing + b.RunParallel |

## 2.3 依赖 Mock

| 依赖 | Mock 方式 | 说明 |
|------|----------|------|
| 缓存 (Secret/Policy) | `mockRpcClient` 提供数据，`cache.InitSingleton` 创建真实 Ristretto 实例 | 全量缓存真实初始化，确保缓存访问路径与实际一致 |
| 数据采集 (Analytics) | `mockAnalyticsStore`，写入 Redis 的调用为空操作 | 避免 IO 阻塞干扰计时 |
| 密钥 (Keys) | `keys.NewKeys` 创建真实 HMAC-SHA256 实现 | Authorize 内部不使用密钥，仅用于 VerifySecretKey 测试 |

---

# 3. 测试用例设计

## 3.1 用例概览

| 编号 | 名称 | 场景 | 并发模型 | 测量目标 |
|------|------|------|---------|---------|
| BM-1 | 1Policy_Hit_Serial | 单条策略匹配命中 | 串行 (b.N) | 基准：一次完整 casbin 链路开销 |
| BM-2 | 1Policy_Hit_Parallel | 单条策略匹配命中 | 并发 (RunParallel) | 锁/池竞争下的退化程度 |
| BM-3 | 5Policies_MidHit_Parallel | 5条策略，第3条命中 | 并发 | 多条策略中间命中的开销 |
| BM-4 | 5Policies_AllMiss_Parallel | 5条策略全部不匹配（遍历到底） | 并发 | 最坏情况：完整遍历+全量 casbin 周期 |
| BM-5 | 3Rules_Parallel | 1条策略含3条rule | 并发 | 单策略多条rule的匹配开销 |
| BM-6 | CacheMiss | secret_id不存在（不走 casbin） | 串行 | `GetPolicyListBySecretID` 纯缓存查找开销 |

## 3.2 策略 DSL 数据

| 用例 | DSL 内容 |
|------|---------|
| BM-1/2 | `[["alice", "/api/resource", "GET"]]` |
| BM-3 | 5条策略，前2条不匹配（sub/obj/act均不同），第3条匹配，后2条不匹配 |
| BM-4 | 5条策略全部不匹配 |
| BM-5 | 1条策略含3条rule，第3条匹配 |

## 3.3 预热策略

每个 Benchmark 函数在 `b.ResetTimer()` 之前先执行 200 次 `Authorize` 调用，将 `sync.Pool` 中的 Casbin Enforcer 实例预热至稳定状态，消除 `NewEnforcer` 冷启动开销对测试结果的干扰。

---

# 4. 测试结果

## 4.1 串行结果 (GOMAXPROCS=1)

| 用例 | 耗时 (ns/op) | 换算 QPS |
|------|-------------|---------|
| BM-1: 1Policy_Hit_Serial | **35,700** | 28,011 |
| BM-5: 3Rules_Parallel (c=1) | 44,814 | 22,314 |
| BM-3: 5Policies_MidHit (c=1) | 92,815 | 10,774 |
| BM-4: 5Policies_AllMiss (c=1) | 149,068 | 6,708 |
| BM-6: CacheMiss | 294.8 | 3,392,000 |

## 4.2 并发结果 (GOMAXPROCS=8, b.RunParallel)

| 用例 | 耗时 (ns/op) | 加速比 | 换算 QPS (8核) |
|------|-------------|-------|---------------|
| BM-2: 1Policy_Hit_Parallel-8 | **6,383** | 5.6x | 156,665 |
| BM-5: 3Rules_Parallel-8 | 7,397 | 6.1x | 135,189 |
| BM-3: 5Policies_MidHit-8 | 15,119 | 6.1x | 66,142 |
| BM-4: 5Policies_AllMiss-8 | 27,927 | 5.3x | 35,808 |
| BM-6: CacheMiss-8 | 189.4 | 1.6x | 5,279,831 |

## 4.3 关键发现

1. **单次 Authorize 仅 35.7μs（串行）**。加上 Gin 解析 (~80μs) + HMAC (~5μs)，完整 handler 约 120-140μs，与 Gin 日志中的 118-199μs 吻合。

2. **并发放大效应显著（5.3-6.1x）**。`sync.Pool` 无锁竞争，多 goroutine 下 Casbin 操作完美缩放。说明 `Authorize` 函数本身不是生产环境的瓶颈。

3. **CacheMiss 仅 294.8ns（串行）**。纯缓存查找路径（不经过 Casbin）开销可以忽略。

4. **最坏情况（5条全Miss）149μs（串行）/ 27.9μs（并发）**。策略数量与耗时近似线性关系。

---

# 5. 性能瓶颈分析与推论

## 5.1 本地单核理论极限

```
单核 QPS 上限 = 1 / 35.7μs ≈ 28,011 QPS
```

## 5.2 生产环境 QPS 瓶颈推论

生产环境实测 6255 QPS 远低于理论值。Benchmark 结果排除了 `Authorize` 函数内部的锁竞争和 Casbin 开销作为主瓶颈的可能性。瓶颈的更可能来源：

| 可能瓶颈 | 推论依据 | 验证方式 |
|---------|---------|---------|
| GC 清空 sync.Pool 导致 NewEnforcer 串行 | BM 预热后性能稳定，但生产环境 GC 周期约 5-8s | pprof CPU/mutex profile |
| 容器 CPU Limit 过小 | 本地 16 核可跑 6K QPS 的理论上限远高于 6255 | 容器监控 (kubectl top) |
| 请求排队（goroutine 调度） | BM 并发加速比正常，说明无锁竞争 | pprof goroutine profile |

---

# 6. 附录

## 6.1 测试边界

- 所有测试均使用 `b.RunParallel` 模拟并发，与实际 Gin HTTP 协程模型一致
- 测试数据均缓存在 Ristretto 实例中，跳过网络加载阶段
- Analytics 写入使用空操作 Mock，非阻塞

## 6.2 复现方式

```bash
cd internal/iam-auth-server/v1/service
go test -bench=BenchmarkAuthorize -benchtime=3s -cpu=1,8 .
```
