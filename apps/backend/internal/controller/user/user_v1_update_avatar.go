package user

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/grand"

	v1 "backend/api/user/v1"
)

func (c *ControllerV1) UpdateAvatar(ctx context.Context, req *v1.UpdateAvatarReq) (res *v1.UpdateAvatarRes, err error) {
	r := g.RequestFromCtx(ctx)
	file := r.GetUploadFile("avatarfile")
	if file == nil {
		return nil, gerror.New("请上传头像文件")
	}

	// Validate file type
	ext := gstr.ToLower(gfile.ExtName(file.Filename))
	allowed := map[string]bool{"jpg": true, "jpeg": true, "png": true, "gif": true, "webp": true}
	if !allowed[ext] {
		return nil, gerror.New("不支持的文件格式，仅支持 jpg/jpeg/png/gif/webp")
	}

	// Validate file size (max 5MB)
	if file.Size > 5*1024*1024 {
		return nil, gerror.New("文件大小不能超过5MB")
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s_%s.%s", gtime.Now().Format("Ymd_His"), grand.S(8), ext)

	// Save to uploads directory
	basePath := g.Cfg().MustGet(ctx, "upload.path", "upload").String()
	uploadDir := gfile.Join(basePath, "avatars")
	gfile.Mkdir(uploadDir)
	file.Filename = filename
	_, err = file.Save(uploadDir)
	if err != nil {
		return nil, gerror.Wrap(err, "保存头像文件失败")
	}

	// Build URL path
	avatarUrl := "/api/uploads/avatars/" + filename

	// Update user avatar in database
	err = c.userSvc.UpdateAvatar(ctx, avatarUrl)
	if err != nil {
		gfile.Remove(gfile.Join(uploadDir, filename))
		return nil, err
	}

	return &v1.UpdateAvatarRes{
		Url: avatarUrl,
	}, nil
}
