package file

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

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
	File  *ghttp.UploadFile
	Scene string
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

// Upload handles file upload: computes SHA-256 hash, checks for duplicates, saves file via storage backend and records metadata in DB.
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

	// Compute SHA-256 hash
	hasher := sha256.New()
	if _, err = io.Copy(hasher, src); err != nil {
		return nil, gerror.Wrap(err, "计算文件散列值失败")
	}
	fileHash := hex.EncodeToString(hasher.Sum(nil))

	// Check for duplicate file by hash
	var existing *entity.SysFile
	err = dao.SysFile.Ctx(ctx).
		Where(dao.SysFile.Columns().Hash, fileHash).
		Scan(&existing)
	if err != nil {
		return nil, gerror.Wrap(err, "查询文件散列值失败")
	}
	if existing != nil {
		// Duplicate file found, reuse existing file's storage info but create a new record
		suffix := gstr.ToLower(gfile.ExtName(file.Filename))
		var userId int64
		if bizCtx := s.bizCtxSvc.Get(ctx); bizCtx != nil {
			userId = int64(bizCtx.UserId)
		}
		result, err := dao.SysFile.Ctx(ctx).Data(do.SysFile{
			Name:      existing.Name,
			Original:  file.Filename,
			Suffix:    suffix,
			Size:      file.Size,
			Hash:      fileHash,
			Url:       existing.Url,
			Path:      existing.Path,
			Engine:    existing.Engine,
			CreatedBy: userId,
		}).Insert()
		if err != nil {
			return nil, gerror.Wrap(err, "保存文件记录失败")
		}
		id, _ := result.LastInsertId()
		// Record usage scene if provided
		if in.Scene != "" {
			s.RecordUsage(ctx, &RecordUsageInput{
				FileId: id,
				Scene:  in.Scene,
			})
		}
		fullUrl := s.getBaseUrl(ctx) + existing.Url
		return &UploadOutput{
			Id:       id,
			Name:     existing.Name,
			Original: file.Filename,
			Url:      fullUrl,
			Suffix:   suffix,
			Size:     file.Size,
		}, nil
	}

	// Reset file reader position for storage
	if _, err = src.Seek(0, io.SeekStart); err != nil {
		return nil, gerror.Wrap(err, "重置文件读取位置失败")
	}

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
		Hash:      fileHash,
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
	// Record usage scene if provided
	if in.Scene != "" {
		s.RecordUsage(ctx, &RecordUsageInput{
			FileId: id,
			Scene:  in.Scene,
		})
	}
	// Return full URL with base URL prefix
	fullUrl := s.getBaseUrl(ctx) + url
	return &UploadOutput{
		Id:       id,
		Name:     storedName,
		Original: file.Filename,
		Url:      fullUrl,
		Suffix:   suffix,
		Size:     file.Size,
	}, nil
}

