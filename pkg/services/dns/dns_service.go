package dns

import (
	"net"
	"strings"

	"github.com/letronghoangminh/reproxy/pkg/config"
)

func GetDynamicUpstreams(dynamicUpstreams []config.DynamicUpstreamConfig) []string {
	var upstreams []string
	for _, upstream := range dynamicUpstreams {
		var protocol, domain, port string
		
		if strings.Contains(upstream.Value, "://") {
			parts := strings.SplitN(upstream.Value, "://", 2)
			protocol = parts[0] + "://"
			domain = parts[1]
		} else {
			domain = upstream.Value
		}
		
		if strings.Contains(domain, ":") {
			parts := strings.Split(domain, ":")
			domain = parts[0]
			port = parts[1]
		}
		
		records, _ := net.LookupHost(domain)
		
		for i := range records {
			if port != "" {
				records[i] = protocol + records[i] + ":" + port
			} else {
				records[i] = protocol + records[i]
			}
		}
		
		upstreams = append(upstreams, records...)
	}
	
	return upstreams
}
