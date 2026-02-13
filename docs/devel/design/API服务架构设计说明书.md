# API服务架构设计说明书

**项目名称：** sup-iam 身份识别与访问管理系统

**编写人：** 沈冬法

**日期：** 2026年2月2日

**版本号：** V1.0

---

# 1.引言

## 1.1编写目的
为核心服务提供者`iam-api-server`提供详细地架构设计，规范如何实现简洁架构

# 2.设计原则和总目标

## 2.1设计原则

1. 独立于框架：该架构不会依赖于某些功能强大的软件库存在。这可以让你使用这样的框架作为工具，而不是让你的系统陷入到框架的约束中。
2. 可测试性：业务规则可以在没有UI、数据库、Web服务或其他外部元素的情况下进行测试，在实际的开发中，我们通过Mock来解耦这些依赖。
3. 独立于UI ：在无需改变系统其他部分的情况下，UI可以轻松地改变。例如，在没有改变业务规则的情况下，Web UI可以替换为控制台UI。
4. 独立于数据库：你可以用Mongo、Oracle、Etcd或者其他数据库来替换MariaDB，你的业务规则不要绑定到数据库。
5. 独立于外部媒介：实际上，你的业务规则可以简单到根本不去了解外部世界。

## 2.2分层设计
为满足上述设计原则，我们需要讲服务实现分层，并且确保每一层都是独立的可测试的，具体分层如下

+ `Model层`
+ `Controler层`
+ `Service层`
+ `Repository层`

四层之间的依赖关系如下,包导入关系要严格按照依赖关系

