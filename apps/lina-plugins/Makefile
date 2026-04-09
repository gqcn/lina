## wasm: 构建全部或指定 runtime wasm 插件；可通过 `p=<plugin-id>` 定向构建
.PHONY: wasm
wasm:
	@set -e; \
	plugin_list=""; \
	if [ -n "$(p)" ]; then \
		plugin_list="$(p)"; \
	else \
		plugin_list=$$(find . -mindepth 2 -maxdepth 2 -name plugin.yaml | sort | while read -r manifest; do \
			if grep -Eq '^type:[[:space:]]*runtime[[:space:]]*$$' "$$manifest"; then \
				basename "$$(dirname "$$manifest")"; \
			fi; \
		done); \
	fi; \
	if [ -z "$$plugin_list" ]; then \
		echo "未发现可构建的 runtime wasm 插件"; \
		exit 0; \
	fi; \
	for plugin in $$plugin_list; do \
		manifest="./$$plugin/plugin.yaml"; \
		if [ ! -f "$$manifest" ]; then \
			echo "插件不存在: $$plugin"; \
			exit 1; \
		fi; \
		if ! grep -Eq '^type:[[:space:]]*runtime[[:space:]]*$$' "$$manifest"; then \
			echo "插件不是 runtime 类型，无法构建 wasm: $$plugin"; \
			exit 1; \
		fi; \
		echo "构建 runtime wasm 插件: $$plugin"; \
		(cd ../../hack/build-wasm && go run . --plugin-dir ../../apps/lina-plugins/$$plugin); \
	done
