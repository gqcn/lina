BACKEND_DIR   := apps/lina-core
FRONTEND_DIR  := apps/lina-vben
PID_DIR       := /tmp/lina-pids
BACKEND_PID   := $(PID_DIR)/backend.pid
FRONTEND_PID  := $(PID_DIR)/frontend.pid
BACKEND_PORT  := 8080
FRONTEND_PORT := 5666
EMBED_DIR     := $(BACKEND_DIR)/internal/packed/public

## 依赖Claude Code，自动生成 commit message 并提交到远程仓库
## 用法: make up [m=xxx] (默认 m=haiku)
## 示例: make up m=glm-5
## 逻辑：有变更则 AI 生成 commit message 并提交推送；无变更但有未推送的 commit 则直接推送
.PHONY: up
up:
	@set -e; \
	if git diff --quiet HEAD && git diff --cached --quiet && [ -z "$$(git ls-files --others --exclude-standard)" ]; then \
		if git diff --quiet HEAD origin/$$(git branch --show-current) 2>/dev/null; then \
			echo "No changes to commit and nothing to push"; \
		else \
			echo "No local changes, pushing unpushed commits..."; \
			git push origin $$(git branch --show-current); \
		fi; \
		exit 0; \
	fi; \
	git add -A; \
	echo "Analyzing changes and generating commit message via AI (model: $(or $(m),haiku))..."; \
	MSG=$$(git diff --cached --stat && echo "---" && git diff --cached | head -2000 | \
		claude -p "Analyze the git diff above and generate a concise commit message (single line, max 72 chars, lowercase, no quotes). Output only the commit message itself, nothing else." \
		--model $(or $(m),haiku)) || { echo "Error: Claude command failed"; exit 1; }; \
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
	@# ── 清空旧日志 ────────────────────────────────────────────
	@> /tmp/lina-core.log
	@> /tmp/lina-vben.log
	@# ── 编译后端 ────────────────────────────────────────────────
	@echo "正在重启服务..."
	@cd $(BACKEND_DIR) && go build -o temp/bin/lina . || { echo "后端编译失败"; exit 1; }
	@# ── 启动后端 ────────────────────────────────────────────────
	@cd $(BACKEND_DIR) && ./temp/bin/lina >> /tmp/lina-core.log 2>&1 & echo $$! > $(BACKEND_PID)
	@sleep 1
	@# ── 启动前端 ────────────────────────────────────────────────
	@cd $(FRONTEND_DIR) && npx turbo run dev --filter=@lina/web-antd >> /tmp/lina-vben.log 2>&1 & echo $$! > $(FRONTEND_PID)
	@sleep 4
	@make status

## stop: 停止前后端开发服务器
.PHONY: stop
stop:
	@echo "正在停止服务..."
	@# ── 辅助函数：递归杀进程树（先杀子进程再杀父进程）──────────
	@# kill_tree <pid>: 通过 pgrep 递归查找子进程并逐一终止
	@_kill_tree() { \
		for child in $$(pgrep -P $$1 2>/dev/null); do \
			_kill_tree $$child; \
		done; \
		kill $$1 2>/dev/null; \
	}; \
	\
	_stop_service() { \
		local name="$$1" pid_file="$$2" port="$$3"; \
		local stopped=false; \
		\
		if [ -f "$$pid_file" ]; then \
			local pid=$$(cat "$$pid_file"); \
			if kill -0 "$$pid" 2>/dev/null; then \
				_kill_tree "$$pid"; \
				stopped=true; \
			fi; \
			rm -f "$$pid_file"; \
		fi; \
		\
		local pids=$$(lsof -ti :"$$port" 2>/dev/null); \
		if [ -n "$$pids" ]; then \
			echo "$$pids" | xargs kill 2>/dev/null; \
			sleep 0.5; \
			pids=$$(lsof -ti :"$$port" 2>/dev/null); \
			if [ -n "$$pids" ]; then \
				echo "$$pids" | xargs kill -9 2>/dev/null; \
			fi; \
			stopped=true; \
		fi; \
		\
		if [ "$$stopped" = true ]; then \
			echo "✓ $$name 已停止"; \
		else \
			echo "  $$name 未在运行"; \
		fi; \
	}; \
	\
	_stop_service "后端" "$(BACKEND_PID)" "$(BACKEND_PORT)"; \
	_stop_service "前端" "$(FRONTEND_PID)" "$(FRONTEND_PORT)"

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
	@echo "║  后端日志:  /tmp/lina-core.log               ║"
	@echo "║  前端日志:  /tmp/lina-vben.log               ║"
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

## build: 构建单体二进制（前端嵌入）
.PHONY: build
build:
	@echo "构建前端..."
	@cd $(FRONTEND_DIR) && pnpm run build
	@rm -rf $(EMBED_DIR)/*
	@mkdir -p $(EMBED_DIR)
	@cp -r $(FRONTEND_DIR)/apps/web-antd/dist/* $(EMBED_DIR)/
	@echo "✓ 前端构建完成"
	@echo "构建后端（嵌入前端静态文件）..."
	@cd $(BACKEND_DIR) && go build -o lina .
	@echo "✓ 单体二进制构建完成: $(BACKEND_DIR)/lina"
