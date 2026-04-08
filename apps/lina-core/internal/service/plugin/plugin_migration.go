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
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// executeManifestSQLFiles executes plugin manifest SQL files and records every attempt in sys_plugin_migration.
func (s *Service) executeManifestSQLFiles(
	ctx context.Context,
	manifest *pluginManifest,
	relativePaths []string,
	direction pluginMigrationDirection,
) error {
	if manifest == nil {
		return gerror.New("plugin manifest cannot be nil")
	}

	for index, relativePath := range relativePaths {
		sqlPath, err := s.resolvePluginResourcePath(manifest.RootDir, relativePath)
		if err != nil {
			return err
		}
		sqlContent := gfile.GetContents(sqlPath)
		if sqlContent == "" {
			return gerror.Newf("插件SQL文件为空: %s", sqlPath)
		}

		checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(sqlContent)))
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
		migration, err := s.getPluginMigration(ctx, manifest.ID, release.Id, direction, migrationKey)
		if err != nil {
			return err
		}
		if migration != nil && migration.Status == pluginMigrationExecutionStatusSucceeded.String() && migration.Checksum == checksum {
			continue
		}

		executedAt := gtime.Now()
		_, execErr := g.DB().Exec(ctx, sqlContent)
		if recordErr := s.recordPluginMigration(ctx, manifest.ID, release.Id, direction, migrationKey, index+1, checksum, executedAt, execErr); recordErr != nil {
			return recordErr
		}
		if execErr != nil {
			return gerror.Wrapf(execErr, "执行插件SQL失败: %s", filepath.Base(sqlPath))
		}
	}
	return nil
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
