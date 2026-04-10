// This file executes plugin SQL migrations and records abstract migration
// history entries for later review and lifecycle reconciliation.

package plugin

import (
	"context"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// pluginSQLAsset describes one install/uninstall SQL step after host extraction.
type pluginSQLAsset struct {
	Key     string
	Content string
}

// executeManifestSQLFiles executes plugin manifest SQL files and records every attempt in sys_plugin_migration.
func (s *Service) executeManifestSQLFiles(
	ctx context.Context,
	manifest *pluginManifest,
	direction pluginMigrationDirection,
) error {
	if manifest == nil {
		return gerror.New("plugin manifest cannot be nil")
	}

	sqlAssets, err := s.resolvePluginSQLAssets(manifest, direction)
	if err != nil {
		return err
	}

	for index, asset := range sqlAssets {
		if asset == nil {
			return gerror.New("插件 SQL 资源不能为空")
		}

		checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(asset.Content)))
		release, err := s.getPluginRelease(ctx, manifest.ID, manifest.Version)
		if err != nil {
			return err
		}
		if release == nil {
			return gerror.Newf("插件发布记录不存在: %s@%s", manifest.ID, manifest.Version)
		}
		// Migration records keep a stable execution key instead of persisting the
		// concrete SQL path. The host still executes files by directory convention,
		// but the database only stores review-oriented lifecycle metadata.
		migrationKey := s.buildPluginMigrationKey(direction, index+1)
		executedAt := gtime.Now()
		_, execErr := g.DB().Exec(ctx, asset.Content)
		if recordErr := s.recordPluginMigration(ctx, manifest.ID, release.Id, direction, migrationKey, index+1, checksum, executedAt, execErr); recordErr != nil {
			return recordErr
		}
		if execErr != nil {
			return gerror.Wrapf(execErr, "执行插件SQL失败: %s", asset.Key)
		}
	}
	return nil
}

// resolvePluginSQLAssets extracts lifecycle SQL either from embedded runtime artifact sections
// or from source-style directory conventions, while preserving execution order.
func (s *Service) resolvePluginSQLAssets(
	manifest *pluginManifest,
	direction pluginMigrationDirection,
) ([]*pluginSQLAsset, error) {
	if manifest == nil {
		return []*pluginSQLAsset{}, nil
	}

	if manifest.RuntimeArtifact != nil {
		embeddedAssets := manifest.RuntimeArtifact.InstallSQLAssets
		if direction == pluginMigrationDirectionUninstall {
			embeddedAssets = manifest.RuntimeArtifact.UninstallSQLAssets
		}
		if len(embeddedAssets) > 0 {
			items := make([]*pluginSQLAsset, 0, len(embeddedAssets))
			for _, asset := range embeddedAssets {
				if asset == nil {
					continue
				}
				items = append(items, &pluginSQLAsset{
					Key:     asset.Key,
					Content: asset.Content,
				})
			}
			return items, nil
		}
	}

	relativePaths := s.listPluginInstallSQLPaths(manifest)
	if direction == pluginMigrationDirectionUninstall {
		relativePaths = s.listPluginUninstallSQLPaths(manifest)
	}
	items := make([]*pluginSQLAsset, 0, len(relativePaths))
	for _, relativePath := range relativePaths {
		sqlContent, err := s.readSourcePluginAssetContent(manifest, relativePath)
		if err != nil {
			return nil, err
		}
		if sqlContent == "" {
			return nil, gerror.Newf("插件SQL文件为空: %s", relativePath)
		}
		items = append(items, &pluginSQLAsset{
			Key:     filepath.Base(relativePath),
			Content: sqlContent,
		})
	}
	return items, nil
}

func (s *Service) buildPluginMigrationKey(direction pluginMigrationDirection, sequenceNo int) string {
	normalizedDirection := strings.TrimSpace(strings.ToLower(direction.String()))
	if normalizedDirection == "" {
		normalizedDirection = pluginMigrationDirectionInstall.String()
	}
	if sequenceNo <= 0 {
		sequenceNo = 1
	}
	return fmt.Sprintf("%s-step-%03d", normalizedDirection, sequenceNo)
}

func (s *Service) getPluginMigration(
	ctx context.Context,
	pluginID string,
	releaseID int,
	phase pluginMigrationDirection,
	migrationKey string,
) (*entity.SysPluginMigration, error) {
	var migration *entity.SysPluginMigration
	err := dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{
			PluginId:     pluginID,
			ReleaseId:    releaseID,
			Phase:        phase.String(),
			MigrationKey: migrationKey,
		}).
		Scan(&migration)
	return migration, err
}

func (s *Service) recordPluginMigration(
	ctx context.Context,
	pluginID string,
	releaseID int,
	phase pluginMigrationDirection,
	migrationKey string,
	sequenceNo int,
	checksum string,
	executedAt *gtime.Time,
	execErr error,
) error {
	status := pluginMigrationExecutionStatusSucceeded
	message := ""
	if execErr != nil {
		status = pluginMigrationExecutionStatusFailed
		message = execErr.Error()
	}

	existing, err := s.getPluginMigration(ctx, pluginID, releaseID, phase, migrationKey)
	if err != nil {
		return err
	}

	data := do.SysPluginMigration{
		PluginId:       pluginID,
		ReleaseId:      releaseID,
		Phase:          phase.String(),
		MigrationKey:   migrationKey,
		ExecutionOrder: sequenceNo,
		Checksum:       checksum,
		Status:         status.String(),
		ErrorMessage:   message,
		ExecutedAt:     executedAt,
	}

	if existing == nil {
		_, err = dao.SysPluginMigration.Ctx(ctx).Data(data).Insert()
		return err
	}

	_, err = dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{Id: existing.Id}).
		Data(data).
		Update()
	return err
}
