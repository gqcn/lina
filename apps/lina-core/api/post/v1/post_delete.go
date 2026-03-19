package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Post Delete API

type DeleteReq struct {
	g.Meta `path:"/post/{ids}" method:"delete" tags:"岗位管理" summary:"删除岗位" dc:"删除一个或多个岗位，若岗位已分配给用户则不允许删除"`
	Ids    string `json:"ids" v:"required" dc:"岗位ID，多个用逗号分隔" eg:"1,2,3"`
}

type DeleteRes struct{}
