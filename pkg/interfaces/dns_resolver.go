package interfaces

import (
	"github.com/letronghoangminh/reproxy/pkg/config"
)

type DNSResolver interface {
	GetDynamicUpstreams(dynamicUpstreams []config.DynamicUpstreamConfig) ([]string, error)
}
