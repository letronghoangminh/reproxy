// Package dns provides functionality to resolve DNS records and cache the results.
package dns

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/letronghoangminh/reproxy/pkg/config"
	"github.com/letronghoangminh/reproxy/pkg/interfaces"
	"github.com/letronghoangminh/reproxy/pkg/utils"
)

type DNSCache struct {
	Records  []string
	ExpireAt time.Time
}

type DNSResolver struct {
	cache map[string]DNSCache
	mutex sync.RWMutex
	ttl   time.Duration
}

func NewDNSResolver(ttl time.Duration) interfaces.DNSResolver {
	return &DNSResolver{
		cache: make(map[string]DNSCache),
		ttl:   ttl,
	}
}

func (r *DNSResolver) GetDynamicUpstreams(dynamicUpstreams []config.DynamicUpstreamConfig) ([]string, error) {
	var upstreams []string
	var errors []error

	for _, upstream := range dynamicUpstreams {
		parsedValue := upstream.Value
		if !strings.Contains(parsedValue, "://") {
			parsedValue = "http://" + parsedValue
		}

		parsedURL, err := url.Parse(parsedValue)
		if err != nil {
			errors = append(errors, fmt.Errorf("invalid URL %q: %w", upstream.Value, err))
			continue
		}

		host := parsedURL.Host
		domain := host
		port := parsedURL.Port()

		if strings.Contains(host, ":") {
			var err error
			domain, port, err = net.SplitHostPort(host)
			if err != nil {
				errors = append(errors, fmt.Errorf("invalid host:port %q: %w", host, err))
				continue
			}
		}

		cacheKey := upstream.Type + ":" + domain
		r.mutex.RLock()
		cache, found := r.cache[cacheKey]
		r.mutex.RUnlock()

		var records []string
		if found && time.Now().Before(cache.ExpireAt) {
			records = cache.Records
			utils.GetLogger().Debug("DNS cache hit", "domain", domain, "type", upstream.Type)
		} else {
			utils.GetLogger().Debug("DNS cache miss", "domain", domain, "type", upstream.Type)

			var lookupErr error
			switch upstream.Type {
			case "A", "AAAA":
				records, lookupErr = net.LookupHost(domain)
			case "CNAME":
				var cname string
				cname, lookupErr = net.LookupCNAME(domain)
				if lookupErr == nil {
					records = []string{strings.TrimSuffix(cname, ".")}
				}
			default:
				lookupErr = fmt.Errorf("unsupported DNS record type: %s", upstream.Type)
			}

			if lookupErr != nil {
				errors = append(errors, fmt.Errorf("DNS lookup failed for %q (%s): %w",
					domain, upstream.Type, lookupErr))
				continue
			}

			r.mutex.Lock()
			r.cache[cacheKey] = DNSCache{
				Records:  records,
				ExpireAt: time.Now().Add(r.ttl),
			}
			r.mutex.Unlock()
		}

		scheme := parsedURL.Scheme
		for _, ip := range records {
			if port != "" {
				upstreams = append(upstreams, fmt.Sprintf("%s://%s:%s", scheme, ip, port))
			} else {
				upstreams = append(upstreams, fmt.Sprintf("%s://%s", scheme, ip))
			}
		}
	}

	if len(errors) > 0 {
		for _, err := range errors {
			utils.GetLogger().Error("DNS resolution error", "error", err)
		}
		return upstreams, fmt.Errorf("encountered %d DNS resolution errors", len(errors))
	}

	return upstreams, nil
}

func (r *DNSResolver) ClearCache() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.cache = make(map[string]DNSCache)
}

var DefaultDNSResolver = NewDNSResolver(5 * time.Minute)

func GetDynamicUpstreams(dynamicUpstreams []config.DynamicUpstreamConfig) ([]string, error) {
	return DefaultDNSResolver.GetDynamicUpstreams(dynamicUpstreams)
}
