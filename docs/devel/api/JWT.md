# JWT / AccessToken 设计说明（sup-iam）

本文档描述了 sup-iam 项目中 **AccessToken（JWT）和 RefreshToken** 的设计，包括 payload、生命周期及使用规范。

---

## 1. 概述

在 sup-iam 中，JWT 用于 **控制面 API 的会话管理**：

- **AccessToken**：短期有效，用于 API 调用
- **RefreshToken**：长期有效，用于刷新 AccessToken，避免用户频繁登录

注意事项：

- AccessToken 与 RefreshToken 必须通过 HTTPS 传输
- 不允许在日志中记录敏感信息
- `/auth/me` 提供 token 的友好解码接口，前端无需直接 decode JWT

---

## 2. Token 类型

| Token 类型       | 作用                        | 过期时间          | 建议存储位置                  |
|----------------|---------------------------|-----------------|----------------------------|
| AccessToken    | API 认证                  | 15 分钟（示例）  | 内存 / HttpOnly Cookie      |
| RefreshToken   | 刷新 AccessToken          | 7 天（示例）     | 安全存储，服务端可选       |

---

## 3. AccessToken Payload（Claims）

AccessToken 是 JWT，包含以下 payload：

| 字段      | 类型    | 描述                                | 示例           |
|----------|--------|-----------------------------------|---------------|
| sub      | string | 用户唯一标识（principal）,在 sup-iam 中使用用户唯一 ID 作为稳定标识  | "user_123"    |
| user_id  | string | 系统中用户 ID                          | "abcde-12345" |
| username | string | 用户名                               | "alice"       |
| role     | string | 粗粒度角色（如 admin/user/readonly）      | "admin"       |
| exp      | int    | 过期时间（Unix 时间戳）                    | 1710000000    |
| iat      | int    | 签发时间（Unix 时间戳）                    | 1709996400    |

> 可选字段：tenant_id、service account 等，支持多租户或系统账户扩展

---

## 4. RefreshToken Payload

RefreshToken 可采用 JWT 或 opaque token，payload 精简：

| 字段      | 类型    | 描述           | 示例       |
|----------|--------|----------------|-----------|
| sub      | string | 用户标识       | "user_123"|
| exp      | int    | 过期时间       | 1710600000|
| iat      | int    | 签发时间       | 1709996400|

> RefreshToken 仅用于 `/auth/refresh` 接口，不可直接用于 API 调用

---

## 5. 生命周期

1. 用户通过 `/auth/login` 登录
    - 返回 AccessToken 和 RefreshToken
    - `ExpiresIn` 字段表示 AccessToken 有效期（秒）
2. AccessToken 到期后（如 15 分钟），必须使用 RefreshToken 刷新
3. RefreshToken 到期后（如 7 天），用户需重新登录
4. 用户登出时，必须同时废弃 AccessToken 和 RefreshToken

---

## 6. 使用指南

- 始终通过 HTTPS 传输 token
- 不在 token 中放置密码等敏感信息
- AccessToken 用于 API 调用
- Role 字段用于粗粒度 UI 或网关控制
- 细粒度权限由 Policy 管理
- 前端尽量调用 `/auth/me` 获取用户信息，而不是直接 decode JWT

---

## 7. 示例

**AccessToken payload 示例：**

```json
{
  "sub": "user_123",
  "user_id": "user_123",
  "username": "alice",
  "role": "admin",
  "exp": 1710000000,
  "iat": 1709996400
}
