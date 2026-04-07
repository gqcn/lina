package cmd

import (
	"context"
	"sort"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/internal/service/config"
)

type InitInput struct {
	g.Meta `name:"init" brief:"initialize database schema and seed data (DDL + seed DML)"`
}
type InitOutput struct{}

func (m *Main) Init(ctx context.Context, in InitInput) (out *InitOutput, err error) {
	sqlDir := config.New().GetInit(ctx).SqlDir
	files, err := scanInitSqlFiles(ctx, sqlDir)
	if err != nil {
		g.Log().Warningf(ctx, "failed to scan SQL files: %v", err)
		return nil, nil
	}
	if len(files) == 0 {
		g.Log().Warning(ctx, "no SQL files found for initialization")
		return
	}
	sort.Strings(files)
	execSqlFiles(ctx, files)

	g.Log().Info(ctx, "Database initialization completed.")
	return
}

func execSqlFiles(ctx context.Context, files []string) {
	for _, file := range files {
		sql := gfile.GetContents(file)
		if sql == "" {
			continue
		}
		g.Log().Infof(ctx, "Executing SQL file: %s", gfile.Basename(file))
		if _, err := g.DB().Exec(ctx, sql); err != nil {
			g.Log().Warningf(ctx, "execute %s: %v", gfile.Basename(file), err)
		}
	}
}

func scanInitSqlFiles(ctx context.Context, sqlDir string) ([]string, error) {
	var (
		files      = make([]string, 0)
		pluginRoot = gfile.RealPath(gfile.Join("..", "lina-plugins"))
	)

	if gfile.Exists(sqlDir) {
		coreFiles, err := gfile.ScanDirFile(sqlDir, "*.sql", false)
		if err != nil {
			return nil, err
		}
		files = append(files, coreFiles...)
	} else {
		g.Log().Warningf(ctx, "SQL directory does not exist: %s", sqlDir)
	}

	if pluginRoot == "" || !gfile.Exists(pluginRoot) {
		return files, nil
	}

	pluginEntries, err := gfile.ScanDir(pluginRoot, "*", false)
	if err != nil {
		return nil, err
	}
	for _, pluginPath := range pluginEntries {
		pluginSqlDir := gfile.Join(pluginPath, "manifest", "sql")
		if !gfile.Exists(pluginSqlDir) {
			continue
		}
		pluginFiles, scanErr := gfile.ScanDirFile(pluginSqlDir, "*.sql", false)
		if scanErr != nil {
			return nil, scanErr
		}
		files = append(files, pluginFiles...)
	}

	return files, nil
}
