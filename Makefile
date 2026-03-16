.PHONY: dev down build tidy lint test migrate

# 启动本地开发环境（中间件）
dev:
	docker compose -f deployments/docker/docker-compose.yml up -d

# 停止本地开发环境
down:
	docker compose -f deployments/docker/docker-compose.yml down

# 编译所有服务
build:
	go build ./cmd/...

# 整理依赖
tidy:
	go mod tidy

# 代码检查
lint:
	golangci-lint run ./...

# 运行测试
test:
	go test -v -race ./...

# 运行用户服务
run-user:
	go run cmd/user/main.go

# 运行商品服务
run-product:
	go run cmd/product/main.go

# 运行订单服务
run-order:
	go run cmd/order/main.go

# 运行网关
run-gateway:
	go run cmd/gateway/main.go
