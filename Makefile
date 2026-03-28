.PHONY: help env-init dev test-env prod run start port build-dev build-test build-prod clean

# 命令行传入的监听变量会传给 go run：make prod HTTP_ADDR=0.0.0.0:8090
export HTTP_ADDR LISTEN_ADDR HOST PORT

# 与 internal/config/env.go 一致：
#   .env → .env.development | .env.test | .env.production（由 APP_ENV 决定）→ .env.local
# 后加载文件中的同名键会覆盖先加载文件（进程启动前已存在的环境变量不会被覆盖）

help:
	@echo "环境变量文件：.env（通用） + .env.development | .env.test | .env.production + 可选 .env.local"
	@echo ""
	@echo "Targets:"
	@echo "  make env-init     从 *.example 生成缺失的 .env / .env.*（不覆盖已有）"
	@echo "  make dev          APP_ENV=development，加载 .env.development"
	@echo "  make test-env     APP_ENV=test，加载 .env.test"
	@echo "  make prod         APP_ENV=production，加载 .env.production"
	@echo "  make run / start  未设置 APP_ENV 时由 Makefile 设为 development（与仅写 .env 时的默认加载一致）"
	@echo "  make port PORT=3000   指定端口"
	@echo "  make build-dev | build-test | build-prod   编译到 bin/"

# 与 make run 相同：直接 go run 时请 export APP_ENV=… 或在 .env 中写明 APP_ENV
start: run

env-init:
	@test -f .env || (cp .env.example .env && echo "created .env from .env.example")
	@test -f .env.development || (cp .env.development.example .env.development && echo "created .env.development")
	@test -f .env.test || (cp .env.test.example .env.test && echo "created .env.test")
	@test -f .env.production || (cp .env.production.example .env.production && echo "created .env.production")
	@echo "env-init done (existing files were not overwritten)"

dev:
	APP_ENV=development go run ./cmd/server

test-env:
	APP_ENV=test go run ./cmd/server

prod:
	APP_ENV=production go run ./cmd/server

run:
	APP_ENV=$${APP_ENV:-development} go run ./cmd/server

port:
	PORT=$${PORT} APP_ENV=$${APP_ENV:-development} go run ./cmd/server

build-dev:
	@mkdir -p bin
	go build -ldflags "-X main.buildEnv=development" -o bin/server-dev ./cmd/server
	@echo "built bin/server-dev"

build-test:
	@mkdir -p bin
	go build -ldflags "-X main.buildEnv=test" -o bin/server-test ./cmd/server
	@echo "built bin/server-test"

build-prod:
	@mkdir -p bin
	go build -ldflags "-X main.buildEnv=production" -o bin/server-prod ./cmd/server
	@echo "built bin/server-prod"

clean:
	rm -rf bin/
