.PHONY: wire run test fmt lint migrate-up migrate-down clean

# 生成Wire依赖注入代码
wire:
	cd wire && wire

# 运行服务
run:
	go run cmd/server/main.go

# 构建
build:
	go build -o bin/server cmd/server/main.go

# 运行测试
test:
	go test -v ./...

# 代码格式化
fmt:
	go fmt ./...
	goimports -w .

# 代码检查
lint:
	golangci-lint run

# 数据库迁移 - 升级
migrate-up:
	go run migrations/migrate.go up

# 数据库迁移 - 降级
migrate-down:
	go run migrations/migrate.go down

# 清理
clean:
	rm -rf bin/
	rm -rf logs/
	find . -name "wire_gen.go" -delete

# 安装工具
install-tools:
	go install github.com/google/wire/cmd/wire@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 帮助
help:
	@echo "可用命令:"
	@echo "  make wire         - 生成Wire依赖注入代码"
	@echo "  make run          - 运行服务"
	@echo "  make build        - 构建可执行文件"
	@echo "  make test         - 运行测试"
	@echo "  make fmt          - 格式化代码"
	@echo "  make lint         - 代码检查"
	@echo "  make migrate-up   - 数据库迁移升级"
	@echo "  make migrate-down - 数据库迁移降级"
	@echo "  make clean        - 清理生成文件"
	@echo "  make install-tools- 安装开发工具"
