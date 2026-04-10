package v1

import "github.com/gogf/gf/v2/frame/g"

// InstallReq is the request for installing a dynamic plugin.
type InstallReq struct {
	g.Meta `path:"/plugins/{id}/install" method:"post" tags:"插件管理" summary:"安装动态插件" dc:"执行动态插件的安装生命周期，包括运行插件声明的安装SQL并将插件状态更新为已安装；源码插件随宿主编译集成，不支持调用该接口安装"`
	Id     string `json:"id" v:"required|length:1,64" dc:"插件唯一标识" eg:"plugin-demo-source"`
}

// InstallRes is the response for installing a dynamic plugin.
type InstallRes struct {
	Id        string `json:"id" dc:"插件唯一标识" eg:"plugin-demo-source"`
	Installed int    `json:"installed" dc:"安装状态：1=已安装 0=未安装" eg:"1"`
	Enabled   int    `json:"enabled" dc:"启用状态：1=启用 0=禁用" eg:"0"`
}
