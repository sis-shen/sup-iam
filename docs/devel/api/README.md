# IAM API 文档说明

本目录包含 sup-iam 项目的全部 API 接口文档定义文件（Swagger / OpenAPI），用于描述系统对外提供的接口规范。

API 文档按职责划分，分为 **IAM 控制面 API** 与 **数据面鉴权 API** 两大类。

---

## 1. API 分类说明

### 1.1 IAM 控制面 API

IAM 控制面 API 用于管理员或控制面用户进行身份、密钥、策略等元数据的管理操作，不直接参与在线鉴权流程。

主要能力包括：

- 控制面用户认证与会话管理
- AccessKey / SecretKey 生命周期管理
- 权限策略（Policy）管理
- Secret 与 Policy 的绑定关系管理
- 安全审计数据查询

路径要求:
+ 当前版本控制面路径前缀：`/api/v1/`
+ 主版本号（v1）发布后保持向后兼容，破坏性变更需升级主版本

### 1.2 数据面鉴权 API

数据面鉴权 API 用于对外部业务请求进行签名校验与权限判定，属于高频、低延迟路径。

> 数据面 API 不提供任何资源管理能力。

路径要求:
+ 当前版本数据面路径前缀：`/auth/v1/`
+ 主版本号（v1）发布后保持向后兼容，破坏性变更需升级主版本

### 1.3 调试与诊断 API（Debug API）

Debug API 用于系统运维、问题排查与鉴权链路调试，不属于正常业务能力，
默认不对外开放，不参与任何在线鉴权或资源管理流程。

典型使用场景包括：

- 签名计算结果调试
- 鉴权失败原因解释（Explain）
- Policy 解析与匹配过程诊断
- 鉴权链路输入 / 输出结构验证

路径要求:
+ Debug API 路径前缀：`/debug/v1/`
+ Debug API 必须支持整体开关控制
+ 非调试模式下，Debug API 必须完全不可访问

## 2. IAM 控制面 API 列表
以下路径省略前缀

### 2.1 auth认证系列
+ 控制面用户认证、登录态管理相关接口

| 方法   | 路径                   | 描述             |
|------|----------------------|----------------|
| POST | /auth/login          | 登陆iam用户        |
| POST | /auth/register       | 注册iam用户,用于自助注册 |
| POST | /auth/logout         | 退出登录状态         |
| GET  | /auth/me             | 获取当前用户信息       |
| POST | /auth/refresh        | Token刷新        |
| PUT  | /auth/password       | 密码修改           | 

+  IAM 控制面用户管理接口

| 方法     | 路径          | 描述               |
|--------|-------------|------------------|
| GET    | /users      | 获取User列表         |
| POST   | /users      | 创建一个User,需要管理员权限 |
| GET    | /users/{id} | 获取指定User         |
| PUT    | /users/{id} | 更新指定User         |
| DELETE | /users/{id} | 删除指定User         |

+   Secrets对象管理接口（含AK/Sk)
`
| 方法     | 路径                  | 描述                    |
|--------|---------------------|-----------------------|
| GET    | /secrets            | 获取Secret列表            |
| POST   | /secrets            | 创建一个Secret            |
| GET    | /secrets/{id}       | 获取指定Secret            |
| PUT    | /secrets/{id}       | 更新指定Secret            |
| DELETE | /secrets/{id}       | 删除指定Secret            |
| PUT    | /secrets/{id}/rotate | 轮换指定Secret            |
| GET    | /secrets/{id}/policies | 获取指定Secret的绑定Policy列表 |

`
+  权限策略（Policy）管理接口

| 方法     | 路径                     | 描述                    |
|--------|------------------------|-----------------------|
| GET    | /policies              | 获取Policy列表            |
| POST   | /policies              | 创建一个Policy            |
| GET    | /policies/{id}         | 获取指定Policy            |
| PUT    | /policies/{id}         | 更新指定Policy            |
| DELETE | /policies/{id}         | 删除指定Policy            |
| GET    | /policies/{id}/secrets | 获取指定Policy的绑定Secret列表 |

+ 绑定关系接口列表

| 方法     | 路径                     | 描述           |
|--------|------------------------|--------------|
| GET    | /bindings              | 获取Binding列表  |
| POST   | /bindings              | 创建一个Binding  |
| GET    | /bindings/{id}         | 获取指定Binding   |
| DELETE | /bindings/{id}         | 删除指定Binding  |
| GET    | /bindings?user_id={id} | 获取指定 用户的 Binding 列表 |

+  审计与历史记录查询接口（只读）

| 方法     | 路径                    | 描述            |
|--------|-----------------------|---------------|
| GET    | /audits/policies/{id} | 获取指定Policy审计  |
| GET    | /audits/policies      | 获取Policy审计列表  |
| GET    | /audits/bindings/{id} | 获取指定Binding审计 |
| GET    | /audits/bindings      | 获取Binding审计列表 |

---

## 3. 数据面鉴权 API 列表

以下 Swagger 文件描述数据面请求鉴权相关接口：

+  请求签名校验与权限判定接口

| 方法   | 路径               | 描述            |
|------|------------------|---------------|
| POST | /verify          | 请求鉴权          |


## 4. Debug API 列表
*待定*

## 5. 通用约定

### 5.1 风格约定
- HTTP通信所使用的API均为RESTful API风格
- 使用swagger文档具体描述API

### 5.2 鉴权约定

- 控制面 API 使用登录态 Token（如 JWT）进行访问控制
- 数据面 API 使用 AccessKey / SecretKey 进行签名鉴权
- 控制面 Token 不可用于数据面请求

### 5.3 安全约束

- SecretKey 仅在创建或重置时返回一次
- 日志、缓存、审计数据中禁止记录 SecretKey
- 审计 API 只提供查询能力，不允许写入或修改

### 5.4 Debug API 约束

- Debug API 默认关闭，仅在显式开启调试模式时可访问
- Debug API 仅允许管理员或系统级凭证访问
- Debug API 返回数据不得包含 SecretKey 明文
- Debug API 的所有访问行为必须被审计

---

## 6. 文档维护说明

- 本目录仅存放接口规范定义文件，不包含业务实现
- 接口字段、示例、错误码等详细信息以 Swagger 文件为准
- API 变更需同步更新对应的 Swagger 文件