![](https://picbed0521.oss-cn-shanghai.aliyuncs.com/blogpic/202602031601063.webp)
# 3. Model层设计
该层定义整个服务中的核心数据结构抽象,,应当实现如下数据结构

+ User
+ Secret
+ Policy
+ Binding
+ PolicyAudit
+ BindingAudit

# 4.Controller层

## 4.1 路由
该层负责控制网络请求的路由，将请求交给对应的服务层函数处理。

## 4.2 鉴权与拦截
将对网络服务的Token进行鉴权与非法访问拦截

### 鉴权通过的合法请求
合法的网络请求将被引导到服务层对应的业务处理回调上

### 鉴权通过的非法请求
如果鉴权通过，但路径非法或者其它验证失败，将会将请求引导到服务层的错误处理回调上

### 鉴权不通过
不通过的网络请求将会被引导到服务层的鉴权失败回调上

# 5.Service层
该层负责接收网络参数，调用Repository层接口进行实际的业务处理。

与API文档不同的是，这里会涉及一些内部服务和DEBUG服务

## 5.1 主要业务

+ auth
  + 用户登录
    + 用户名+密码登录
    + 电话+密码登录
    + 邮箱+密码登录
  + 用户注册
  + 用户登出
  + 用户自查询
  + Token刷新
  + Token创建
  + Token管理
  + Token查询
  + Token反查询用户
  + 密码修改
+ user
  + 管理员创建user
  + 管理员查询user
  + 管理员查询user列表
  + 管理员更新user
  + 管理员删除user
+ secret
  + 根据ID获取secret
  + 更新secret
  + 创建secret
  + 删除secret
  + 根据用户获取secret列表
  + 轮换secret
  + 获取secret绑定的policy列表
+ policy
  + 根据ID获取policy
  + 创建policy
  + 更新policy
  + 删除policy
  + 获取policy列表
  + 获取policy绑定的secret列表
+ binding
  + 根据ID查询binding 
  + 删除binding
  + 根据用户获取binding列表
  + 创建binding
+ audit 限管理员
  + 根据ID获取policy audit
  + 管理员获取policy audit列表
  + 根据ID获取binding audit
  + 管理员获取binding audit列表

## 5.2错误处理业务

+ 未知路径处理业务
+ 请求错误处理业务
+ 鉴权失败处理业务
+ Token过期处理业务

## 5.3 DEBUG等其它业务
+ 请求响应业务


# 6.Repository层
该层负责数据持久化，向Service层提供数据持久化和管理服务，并封装隐藏具体的实现细节

服务接口设计按照存储的数据结构划分，设计原则是**最小实现**

## 6.1 User
业务主键：user id

+ 新建user存储
+ 根据user id获取user
+ 根据username获取user
+ 根据email获取user
+ 根据phone获取user
+ 使用user对象更新user字段
+ 根据user id删除user
+ 分页获取user列表

## 6.2 Secret

+ 新建secret存储
+ 根据secret id获取secret
+ 使用secret对象更新secret字段
+ 根据secret id删除secret
+ 分页根据用户id获取secret列表
+ 分页根据secret id获取绑定的policy列表

## 6.3 policy

+ 根据policy id获取policy
+ 创建policy存储
+ 更新policy
+ 根据id删除policy
+ 根据user id获取policy列表
+ 分页获取指定policy id绑定的secret列表

## 6.4 binding
+ 根据ID查询binding
+ 删除binding
+ 分页根据user id获取binding列表
+ 创建binding

## 6.5 audit
+ 根据ID获取policy audit
+ 分页获取policy audit列表
+ 根据ID获取binding audit
+ 分页获取binding audit列表

## 6.6 列表查询接口抽象
列表查询我们有如下设计原则

+ 独立于数据库
+ Repository 层长期稳定
+ Service 层不被存储反向牵引

因此我们有如下抽象:


```go
type Order string

const (
OrderAsc  Order = "asc"
OrderDesc Order = "desc"
)
type PageQuery struct {
	Limit   int
	Cursor  string
	OrderBy string
	Order   Order
}
```

```go
type PageResult[T any] struct {
	Items  []T
	Total  int64
	Next   string // 下一页 cursor
}
```

### 特化示例：MySQL

在 MySQL 存储实现中，列表查询通常依赖有序索引，并结合 `LIMIT` 进行结果集裁剪。  
Repository 层并不直接暴露 `OFFSET` 语义，而是通过 **游标（Cursor）** 将分页状态向上抽象。

#### 实现策略说明

- **排序约束**  
  MySQL 实现要求查询必须基于确定性排序字段（如 `id`、`created_at`），并且该字段应建立索引。

- **Cursor 设计**  
  Cursor 通常编码为：
  - 上一页最后一条记录的排序字段值
  - 或主键 `id`（推荐）

  Cursor 的编码与解析逻辑仅存在于 Repository 实现中。

- **PageQuery 到 SQL 的映射关系**

  | PageQuery 字段 | MySQL 查询语义示例 |
  |---------------|--------------------|
  | `Limit`       | `LIMIT N`          |
  | `Cursor`      | `WHERE id > ?`     |
  | `OrderBy`     | `ORDER BY id`      |
  | `Order`       | `ASC / DESC`       |

#### 示例查询逻辑（示意）

```sql
SELECT *
FROM users
WHERE id > :cursor
ORDER BY id ASC
LIMIT :limit;
````

#### Next Cursor 生成规则

* 若返回结果集非空：

  * 取最后一条记录的 `id` 作为下一页 Cursor
* 若结果集为空：

  * 不返回 Next Cursor，表示分页结束

该实现方式避免了大 OFFSET 带来的性能问题，同时保持 Repository 接口稳定。

---

### 特化示例：MongoDB

在 MongoDB 存储实现中，列表查询天然适合使用基于 `_id` 或索引字段的游标分页模型。

#### 实现策略说明

* **排序约束**
  默认使用 `_id` 作为排序与分页基准字段，确保顺序稳定且具备索引支持。

* **Cursor 设计**
  Cursor 通常为：

  * `_id` 的字符串表示形式（如 ObjectID 的 hex 值）

* **PageQuery 到 MongoDB 查询的映射关系**

  | PageQuery 字段 | MongoDB 查询语义示例             |
    | ------------ | -------------------------- |
  | `Limit`      | `limit(n)`                 |
  | `Cursor`     | `{ _id: { $gt: cursor } }` |
  | `OrderBy`    | `{ _id: 1 }`               |
  | `Order`      | `1 / -1`                   |

#### 示例查询逻辑（示意）

```js
db.users.find(
  { _id: { $gt: ObjectId(cursor) } }
).sort({ _id: 1 }).limit(limit)
```

#### Next Cursor 生成规则

* 若返回结果集非空：

  * 取最后一条文档的 `_id` 作为 Next Cursor
* 若结果集为空：

  * 不返回 Next Cursor，表示分页结束

该实现方式充分利用 MongoDB 的游标与索引机制，避免了基于页号的低效分页方式。


