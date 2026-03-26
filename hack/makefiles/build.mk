# Lina Build Target
# =================

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