package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// File Upload API

type UploadReq struct {
	g.Meta `path:"/file/upload" method:"post" mime:"multipart/form-data" tags:"文件管理" summary:"上传文件" dc:"上传单个文件到服务器，支持常见文件格式，文件信息自动记录到文件管理表中"`
}

type UploadRes struct {
	Id       int64  `json:"id" dc:"文件ID" eg:"1"`
	Name     string `json:"name" dc:"存储文件名" eg:"20260319_abc12345.png"`
	Original string `json:"original" dc:"原始文件名" eg:"avatar.png"`
	Url      string `json:"url" dc:"文件访问URL" eg:"/api/v1/uploads/2026/03/20260319_abc12345.png"`
	Suffix   string `json:"suffix" dc:"文件后缀" eg:"png"`
	Size     int64  `json:"size" dc:"文件大小（字节）" eg:"102400"`
}
