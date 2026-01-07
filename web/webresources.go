package web

import (
	"embed"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed static/*
//go:embed templates/*
var FS embed.FS

// getContentType returns the appropriate content type for a file path
func getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".js":
		return "application/javascript"
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}

// HandleEmbeddedFile serves a file from the embedded filesystem
func HandleEmbeddedFile(w http.ResponseWriter, r *http.Request) {
	realPath := strings.TrimPrefix(r.URL.Path, "/")
	if realPath == "" {
		http.NotFound(w, r)
		return
	}

	// Read the file from embedded realPath
	content, err := FS.ReadFile(realPath)
	if err != nil {
		log.Printf("Error reading embedded file %s: %v", realPath, err)
		http.NotFound(w, r)
		return
	}
	// Set content type
	w.Header().Set("Content-Type", getContentType(r.URL.Path))

	// Set cache headers for static assets
	if strings.HasPrefix(realPath, "static/") {
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}
	// Write the content
	w.Write(content)
}
