# Database Deployment SQL

本目录包含 sup-iam 项目的 MySQL 数据库初始化与表结构定义 SQL，
用于系统首次部署或全量重建数据库。

## 说明

- 仅包含 schema 定义与必要的触发器
- 不包含业务初始化数据（如用户、密钥等）

## 执行顺序

```bash
mysql -u <user> -p < init.sql
...
