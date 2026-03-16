.PHONY: dev prod run build-dev build-prod port

# 开发环境（APP_ENV=development, 端口 8080）
dev:
	APP_ENV=development go run ./cmd/server

# 生产环境（APP_ENV=prod, 端口 8090）
prod:
	APP_ENV=prod go run ./cmd/server

# 支持自定义端口的开发模式 (例如: make port PORT=3000)
port:
	PORT=$(PORT) APP_ENV=development go run ./cmd/server

# 打包 development / prod
build-dev:
	@mkdir -p bin
	go build -ldflags "-X main.buildEnv=development" -o bin/server-dev ./cmd/server
	@echo "built bin/server-dev (port 8080)"

# 打包 prod：生成 bin/server-prod，直接运行即 8090
build-prod:
	@mkdir -p bin
	go build -ldflags "-X main.buildEnv=prod" -o bin/server-prod ./cmd/server
	@echo "built bin/server-prod (port 8090)"