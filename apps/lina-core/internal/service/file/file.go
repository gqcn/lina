package file

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
)

const (
	EngineLocal = "local"
)

// Service provides file management operations.
type Service struct {
	storage   Storage
	bizCtxSvc *bizctx.Service
}

// New creates and returns a new Service instance with local storage.
func New() *Service {
	ctx := context.Background()
	return &Service{
		storage:   NewLocalStorage(ctx, ""),
		bizCtxSvc: bizctx.New(),
	}
}

// UploadInput defines input for file upload.
type UploadInput struct {
	File *ghttp.UploadFile
}

// UploadOutput defines output for file upload.
type UploadOutput struct {
	Id       int64  `json:"id"`
	Name     string `json:"name"`
	Original string `json:"original"`
	Url      string `json:"url"`
	Suffix   string `json:"suffix"`
	Size     int64  `json:"size"`
}

// Upload handles file upload: saves file via storage backend and records metadata in DB.
func (s *Service) Upload(ctx context.Context, in *UploadInput) (*UploadOutput, error) {
	file := in.File
	if file == nil {
		return nil, gerror.New("请上传文件")
	}

	// Validate file size (max from config, default 10MB)
	maxSize := g.Cfg().MustGet(ctx, "upload.maxSize", 10).Int64()
	if file.Size > maxSize*1024*1024 {
		return nil, gerror.Newf("文件大小不能超过%dMB", maxSize)
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, gerror.Wrap(err, "打开上传文件失败")
	}
	defer src.Close()

	// Save via storage backend
	storagePath, err := s.storage.Put(ctx, file.Filename, src)
	if err != nil {
		return nil, gerror.Wrap(err, "保存文件失败")
	}

	// Build file metadata
	suffix := gstr.ToLower(gfile.ExtName(file.Filename))
	storedName := gfile.Basename(storagePath)
	url := s.storage.Url(ctx, storagePath)

	// Get current user ID
	var userId int64
	if bizCtx := s.bizCtxSvc.Get(ctx); bizCtx != nil {
		userId = int64(bizCtx.UserId)
	}

	// Insert file record
	result, err := dao.SysFile.Ctx(ctx).Data(do.SysFile{
		Name:      storedName,
		Original:  file.Filename,
		Suffix:    suffix,
		Size:      file.Size,
		Url:       url,
		Path:      storagePath,
		Engine:    EngineLocal,
		CreatedBy: userId,
	}).Insert()
	if err != nil {
		// Clean up stored file on DB error
		s.storage.Delete(ctx, storagePath)
		return nil, gerror.Wrap(err, "保存文件记录失败")
	}

	id, _ := result.LastInsertId()
	return &UploadOutput{
		Id:       id,
		Name:     storedName,
		Original: file.Filename,
		Url:      url,
		Suffix:   suffix,
		Size:     file.Size,
	}, nil
}

// ListInput defines input for file list query.
type ListInput struct {
	PageNum   int
	PageSize  int
	Name      string
	Original  string
	Suffix    string
	BeginTime string
	EndTime   string
}

// ListOutput defines output for file list.
type ListOutput struct {
	List  []*ListOutputItem `json:"list"`
	Total int               `json:"total"`
}

// ListOutputItem defines a single file item in list output.
type ListOutputItem struct {
	*entity.SysFile
	CreatedByName string `json:"createdByName"`
}

// List returns paginated file records.
func (s *Service) List(ctx context.Context, in *ListInput) (*ListOutput, error) {
	m := dao.SysFile.Ctx(ctx)

	if in.Name != "" {
		m = m.WhereLike(dao.SysFile.Columns().Name, fmt.Sprintf("%%%s%%", in.Name))
	}
	if in.Original != "" {
		m = m.WhereLike(dao.SysFile.Columns().Original, fmt.Sprintf("%%%s%%", in.Original))
	}
	if in.Suffix != "" {
		m = m.Where(dao.SysFile.Columns().Suffix, in.Suffix)
	}
	if in.BeginTime != "" {
		m = m.WhereGTE(dao.SysFile.Columns().CreatedAt, in.BeginTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(dao.SysFile.Columns().CreatedAt, in.EndTime)
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	var files []*entity.SysFile
	err = m.OrderDesc(dao.SysFile.Columns().Id).
		Page(in.PageNum, in.PageSize).
		Scan(&files)
	if err != nil {
		return nil, err
	}

	// Collect unique creator user IDs for name resolution
	userIdMap := make(map[int64]bool)
	for _, f := range files {
		if f.CreatedBy > 0 {
			userIdMap[f.CreatedBy] = true
		}
	}
	userNameMap := make(map[int64]string)
	if len(userIdMap) > 0 {
		userIds := make([]int64, 0, len(userIdMap))
		for uid := range userIdMap {
			userIds = append(userIds, uid)
		}
		var users []*entity.SysUser
		err = dao.SysUser.Ctx(ctx).
			WhereIn(dao.SysUser.Columns().Id, userIds).
			Scan(&users)
		if err == nil {
			for _, u := range users {
				name := u.Nickname
				if name == "" {
					name = u.Username
				}
				userNameMap[int64(u.Id)] = name
			}
		}
	}

	items := make([]*ListOutputItem, len(files))
	for i, f := range files {
		items[i] = &ListOutputItem{
			SysFile:       f,
			CreatedByName: userNameMap[f.CreatedBy],
		}
	}

	return &ListOutput{
		List:  items,
		Total: total,
	}, nil
}

// Info returns file info by ID.
func (s *Service) Info(ctx context.Context, id int64) (*entity.SysFile, error) {
	var file *entity.SysFile
	err := dao.SysFile.Ctx(ctx).Where(dao.SysFile.Columns().Id, id).Scan(&file)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, gerror.New("文件不存在")
	}
	return file, nil
}

// InfoByIds returns file info by multiple IDs.
func (s *Service) InfoByIds(ctx context.Context, ids []int64) ([]*entity.SysFile, error) {
	var files []*entity.SysFile
	err := dao.SysFile.Ctx(ctx).WhereIn(dao.SysFile.Columns().Id, ids).Scan(&files)
	if err != nil {
		return nil, err
	}
	return files, nil
}

// Delete removes files by IDs (soft delete in DB, also removes physical files).
func (s *Service) Delete(ctx context.Context, idsStr string) error {
	ids := gstr.SplitAndTrim(idsStr, ",")
	if len(ids) == 0 {
		return gerror.New("请选择要删除的文件")
	}

	idList := make([]int64, 0, len(ids))
	for _, idStr := range ids {
		idList = append(idList, gconv.Int64(idStr))
	}

	// Get file records first to delete physical files
	var files []*entity.SysFile
	err := dao.SysFile.Ctx(ctx).WhereIn(dao.SysFile.Columns().Id, idList).Scan(&files)
	if err != nil {
		return err
	}

	// Soft delete from DB
	_, err = dao.SysFile.Ctx(ctx).WhereIn(dao.SysFile.Columns().Id, idList).Delete()
	if err != nil {
		return err
	}

	// Delete physical files (best effort, don't fail on cleanup errors)
	for _, f := range files {
		s.storage.Delete(ctx, f.Path)
	}

	return nil
}

// GetStorage returns the underlying storage backend (for download use).
func (s *Service) GetStorage() Storage {
	return s.storage
}
