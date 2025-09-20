package web

import (
	"embed"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strings"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed dist/* dist/**/*
var distContent embed.FS

var distFS fs.FS

func init() {
	var err error
	distFS, err = fs.Sub(distContent, "dist")
	if err != nil {
		distFS = distContent
	}
}

func ServeTemplate(w http.ResponseWriter, templateName string) error {
	content, err := templateFS.ReadFile(path.Join("templates", templateName))
	if err != nil {
		return err
	}
	_, err = w.Write(content)
	return err
}

func ServeAsset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	relative := strings.TrimPrefix(r.URL.Path, "/")
	if relative == "" {
		http.NotFound(w, r)
		return
	}
	if err := serveDistFile(w, relative); err != nil {
		http.NotFound(w, r)
	}
}

func ServeSPA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	relative := strings.TrimPrefix(r.URL.Path, "/")
	switch relative {
	case "", "table", "settings":
		if err := serveDistFile(w, "index.html"); err != nil {
			http.Error(w, "Failed to serve application", http.StatusInternalServerError)
		}
		return
	}
	if isAssetPath(relative) {
		if err := serveDistFile(w, relative); err != nil {
			http.NotFound(w, r)
		}
		return
	}
	if err := serveDistFile(w, "index.html"); err != nil {
		http.Error(w, "Failed to serve application", http.StatusInternalServerError)
	}
}

func isAssetPath(pathname string) bool {
	if strings.HasPrefix(pathname, "assets/") {
		return true
	}
	if strings.HasPrefix(pathname, "pwa/") {
		return true
	}
	if strings.HasPrefix(pathname, "webfonts/") {
		return true
	}
	switch pathname {
	case "manifest.json", "sw.js", "logo.png", "favicon.ico", "fa.min.css":
		return true
	}
	return false
}

func serveDistFile(w http.ResponseWriter, pathname string) error {
	cleanPath := path.Clean(pathname)
	if cleanPath == "." || strings.HasPrefix(cleanPath, "../") {
		return fs.ErrNotExist
	}
	file, err := distFS.Open(cleanPath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fs.ErrNotExist
	}

	ext := filepath.Ext(cleanPath)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		switch ext {
		case ".js":
			contentType = "application/javascript"
		case ".css":
			contentType = "text/css"
		case ".json":
			contentType = "application/json"
		case ".svg":
			contentType = "image/svg+xml"
		case ".woff":
			contentType = "font/woff"
		case ".woff2":
			contentType = "font/woff2"
		default:
			contentType = "application/octet-stream"
		}
	}
	w.Header().Set("Content-Type", contentType)
	_, err = io.Copy(w, file)
	return err
}
