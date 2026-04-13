package builder

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/pkg/pluginbridge"
)

func TestBuildRuntimeWasmArtifactFromSourceEmbedsDeclaredAssets(t *testing.T) {
	pluginDir := t.TempDir()

	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dynamic-builder\nname: Dynamic Builder\nversion: v0.1.0\ntype: dynamic\ndescription: standalone builder test\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "frontend", "pages", "standalone.html"),
		"<!doctype html><html><body>it works</body></html>",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "sql", "001-plugin-dynamic-builder.sql"),
		"SELECT 1;",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "sql", "uninstall", "001-plugin-dynamic-builder.sql"),
		"SELECT 2;",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "hooks", "001-login.yaml"),
		"event: auth.login.succeeded\naction: sleep\ntimeoutMs: 50\nsleepMs: 10\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "resources", "001-records.yaml"),
		"key: records\ntype: table-list\ntable: plugin_runtime_records\nfields:\n  - name: id\n    column: id\norderBy:\n  column: id\n  direction: asc\ndataScope:\n  userColumn: owner_user_id\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "api", "dynamic", "v1", "review_summary.go"),
		"package v1\n\nimport \"github.com/gogf/gf/v2/frame/g\"\n\ntype ReviewSummaryReq struct {\n\tg.Meta `path:\"/review-summary\" method:\"get\" tags:\"动态插件示例\" summary:\"查询摘要\" dc:\"返回一个动态插件摘要\" access:\"login\" permission:\"plugin-dynamic-builder:review:view\" operLog:\"other\"`\n}\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "main.go"),
		"package main\n\nfunc main() {}\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin_embed.go"),
		"package main\n\nimport \"embed\"\n\n//go:embed plugin.yaml frontend manifest\nvar EmbeddedFiles embed.FS\n",
	)

	out, err := BuildRuntimeWasmArtifactFromSource(pluginDir)
	if err != nil {
		t.Fatalf("expected dynamic build to succeed, got error: %v", err)
	}
	if out.RuntimePath != "" {
		t.Cleanup(func() {
			_ = os.RemoveAll(filepath.Dir(out.RuntimePath))
		})
	}
	if expected := filepath.Join(pluginDir, "temp", "plugin-dynamic-builder.wasm"); out.ArtifactPath != expected {
		t.Fatalf("expected artifact path %s, got %s", expected, out.ArtifactPath)
	}

	sections, err := parseWasmCustomSections(out.Content)
	if err != nil {
		t.Fatalf("expected wasm custom sections to parse, got error: %v", err)
	}

	manifest := &dynamicArtifactManifest{}
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionManifest], manifest); err != nil {
		t.Fatalf("expected manifest section json to unmarshal, got error: %v", err)
	}
	if manifest.ID != "plugin-dynamic-builder" {
		t.Fatalf("expected embedded manifest id plugin-dynamic-builder, got %s", manifest.ID)
	}

	metadata := &dynamicArtifactMetadata{}
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionDynamic], metadata); err != nil {
		t.Fatalf("expected dynamic section json to unmarshal, got error: %v", err)
	}
	if metadata.FrontendAssetCount != 1 || metadata.SQLAssetCount != 2 {
		t.Fatalf("expected dynamic metadata counts 1/2, got %#v", metadata)
	}

	var frontend []*frontendAsset
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionFrontend], &frontend); err != nil {
		t.Fatalf("expected frontend section json to unmarshal, got error: %v", err)
	}
	if len(frontend) != 1 || frontend[0].Path != "standalone.html" {
		t.Fatalf("unexpected embedded frontend assets: %#v", frontend)
	}

	var hooks []*hookSpec
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionBackendHooks], &hooks); err != nil {
		t.Fatalf("expected hook section json to unmarshal, got error: %v", err)
	}
	if len(hooks) != 1 || hooks[0].Action != hookActionSleep {
		t.Fatalf("unexpected embedded hook specs: %#v", hooks)
	}

	var resources []*resourceSpec
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionBackendRes], &resources); err != nil {
		t.Fatalf("expected resource section json to unmarshal, got error: %v", err)
	}
	if len(resources) != 1 || resources[0].DataScope == nil || resources[0].DataScope.UserColumn != "owner_user_id" {
		t.Fatalf("unexpected embedded resource specs: %#v", resources)
	}

	var routes []*pluginbridge.RouteContract
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionBackendRoutes], &routes); err != nil {
		t.Fatalf("expected route section json to unmarshal, got error: %v", err)
	}
	if len(routes) != 1 || routes[0].Permission != "plugin-dynamic-builder:review:view" {
		t.Fatalf("unexpected embedded route specs: %#v", routes)
	}

	bridgeSpec := &pluginbridge.BridgeSpec{}
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionBackendBridge], bridgeSpec); err != nil {
		t.Fatalf("expected bridge section json to unmarshal, got error: %v", err)
	}
	if !bridgeSpec.RouteExecution || bridgeSpec.RequestCodec != pluginbridge.CodecProtobuf {
		t.Fatalf("unexpected embedded bridge spec: %#v", bridgeSpec)
	}
	if out.RuntimePath == "" {
		t.Fatal("expected executable guest runtime path to be generated")
	}
	if _, err = os.Stat(filepath.Join(pluginDir, "temp", "runtime-plugin.wasm")); !os.IsNotExist(err) {
		t.Fatalf("expected guest runtime wasm to stop being written into plugin temp/, got err=%v", err)
	}
	runtimeStrings, err := readCommandOutput("strings", out.RuntimePath)
	if err != nil {
		t.Fatalf("expected runtime wasm strings inspection to succeed, got error: %v", err)
	}
	if !strings.Contains(runtimeStrings, "_initialize") {
		t.Fatalf("expected runtime guest wasm to expose _initialize, got output: %s", runtimeStrings)
	}
}

