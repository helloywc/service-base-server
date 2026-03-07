.PHONY: dev prod run build-dev build-prod port

# 默认开发环境
dev:
	APP_ENV=dev go run ./cmd/server

# 生产环境模式
prod:
	APP_ENV=prod go run ./cmd/server

# 支持自定义端口的开发模式 (例如: make port PORT=3000)
port:
	PORT=$(PORT) APP_ENV=dev go run ./cmd/server

# 打包 dev / prod（Make 目标名用 - 不用 :，即 make build-dev / make build-prod）
# 打包 dev：生成 bin/server-dev，直接运行即 8080
build-dev:
	@mkdir -p bin
	go build -ldflags "-X main.buildEnv=dev" -o bin/server-dev ./cmd/server
	@echo "built bin/server-dev (port 8080)"

# 打包 prod：生成 bin/server-prod，直接运行即 8090
build-prod:
	@mkdir -p bin
	go build -ldflags "-X main.buildEnv=prod" -o bin/server-prod ./cmd/server
	@echo "built bin/server-prod (port 8090)"