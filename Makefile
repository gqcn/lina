# Lina Admin - Root Makefile
# ===========================

BACKEND_DIR   := apps/lina-core
FRONTEND_DIR  := apps/lina-vben
PID_DIR       := /tmp/lina-pids
BACKEND_PID   := $(PID_DIR)/backend.pid
FRONTEND_PID  := $(PID_DIR)/frontend.pid
BACKEND_PORT  := 8080
FRONTEND_PORT := 5666
EMBED_DIR     := $(BACKEND_DIR)/internal/packed/public

# 引用复杂指令子文件
include hack/makefiles/dev.mk
include hack/makefiles/build.mk
include hack/makefiles/up.mk

## test: 运行完整 E2E 测试套件
.PHONY: test
test:
	@echo "🧪 运行 E2E 测试套件..."
	cd hack/tests && npx playwright test

## init: 初始化数据库（仅执行 DDL 建表和 Seed 数据）
.PHONY: init
init:
	@cd $(BACKEND_DIR) && make init

## mock: 加载 Mock 演示数据（需先执行 init）
.PHONY: mock
mock:
	@cd $(BACKEND_DIR) && make mock

## help: 显示帮助信息
.PHONY: help
help:
	@grep -E '^## [a-z]+:' hack/makefiles/*.mk Makefile | sed 's/.*## /  /'