func TestBuildRuntimeWasmArtifactFromSourceFailsWhenEmbeddedResourcesOmitManifest(t *testing.T) {
	pluginDir := t.TempDir()

	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dynamic-missing-embed\nname: Dynamic Missing Embed\nversion: v0.1.0\ntype: dynamic\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "frontend", "pages", "standalone.html"),
		"<!doctype html><html><body>it works</body></html>",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin_embed.go"),
		"package main\n\nimport \"embed\"\n\n//go:embed frontend\nvar EmbeddedFiles embed.FS\n",
	)

	_, err := BuildRuntimeWasmArtifactFromSource(pluginDir)
	if err == nil {
		t.Fatal("expected embedded resource build without plugin.yaml to fail")
	}
	if !strings.Contains(err.Error(), "missing plugin.yaml") {
		t.Fatalf("expected missing embedded manifest error, got %v", err)
	}
}

func TestWriteRuntimeWasmArtifactFromSourceWritesGeneratedFile(t *testing.T) {
	pluginDir := t.TempDir()
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dynamic-write\nname: Dynamic Write\nversion: v0.1.0\ntype: dynamic\n",
	)

	repoRoot, ok := findRuntimeBuildRepoRoot(".")
	if !ok {
		t.Fatal("expected builder test to resolve repo root")
	}
	out, err := WriteRuntimeWasmArtifactFromSource(pluginDir, "")
	if err != nil {
		t.Fatalf("expected dynamic artifact write to succeed, got error: %v", err)
	}
	expectedPath := filepath.Join(repoRoot, defaultRuntimeOutputDir, "plugin-dynamic-write.wasm")
	if out.ArtifactPath != expectedPath {
		t.Fatalf("expected generated dynamic artifact path %s, got %s", expectedPath, out.ArtifactPath)
	}
	t.Cleanup(func() {
		_ = os.Remove(out.ArtifactPath)
		_ = os.RemoveAll(filepath.Join(repoRoot, defaultRuntimeOutputDir, runtimeWorkspaceDirName, "plugin-dynamic-write"))
	})

	content, err := os.ReadFile(out.ArtifactPath)
	if err != nil {
		t.Fatalf("expected generated dynamic artifact to exist, got error: %v", err)
	}
	if len(content) == 0 {
		t.Fatalf("expected generated dynamic artifact to contain bytes")
	}
	if _, err = os.Stat(filepath.Join(pluginDir, "temp", "plugin-dynamic-write.wasm")); !os.IsNotExist(err) {
		t.Fatalf("expected generated dynamic artifact to stop being written into plugin temp/, got err=%v", err)
	}
}

func TestWriteRuntimeWasmArtifactFromSourceSupportsExternalOutputDir(t *testing.T) {
	pluginDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "output")
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dynamic-output\nname: Dynamic Output\nversion: v0.1.0\ntype: dynamic\n",
	)

	out, err := WriteRuntimeWasmArtifactFromSource(pluginDir, outputDir)
	if err != nil {
		t.Fatalf("expected dynamic artifact write to external dir to succeed, got error: %v", err)
	}
	if expected := filepath.Join(outputDir, "plugin-dynamic-output.wasm"); out.ArtifactPath != expected {
		t.Fatalf("expected generated dynamic artifact path %s, got %s", expected, out.ArtifactPath)
	}
	if _, err = os.Stat(out.ArtifactPath); err != nil {
		t.Fatalf("expected generated dynamic artifact to exist in external dir, got error: %v", err)
	}
	if _, err = os.Stat(filepath.Join(pluginDir, "temp", "runtime-plugin.wasm")); !os.IsNotExist(err) {
		t.Fatalf("expected guest runtime wasm to stop being written into plugin temp/, got err=%v", err)
	}
}

func mustWriteFile(t *testing.T, filePath string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("failed to create directory for %s: %v", filePath, err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write file %s: %v", filePath, err)
	}
}

func readCommandOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func parseWasmCustomSections(content []byte) (map[string][]byte, error) {
	if len(content) < 8 {
		return nil, os.ErrInvalid
	}
	if string(content[:4]) != "\x00asm" {
		return nil, os.ErrInvalid
	}

	sections := make(map[string][]byte)
	cursor := 8
	for cursor < len(content) {
		sectionID := content[cursor]
		cursor++

		sectionSize, nextCursor, err := readULEB128(content, cursor)
		if err != nil {
			return nil, err
		}
		cursor = nextCursor

		end := cursor + int(sectionSize)
		if end > len(content) {
			return nil, os.ErrInvalid
		}

		if sectionID == 0 {
			nameLength, nameCursor, err := readULEB128(content, cursor)
			if err != nil {
				return nil, err
			}
			nameEnd := nameCursor + int(nameLength)
			if nameEnd > end {
				return nil, os.ErrInvalid
			}
			name := string(content[nameCursor:nameEnd])
			sections[name] = append([]byte(nil), content[nameEnd:end]...)
		}
		cursor = end
	}

	return sections, nil
}

func readULEB128(content []byte, cursor int) (uint32, int, error) {
	var (
		shift uint
		value uint32
	)

	for {
		if cursor >= len(content) {
			return 0, cursor, os.ErrInvalid
		}
		part := content[cursor]
		cursor++

		value |= uint32(part&0x7f) << shift
		if part&0x80 == 0 {
			return value, cursor, nil
		}
		shift += 7
		if shift > 28 {
			return 0, cursor, os.ErrInvalid
		}
	}
}
