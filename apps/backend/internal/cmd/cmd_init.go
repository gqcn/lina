package cmd

import (
	"context"
	"sort"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
)

type InitInput struct {
	g.Meta `name:"init" brief:"initialize database by executing SQL files"`
}
type InitOutput struct{}

func (m *Main) Init(ctx context.Context, in InitInput) (out *InitOutput, err error) {
	sqlDir := g.Cfg().MustGet(ctx, "init.sqlDir", "manifest/sql").String()
	if !gfile.Exists(sqlDir) {
		g.Log().Warningf(ctx, "SQL directory does not exist: %s", sqlDir)
		return
	}

	// Scan SQL files in the directory (non-recursive)
	files, err := gfile.ScanDirFile(sqlDir, "*.sql", false)
	if err != nil {
		g.Log().Warningf(ctx, "failed to scan SQL directory %s: %v", sqlDir, err)
		return nil, nil
	}
	if len(files) == 0 {
		g.Log().Warning(ctx, "no SQL files found in directory: ", sqlDir)
		return
	}

	// Sort by ASCII order (ascending)
	sort.Strings(files)

	// Execute SQL files in order
	for _, file := range files {
		sql := gfile.GetContents(file)
		if sql == "" {
			continue
		}
		g.Log().Infof(ctx, "Executing SQL file: %s", gfile.Basename(file))
		if _, err = g.DB().Exec(ctx, sql); err != nil {
			g.Log().Warningf(ctx, "execute %s: %v", gfile.Basename(file), err)
		}
	}

	g.Log().Info(ctx, "Database initialization completed.")
	return
}
