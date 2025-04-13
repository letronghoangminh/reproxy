package matcher

import (
	"net"
	"net/http"
	"slices"
	"strings"

	"github.com/letronghoangminh/reproxy/pkg/config"
	"github.com/letronghoangminh/reproxy/pkg/interfaces"
	"github.com/letronghoangminh/reproxy/pkg/utils"
)

type RequestMatcher struct {
	logger interfaces.Logger
}

func NewRequestMatcher(logger interfaces.Logger) interfaces.Matcher {
	if logger == nil {
		logger = utils.GetLogger()
	}

	return &RequestMatcher{
		logger: logger,
	}
}

func (m *RequestMatcher) MatchHandler(r *http.Request, handlers []*config.HandlerConfig) *config.HandlerConfig {
	m.logger.Debug("Matching request",
		"path", r.URL.Path,
		"method", r.Method,
		"host", r.Host,
		"remote_addr", r.RemoteAddr)

	for i, handler := range handlers {
		if !m.matchMethod(r, handler) {
			continue
		}

		if !m.matchPath(r, handler) {
			continue
		}

		if !m.matchHeaders(r, handler) {
			continue
		}

		if !m.matchQuery(r, handler) {
			continue
		}

		if !m.matchClientCIDR(r, handler) {
			continue
		}

		m.logger.Debug("Request matched", "handler_index", i)
		return handler
	}

	m.logger.Debug("No handler matched")
	return nil
}

func (m *RequestMatcher) matchMethod(r *http.Request, handler *config.HandlerConfig) bool {
	if len(handler.Matchers.Method) == 0 {
		return true
	}

	return slices.Contains(handler.Matchers.Method, r.Method) ||
		slices.Contains(handler.Matchers.Method, "*")
}

func (m *RequestMatcher) matchPath(r *http.Request, handler *config.HandlerConfig) bool {
	if handler.Matchers.Path == "" {
		return true
	}

	return strings.HasPrefix(r.URL.Path, handler.Matchers.Path)
}

func (m *RequestMatcher) matchHeaders(r *http.Request, handler *config.HandlerConfig) bool {
	if len(handler.Matchers.Headers) == 0 {
		return true
	}

	for key, value := range handler.Matchers.Headers {
		if r.Header.Get(key) != value {
			return false
		}
	}

	return true
}

func (m *RequestMatcher) matchQuery(r *http.Request, handler *config.HandlerConfig) bool {
	if len(handler.Matchers.Query) == 0 {
		return true
	}

	for key, value := range handler.Matchers.Query {
		if r.URL.Query().Get(key) != value {
			return false
		}
	}

	return true
}

func (m *RequestMatcher) matchClientCIDR(r *http.Request, handler *config.HandlerConfig) bool {
	if len(handler.Matchers.ClientCIDRs) == 0 {
		return true
	}

	clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		m.logger.Error("Failed to parse client IP", "remote_addr", r.RemoteAddr)
		clientIP = r.RemoteAddr // Fall back to using the whole string
	}

	parsedIP := net.ParseIP(clientIP)
	if parsedIP == nil {
		m.logger.Error("Invalid client IP", "client_ip", clientIP)
		return false
	}

	for _, cidr := range handler.Matchers.ClientCIDRs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			m.logger.Error("Invalid CIDR", "cidr", cidr)
			continue
		}

		if ipNet.Contains(parsedIP) {
			return true
		}
	}

	return false
}

var DefaultMatcher = NewRequestMatcher(nil)

func MatchHandler(r *http.Request, handlers []*config.HandlerConfig) *config.HandlerConfig {
	return DefaultMatcher.MatchHandler(r, handlers)
}
