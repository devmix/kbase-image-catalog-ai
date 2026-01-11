package web

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var FS fs.FS

//go:embed static/*
//go:embed templates/*
var embedFS embed.FS
var localFS fs.FS

var useLocal bool

// InitTemplateFS initializes the template filesystem based on environment variable
func InitTemplateFS(useLocalFileSystem bool) {
	useLocal = useLocalFileSystem
	if useLocal {
		localFS = os.DirFS("web")
		FS = localFS
	} else {
		FS = embedFS
	}
}

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

	var bytes []byte
	if useLocal {
		content, err := fs.ReadFile(localFS, realPath)
		if err != nil {
			log.Printf("Error reading file system file %s: %v", realPath, err)
			http.NotFound(w, r)
			return
		}
		bytes = content
	} else {
		// Read the file from embedded realPath
		content, err := embedFS.ReadFile(realPath)
		if err != nil {
			log.Printf("Error reading embedded file %s: %v", realPath, err)
			http.NotFound(w, r)
			return
		}
		bytes = content
	}

	// Set content type
	w.Header().Set("Content-Type", getContentType(r.URL.Path))

	// Set cache headers for static assets
	if strings.HasPrefix(realPath, "static/") {
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}
	// Write the content
	w.Write(bytes)
}