// ListInput defines input for file list query.
type ListInput struct {
	PageNum        int
	PageSize       int
	Name           string
	Original       string
	Suffix         string
	Scene          string
	BeginTime      string
	EndTime        string
	OrderBy        string
	OrderDirection string
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
	if in.Scene != "" {
		// Filter files that have the specified usage scene via subquery
		m = m.Where(
			dao.SysFile.Columns().Id+" IN (?)",
			dao.SysFileUsage.Ctx(ctx).Fields(dao.SysFileUsage.Columns().FileId).
				Where(dao.SysFileUsage.Columns().Scene, in.Scene),
		)
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	cols := dao.SysFile.Columns()
	orderBy := cols.Id
	allowedSortFields := map[string]string{
		"size":      cols.Size,
		"createdAt": cols.CreatedAt,
	}
	if in.OrderBy != "" {
		if field, ok := allowedSortFields[in.OrderBy]; ok {
			orderBy = field
		}
	}
	direction := "DESC"
	if in.OrderDirection == "asc" {
		direction = "ASC"
	}

	var files []*entity.SysFile
	err = m.Page(in.PageNum, in.PageSize).
		Order(orderBy + " " + direction).
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
				userNameMap[int64(u.Id)] = u.Username
			}
		}
	}

	// Build full URL prefix from HTTP request context
	baseUrl := s.getBaseUrl(ctx)

	items := make([]*ListOutputItem, len(files))
	for i, f := range files {
		fileCopy := *f
		if fileCopy.Url != "" && baseUrl != "" {
			fileCopy.Url = baseUrl + fileCopy.Url
		}
		items[i] = &ListOutputItem{
			SysFile:       &fileCopy,
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
	// Build full URL prefix from HTTP request context
	baseUrl := s.getBaseUrl(ctx)
	if baseUrl != "" {
		for _, f := range files {
			if f.Url != "" {
				f.Url = baseUrl + f.Url
			}
		}
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

// getBaseUrl returns the base URL (scheme + host) from the current HTTP request context.
func (s *Service) getBaseUrl(ctx context.Context) string {
	r := g.RequestFromCtx(ctx)
	if r == nil {
		return ""
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// SceneLabelMap maps scene identifiers to display labels.
var SceneLabelMap = map[string]string{
	"avatar":              "用户头像",
	"notice_image":        "通知公告图片",
	"notice_attachment":   "通知公告附件",
	"other":               "其他",
}

// SceneLabel returns the display label for a scene identifier.
func SceneLabel(scene string) string {
	if label, ok := SceneLabelMap[scene]; ok {
		return label
	}
	return scene
}

// UsageScenesOutput defines output for usage scenes list.
type UsageScenesOutput struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// UsageScenes returns all predefined usage scenes from SceneLabelMap.
func (s *Service) UsageScenes(ctx context.Context) ([]*UsageScenesOutput, error) {
	items := make([]*UsageScenesOutput, 0, len(SceneLabelMap))
	for value, label := range SceneLabelMap {
		items = append(items, &UsageScenesOutput{
			Value: value,
			Label: label,
		})
	}
	return items, nil
}

// DetailOutput defines output for file detail.
type DetailOutput struct {
	*entity.SysFile
	CreatedByName string             `json:"createdByName"`
	UsageScenes   []*DetailUsageItem `json:"usageScenes"`
}

// DetailUsageItem defines a single usage scene item in detail output.
type DetailUsageItem struct {
	Scene     string `json:"scene"`
	Label     string `json:"label"`
	CreatedAt string `json:"createdAt"`
}

// Detail returns file info with usage scenes.
func (s *Service) Detail(ctx context.Context, id int64) (*DetailOutput, error) {
	// Get file info
	var file *entity.SysFile
	err := dao.SysFile.Ctx(ctx).Where(dao.SysFile.Columns().Id, id).Scan(&file)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, gerror.New("文件不存在")
	}

	// Build full URL
	baseUrl := s.getBaseUrl(ctx)
	if baseUrl != "" && file.Url != "" {
		file.Url = baseUrl + file.Url
	}

	// Get uploader name
	var createdByName string
	if file.CreatedBy > 0 {
		var user *entity.SysUser
		err = dao.SysUser.Ctx(ctx).
			Where(dao.SysUser.Columns().Id, file.CreatedBy).
			Scan(&user)
		if err == nil && user != nil {
			createdByName = user.Username
		}
	}

	// Get usage scenes
	var usages []*entity.SysFileUsage
	err = dao.SysFileUsage.Ctx(ctx).
		Where(dao.SysFileUsage.Columns().FileId, id).
		OrderAsc(dao.SysFileUsage.Columns().CreatedAt).
		Scan(&usages)
	if err != nil {
		return nil, err
	}

	usageItems := make([]*DetailUsageItem, 0, len(usages))
	for _, u := range usages {
		createdAtStr := ""
		if u.CreatedAt != nil {
			createdAtStr = u.CreatedAt.String()
		}
		usageItems = append(usageItems, &DetailUsageItem{
			Scene:     u.Scene,
			Label:     SceneLabel(u.Scene),
			CreatedAt: createdAtStr,
		})
	}

	return &DetailOutput{
		SysFile:       file,
		CreatedByName: createdByName,
		UsageScenes:   usageItems,
	}, nil
}

// RecordUsageInput defines input for recording file usage.
type RecordUsageInput struct {
	FileId int64
	Scene  string
}

// RecordUsage creates a file usage record linking a file to a usage scene.
func (s *Service) RecordUsage(ctx context.Context, in *RecordUsageInput) error {
	if in.FileId <= 0 {
		return nil
	}
	_, err := dao.SysFileUsage.Ctx(ctx).Data(do.SysFileUsage{
		FileId: in.FileId,
		Scene:  in.Scene,
	}).Insert()
	return err
}

// DeleteUsageByScene deletes file usage records by scene.
func (s *Service) DeleteUsageByScene(ctx context.Context, scene string) error {
	_, err := dao.SysFileUsage.Ctx(ctx).
		Where(dao.SysFileUsage.Columns().Scene, scene).
		Delete()
	return err
}
