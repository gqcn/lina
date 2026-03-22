package cmd

import (
	"context"
	"sort"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/internal/service/config"
)

type MockInput struct {
	g.Meta `name:"mock" brief:"load mock/demo data from manifest/sql/mock-data/"`
}
type MockOutput struct{}

func (m *Main) Mock(ctx context.Context, in MockInput) (out *MockOutput, err error) {
	sqlDir := config.New().GetInit(ctx).SqlDir
	mockDir := gfile.Join(sqlDir, "mock-data")
	if !gfile.Exists(mockDir) {
		g.Log().Warningf(ctx, "mock-data directory does not exist: %s", mockDir)
		return
	}

	files, err := gfile.ScanDirFile(mockDir, "*.sql", false)
	if err != nil {
		g.Log().Warningf(ctx, "failed to scan mock-data directory %s: %v", mockDir, err)
		return nil, nil
	}
	if len(files) == 0 {
		g.Log().Warning(ctx, "no SQL files found in directory: ", mockDir)
		return
	}
	sort.Strings(files)
	execSqlFiles(ctx, files)

	g.Log().Info(ctx, "Mock data loaded.")
	return
}
