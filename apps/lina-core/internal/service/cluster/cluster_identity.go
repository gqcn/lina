package cluster

import (
	"os"

	"github.com/gogf/gf/v2/net/gipv4"
)

func generateNodeIdentifier() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname, _ = gipv4.GetIntranetIp()
	}
	if hostname == "" {
		hostname = "local-node"
	}
	return hostname
}
