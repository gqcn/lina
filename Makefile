BACKEND_DIR   := apps/lina-core
FRONTEND_DIR  := apps/lina-vben
PID_DIR       := /tmp/lina-pids
BACKEND_PID   := $(PID_DIR)/backend.pid
FRONTEND_PID  := $(PID_DIR)/frontend.pid
BACKEND_PORT  := 8080
FRONTEND_PORT := 5666

## 依赖Claude Code，自动生成 commit message 并提交到远程仓库
.PHONY: up
up:
	@if git diff --quiet HEAD && git diff --cached --quiet && [ -z "$$(git ls-files --others --exclude-standard)" ]; then \
		echo "No changes to commit"; \
		exit 0; \
	fi
	@git add -A
	@echo "Analyzing changes and generating commit message via AI..."
	@set -e; \
	MSG=$$(git diff --cached --stat && echo "---" && git diff --cached | head -2000 | \
		claude -p "Analyze the git diff above and generate a concise commit message (single line, max 72 chars, lowercase, no quotes). Output only the commit message itself, nothing else." \
		--model haiku) || { echo "Error: Claude command failed"; exit 1; }; \
	COMMIT_MSG=$$(echo "$$MSG" | tail -1); \
	if [ -z "$$COMMIT_MSG" ]; then \
		echo "Error: Failed to generate commit message"; \
		exit 1; \
	fi; \
	echo "Commit: $$COMMIT_MSG"; \
	git commit -m "$$COMMIT_MSG" && \
	git push origin $$(git branch --show-current)
	
## dev: 启动前后端开发服务器
.PHONY: dev
dev: stop
	@mkdir -p $(PID_DIR)
	@# ── 编译后端 ────────────────────────────────────────────────
	@echo "正在重启服务..."
	@cd $(BACKEND_DIR) && go build -o temp/bin/lina . || { echo "后端编译失败"; exit 1; }
	@# ── 启动后端 ────────────────────────────────────────────────
	@cd $(BACKEND_DIR) && ./temp/bin/lina >> /tmp/lina-core.log 2>&1 & echo $$! > $(BACKEND_PID)
	@sleep 1
	@# ── 启动前端 ────────────────────────────────────────────────
	@cd $(FRONTEND_DIR) && npx turbo run dev --filter=@lina/web-antd >> /tmp/lina-vben.log 2>&1 & echo $$! > $(FRONTEND_PID)
	@sleep 2
	@echo ""
	@echo "╔══════════════════════════════════════════════╗"
	@echo "║           Lina Admin - Dev                   ║"
	@echo "╠══════════════════════════════════════════════╣"
	@echo "║  前端地址:  http://localhost:$(FRONTEND_PORT)            ║"
	@echo "║  后端地址:  http://localhost:$(BACKEND_PORT)            ║"
	@echo "║  后端日志:  /tmp/lina-core.log            ║"
	@echo "║  前端日志:  /tmp/lina-vben.log           ║"
	@echo "╚══════════════════════════════════════════════╝"
	@echo ""

## stop: 停止前后端开发服务器
.PHONY: stop
stop:
	@echo "正在停止服务..."
	@if lsof -ti :$(BACKEND_PORT) >/dev/null 2>&1; then \
		kill $$(lsof -ti :$(BACKEND_PORT)) 2>/dev/null; rm -f $(BACKEND_PID); echo "✓ 后端已停止"; \
	else \
		rm -f $(BACKEND_PID); echo "  后端未在运行"; \
	fi
	@if lsof -ti :$(FRONTEND_PORT) >/dev/null 2>&1; then \
		kill $$(lsof -ti :$(FRONTEND_PORT)) 2>/dev/null; rm -f $(FRONTEND_PID); echo "✓ 前端已停止"; \
	else \
		rm -f $(FRONTEND_PID); echo "  前端未在运行"; \
	fi

## status: 查看前后端运行状态及日志路径
.PHONY: status
status:
	@echo ""
	@echo "╔══════════════════════════════════════════════╗"
	@echo "║           Lina Admin - Status                ║"
	@echo "╠══════════════════════════════════════════════╣"
	@if lsof -ti :$(BACKEND_PORT) >/dev/null 2>&1; then \
		echo "║  后端: ✓ 运行中  http://localhost:$(BACKEND_PORT)       ║"; \
	else \
		echo "║  后端: ✗ 未运行  (端口 $(BACKEND_PORT))                 ║"; \
	fi
	@if lsof -ti :$(FRONTEND_PORT) >/dev/null 2>&1; then \
		echo "║  前端: ✓ 运行中  http://localhost:$(FRONTEND_PORT)       ║"; \
	else \
		echo "║  前端: ✗ 未运行  (端口 $(FRONTEND_PORT))                 ║"; \
	fi
	@echo "╠══════════════════════════════════════════════╣"
	@echo "║  后端日志:  /tmp/lina-core.log            ║"
	@echo "║  前端日志:  /tmp/lina-vben.log           ║"
	@echo "╚══════════════════════════════════════════════╝"
	@echo ""

## test: 运行完整 E2E 测试套件
.PHONY: test
test:
	@echo "🧪 运行 E2E 测试套件..."
	cd hack/tests && npx playwright test

## help: 显示帮助信息
.PHONY: help
help:
	@grep -E '^##' Makefile | sed 's/## //'

## init: 初始化数据库（仅执行 DDL 建表和 Seed 数据）
.PHONY: init
init:
	@cd $(BACKEND_DIR) && make init

## mock: 加载 Mock 演示数据（需先执行 init）
.PHONY: mock
mock:
	@cd $(BACKEND_DIR) && make mock
