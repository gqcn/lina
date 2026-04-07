# Lina Development Server Targets
# ================================

## dev: 启动前后端开发服务器
.PHONY: dev
dev: stop
	@mkdir -p $(PID_DIR)
	@> /tmp/lina-core.log
	@> /tmp/lina-vben.log
	@echo "正在重启服务..."
	@cd $(BACKEND_DIR) && go build -o temp/bin/lina . || { echo "后端编译失败"; exit 1; }
	@nohup sh -c 'cd "$(BACKEND_DIR)" && exec ./temp/bin/lina' >> /tmp/lina-core.log 2>&1 < /dev/null & echo $$! > $(BACKEND_PID)
	@sleep 1
	@nohup sh -c 'cd "$(FRONTEND_DIR)" && exec npx turbo run dev --filter=@lina/web-antd' >> /tmp/lina-vben.log 2>&1 < /dev/null & echo $$! > $(FRONTEND_PID)
	@sleep 4
	@make status

## stop: 停止前后端开发服务器
.PHONY: stop
stop:
	@echo "正在停止服务..."
	@_kill_tree() { \
		for child in $$(pgrep -P $$1 2>/dev/null); do \
			_kill_tree $$child; \
		done; \
		kill $$1 2>/dev/null; \
	}; \
	_stop_service() { \
		local name="$$1" pid_file="$$2" port="$$3"; \
		local stopped=false; \
		if [ -f "$$pid_file" ]; then \
			local pid=$$(cat "$$pid_file"); \
			if kill -0 "$$pid" 2>/dev/null; then \
				_kill_tree "$$pid"; \
				stopped=true; \
			fi; \
			rm -f "$$pid_file"; \
		fi; \
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
		if [ "$$stopped" = true ]; then \
			echo "✓ $$name 已停止"; \
		else \
			echo "  $$name 未在运行"; \
		fi; \
	}; \
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
