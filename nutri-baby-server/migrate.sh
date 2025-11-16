#!/bin/bash
# 数据库迁移脚本
# 使用方法: ./migrate.sh up 或 ./migrate.sh down

# 读取配置
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_NAME=${DB_NAME:-nutri_baby}
DB_PASSWORD=${DB_PASSWORD:-}

# 获取迁移方向
DIRECTION=${1:-up}

# 迁移文件目录
MIGRATIONS_DIR="$(dirname "$0")/migrations"

if [ "$DIRECTION" = "up" ]; then
    echo "执行数据库迁移升级..."
    # 按文件名顺序执行所有 SQL 文件
    for sql_file in $(ls -1 "$MIGRATIONS_DIR"/*.sql | sort); do
        echo "执行迁移: $(basename $sql_file)"
        if [ -z "$DB_PASSWORD" ]; then
            psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$sql_file"
        else
            PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$sql_file"
        fi

        if [ $? -eq 0 ]; then
            echo "✓ $(basename $sql_file) 迁移成功"
        else
            echo "✗ $(basename $sql_file) 迁移失败"
            exit 1
        fi
    done
    echo "所有迁移完成！"
elif [ "$DIRECTION" = "down" ]; then
    echo "数据库降级功能需要在迁移文件中定义回滚脚本"
    echo "目前暂不支持自动降级"
else
    echo "使用方法: ./migrate.sh [up|down]"
    exit 1
fi
