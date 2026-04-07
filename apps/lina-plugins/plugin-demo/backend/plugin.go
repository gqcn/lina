package backend

import "lina-core/pkg/pluginhost"

func init() {
	pluginhost.RegisterSourcePlugin(&pluginhost.SourcePlugin{
		ID: "plugin-demo",
		Hooks: []*pluginhost.HookSpec{
			{
				Event:  "auth.login.succeeded",
				Action: "insert",
				Table:  "plugin_demo_login_audit",
				Fields: map[string]string{
					"user_name":   "event.userName",
					"status":      "event.status",
					"ip":          "event.ip",
					"client_type": "event.clientType",
					"message":     "event.message",
					"login_time":  "now",
				},
			},
		},
		Resources: []*pluginhost.ResourceSpec{
			{
				Key:   "login-audits",
				Type:  "table-list",
				Table: "plugin_demo_login_audit",
				Fields: []*pluginhost.ResourceField{
					{Name: "id", Column: "id"},
					{Name: "userName", Column: "user_name"},
					{Name: "status", Column: "status"},
					{Name: "ip", Column: "ip"},
					{Name: "clientType", Column: "client_type"},
					{Name: "message", Column: "message"},
					{Name: "loginTime", Column: "login_time"},
				},
				Filters: []*pluginhost.ResourceFilter{
					{Param: "userName", Column: "user_name", Operator: "like"},
					{Param: "ip", Column: "ip", Operator: "like"},
					{Param: "status", Column: "status", Operator: "eq"},
					{Param: "beginTime", Column: "login_time", Operator: "gte-date"},
					{Param: "endTime", Column: "login_time", Operator: "lte-date"},
				},
				OrderBy: &pluginhost.OrderBySpec{
					Column:    "id",
					Direction: "desc",
				},
			},
		},
	})
}
