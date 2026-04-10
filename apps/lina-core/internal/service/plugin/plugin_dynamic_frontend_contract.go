// This file validates how dynamic plugin menus consume hosted frontend assets.
// The host serves these assets from wasm-backed in-memory bundles, and enable-
// time validation keeps broken runtime menus from entering the router.

package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
)

const (
	pluginDynamicHostedAssetURLPrefix     = "/plugin-assets/"
	pluginDynamicPageComponentPath        = "system/plugin/dynamic-page"
	pluginDynamicMenuQueryKeyAccessMode   = "pluginAccessMode"
	pluginDynamicMenuAccessModeEmbedded   = "embedded-mount"
	pluginDynamicMenuEmbeddedJSExtension  = ".js"
	pluginDynamicMenuEmbeddedMJSExtension = ".mjs"
)

// ValidateRuntimeFrontendMenuBindings verifies that dynamic plugin menus only
// reference hosted assets that currently exist in the extracted workspace.
func (s *Service) ValidateRuntimeFrontendMenuBindings(ctx context.Context, manifest *pluginManifest) error {
	if manifest == nil || normalizePluginType(manifest.Type) != pluginTypeDynamic {
		return nil
	}

	menus, err := s.listPluginOwnedMenus(ctx, manifest.ID)
	if err != nil {
		return err
	}
	return s.validateRuntimeHostedMenuBindings(ctx, manifest, menus)
}

func (s *Service) listPluginOwnedMenus(ctx context.Context, pluginID string) ([]*entity.SysMenu, error) {
	columns := dao.SysMenu.Columns()
	prefixPattern := pluginMenuKeyPrefix + pluginID + ":%"
	remarkPattern := pluginMenuRemarkPrefix + pluginID + "%"

	var menus []*entity.SysMenu
	if err := dao.SysMenu.Ctx(ctx).
		WhereLike(columns.MenuKey, prefixPattern).
		WhereOrLike(columns.Remark, remarkPattern).
		OrderAsc(columns.Id).
		Scan(&menus); err != nil {
		return nil, err
	}
	return menus, nil
}

func (s *Service) validateRuntimeHostedMenuBindings(ctx context.Context, manifest *pluginManifest, menus []*entity.SysMenu) error {
	if manifest == nil || manifest.RuntimeArtifact == nil || len(menus) == 0 {
		return nil
	}

	var runtimeBundle *runtimeFrontendBundle
	for _, menu := range menus {
		if menu == nil || s.parsePluginIDFromMenu(menu) != manifest.ID {
			continue
		}

		relativeAssetPath, usesHostedAsset, err := s.resolveRuntimeHostedMenuAssetPath(manifest, menu.Path)
		if err != nil {
			return s.wrapRuntimeMenuValidationError(menu, err)
		}
		if !usesHostedAsset {
			continue
		}

		if runtimeBundle == nil {
			runtimeBundle, err = s.ensureRuntimeFrontendBundle(ctx, manifest)
			if err != nil {
				return s.wrapRuntimeMenuValidationError(menu, err)
			}
		}
		if !runtimeBundle.HasAsset(relativeAssetPath) {
			return s.wrapRuntimeMenuValidationError(
				menu,
				gerror.Newf("菜单引用的运行时前端资源不存在: %s", relativeAssetPath),
			)
		}

		queryParams, err := parseRuntimeMenuQueryParams(menu.QueryParam)
		if err != nil {
			return s.wrapRuntimeMenuValidationError(menu, err)
		}
		if err = validateRuntimeHostedMenuMode(menu, queryParams, relativeAssetPath); err != nil {
			return s.wrapRuntimeMenuValidationError(menu, err)
		}
	}
	return nil
}

func (s *Service) resolveRuntimeHostedMenuAssetPath(
	manifest *pluginManifest,
	menuPath string,
) (string, bool, error) {
	normalizedPath := normalizeRuntimeHostedMenuPath(menuPath)
	if !strings.HasPrefix(normalizedPath, pluginDynamicHostedAssetURLPrefix) {
		return "", false, nil
	}

	expectedPrefix := s.BuildRuntimeFrontendPublicBaseURL(manifest.ID, manifest.Version)
	if !strings.HasPrefix(normalizedPath, expectedPrefix) {
		return "", true, gerror.Newf(
			"菜单必须引用当前插件版本的托管资源: expected prefix %s",
			expectedPrefix,
		)
	}

	relativeAssetPath := strings.TrimPrefix(normalizedPath, expectedPrefix)
	if strings.TrimSpace(relativeAssetPath) == "" {
		relativeAssetPath = "index.html"
	}
	return normalizeRuntimeFrontendAssetPath(relativeAssetPath), true, nil
}

func (s *Service) wrapRuntimeMenuValidationError(menu *entity.SysMenu, err error) error {
	if menu == nil {
		return err
	}
	return gerror.Wrapf(err, "插件菜单校验失败[%s/%s]", strings.TrimSpace(menu.Name), strings.TrimSpace(menu.MenuKey))
}

func normalizeRuntimeHostedMenuPath(menuPath string) string {
	trimmedPath := strings.TrimSpace(menuPath)
	if trimmedPath == "" {
		return ""
	}
	if strings.HasPrefix(trimmedPath, "/") {
		return trimmedPath
	}
	return "/" + trimmedPath
}

func parseRuntimeMenuQueryParams(rawQuery string) (map[string]string, error) {
	trimmedQuery := strings.TrimSpace(rawQuery)
	if trimmedQuery == "" {
		return map[string]string{}, nil
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(trimmedQuery), &decoded); err != nil {
		return nil, gerror.Wrap(err, "菜单 query_param 不是合法 JSON")
	}

	queryParams := make(map[string]string, len(decoded))
	for key, value := range decoded {
		if strings.TrimSpace(key) == "" {
			continue
		}
		queryParams[key] = fmt.Sprint(value)
	}
	return queryParams, nil
}

func validateRuntimeHostedMenuMode(
	menu *entity.SysMenu,
	queryParams map[string]string,
	relativeAssetPath string,
) error {
	componentPath := strings.TrimSpace(menu.Component)
	accessMode := strings.TrimSpace(queryParams[pluginDynamicMenuQueryKeyAccessMode])
	isEmbeddedComponent := componentPath == pluginDynamicPageComponentPath

	if accessMode == pluginDynamicMenuAccessModeEmbedded {
		if !isEmbeddedComponent {
			return gerror.Newf(
				"宿主内嵌挂载菜单必须使用组件 %s",
				pluginDynamicPageComponentPath,
			)
		}
		if menu.IsFrame != 0 {
			return gerror.New("宿主内嵌挂载菜单不能声明为外链")
		}
		extension := strings.ToLower(filepath.Ext(relativeAssetPath))
		if extension != pluginDynamicMenuEmbeddedJSExtension && extension != pluginDynamicMenuEmbeddedMJSExtension {
			return gerror.New("宿主内嵌挂载入口必须指向 .js 或 .mjs ESM 资源")
		}
		return nil
	}

	if isEmbeddedComponent {
		return gerror.Newf(
			"使用组件 %s 的托管资源菜单必须声明 query_param.%s=%s",
			pluginDynamicPageComponentPath,
			pluginDynamicMenuQueryKeyAccessMode,
			pluginDynamicMenuAccessModeEmbedded,
		)
	}
	return nil
}
