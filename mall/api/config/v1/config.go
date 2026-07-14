package v1

import "github.com/gogf/gf/v2/frame/g"

type AppReq struct {
	g.Meta `path:"/config/app" method:"get" tags:"Config" summary:"Get app config"`
}

type AppRes struct {
	Name         string `json:"name" dc:"应用名称"`
	Env          string `json:"env" dc:"配置文件中的运行环境"`
	EffectiveEnv string `json:"effectiveEnv" dc:"命令行或环境变量覆盖后的运行环境"`
	Debug        bool   `json:"debug" dc:"是否开启调试"`
	Address      string `json:"address" dc:"HTTP监听地址"`
	PageSize     int    `json:"pageSize" dc:"默认分页大小"`
	DefaultSort  string `json:"defaultSort" dc:"默认排序"`
}
