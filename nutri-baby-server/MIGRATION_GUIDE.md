# 数据库迁移指南

## 安装 PostgreSQL 客户端工具

### macOS
```bash
# 使用 Homebrew 安装
brew install postgresql@14

# 或使用 Postgres.app 自带的命令行工具
# 添加到 PATH: export PATH=/Applications/Postgres.app/Contents/Versions/latest/bin:$PATH
```

### Linux (Ubuntu/Debian)
```bash
sudo apt-get install postgresql-client
```

### Windows
下载并安装 PostgreSQL 官方客户端: https://www.postgresql.org/download/windows/

## 执行迁移

### 方式一: 使用 psql 命令行

```bash
cd nutri-baby-server

# 执行订阅消息迁移脚本
psql -h 101.200.47.93 -U postgres -d postgres -f migrations/003_subscribe_message.sql

# 如果需要密码,会提示输入
```

### 方式二: 使用数据库管理工具

可以使用以下任意工具:
- **DBeaver** (推荐): https://dbeaver.io/
- **pgAdmin**: https://www.pgadmin.org/
- **TablePlus**: https://tableplus.com/
- **DataGrip**: https://www.jetbrains.com/datagrip/

步骤:
1. 连接到数据库服务器 (Host: 101.200.47.93, Database: postgres, User: postgres)
2. 打开 `migrations/003_subscribe_message.sql` 文件
3. 执行 SQL 脚本

### 方式三: 使用 Go 代码自动迁移

在 `cmd/server/main.go` 中添加 GORM 自动迁移:

```go
// 自动迁移订阅消息相关表
db.AutoMigrate(
    &entity.SubscribeRecord{},
    &entity.MessageSendLog{},
    &entity.MessageSendQueue{},
)
```

## 验证迁移

执行以下 SQL 验证表是否创建成功:

```sql
-- 检查表是否存在
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
AND table_name IN ('subscribe_records', 'message_send_logs', 'message_send_queue');

-- 检查表结构
\d subscribe_records
\d message_send_logs
\d message_send_queue
```

## 回滚迁移

如果需要回滚,执行:

```sql
DROP TABLE IF EXISTS message_send_queue CASCADE;
DROP TABLE IF EXISTS message_send_logs CASCADE;
DROP TABLE IF EXISTS subscribe_records CASCADE;
```

## 注意事项

1. **生产环境**: 请在执行迁移前备份数据库
2. **权限**: 确保数据库用户有 CREATE TABLE 权限
3. **时区**: 数据库时区应设置为 UTC 或与应用服务器一致
4. **索引**: 迁移脚本已包含必要的索引,确保性能优化
