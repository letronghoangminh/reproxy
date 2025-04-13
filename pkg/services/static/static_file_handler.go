// Package static provides functionality to serve static files and handle static responses.
package static

import (
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/letronghoangminh/reproxy/pkg/config"
	"github.com/letronghoangminh/reproxy/pkg/interfaces"
	"github.com/letronghoangminh/reproxy/pkg/utils"
)

type StaticFileHandler struct {
	logger interfaces.Logger
}

func NewStaticFileHandler(logger interfaces.Logger) *StaticFileHandler {
	if logger == nil {
		logger = utils.GetLogger()
	}

	return &StaticFileHandler{
		logger: logger,
	}
}

func (h *StaticFileHandler) ServeFile(w http.ResponseWriter, r *http.Request, cfg *config.HandlerConfig) error {
	if cfg.StaticFiles.Root == "" {
		return errors.New("static file root is not configured")
	}

	requestPath := strings.TrimPrefix(r.URL.Path, cfg.Matchers.Path)
	cleanPath, err := h.sanitizePath(requestPath)
	if err != nil {
		h.logger.Error("Invalid file path", "path", requestPath, "error", err)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return err
	}

	filePath := filepath.Join(cfg.StaticFiles.Root, cleanPath)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			h.logger.Debug("File not found", "path", filePath)
			http.Error(w, "File not found", http.StatusNotFound)
			return err
		}
		h.logger.Error("Error accessing file", "path", filePath, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return err
	}

	if fileInfo.IsDir() {
		h.logger.Debug("Attempted to access directory", "path", filePath)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return errors.New("attempted to access directory: " + filePath)
	}

	h.addSecurityHeaders(w)

	h.logger.Debug("Serving static file", "file_path", filePath)
	addPoweredByHeader(w)

	http.ServeFile(w, r, filePath)
	return nil
}

func (h *StaticFileHandler) sanitizePath(requestPath string) (string, error) {
	requestPath = strings.TrimPrefix(requestPath, "/")

	cleanPath := path.Clean(requestPath)

	if cleanPath == ".." || strings.HasPrefix(cleanPath, "../") || strings.Contains(cleanPath, "/../") {
		return "", errors.New("invalid path: " + requestPath)
	}

	return cleanPath, nil
}

func (h *StaticFileHandler) addSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	w.Header().Set("Content-Security-Policy", "default-src 'self'")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
}

func addPoweredByHeader(w http.ResponseWriter) {
	w.Header().Set("X-Powered-By", "Reproxy")
}

func (h *StaticFileHandler) ServeStaticResponse(w http.ResponseWriter, r *http.Request, cfg *config.HandlerConfig) error {
	statusCode := cfg.StaticResponse.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	h.addSecurityHeaders(w)

	w.WriteHeader(statusCode)
	addPoweredByHeader(w)
	_, err := w.Write([]byte(cfg.StaticResponse.Body))
	if err != nil {
		h.logger.Error("Error writing response", "error", err)
		return err
	}

	return nil
}

var DefaultStaticFileHandler = NewStaticFileHandler(nil)

func ServeFile(w http.ResponseWriter, r *http.Request, cfg *config.HandlerConfig) error {
	return DefaultStaticFileHandler.ServeFile(w, r, cfg)
}

func ServeStaticResponse(w http.ResponseWriter, r *http.Request, cfg *config.HandlerConfig) error {
	return DefaultStaticFileHandler.ServeStaticResponse(w, r, cfg)
}
