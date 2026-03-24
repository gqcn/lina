package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Notice Delete API

type DeleteReq struct {
	g.Meta `path:"/notice/{ids}" method:"delete" tags:"通知公告" summary:"删除通知公告" dc:"删除一条或多条通知公告"`
	Ids    string `json:"ids" v:"required" dc:"公告ID，多个用逗号分隔" eg:"1,2,3"`
}

// DeleteRes Notice delete response
type DeleteRes struct{}
