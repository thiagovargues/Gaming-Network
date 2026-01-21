package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"backend/internal/config"
	"backend/internal/http/middleware"
	"backend/internal/repo"
)

var allowedMIMEs = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
}

func UploadMedia(cfg config.Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := middleware.CurrentUser(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid form"})
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "file required"})
			return
		}
		defer file.Close()

		buf := make([]byte, 512)
		_, _ = file.Read(buf)
		_, _ = file.Seek(0, io.SeekStart)
		mimeType := http.DetectContentType(buf)
		if !allowedMIMEs[mimeType] {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "unsupported media type"})
			return
		}
		ext := extensionFor(mimeType, header.Filename)
		name, err := randomName()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "upload failed"})
			return
		}
		if err := os.MkdirAll(cfg.MediaDir, 0o755); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "upload failed"})
			return
		}

		filename := name + ext
		relPath := filepath.ToSlash(filepath.Join("media", filename))
		absPath := filepath.Join(cfg.MediaDir, filename)
		out, err := os.Create(absPath)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "upload failed"})
			return
		}
		defer out.Close()
		if _, err := io.Copy(out, file); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "upload failed"})
			return
		}

		_, _ = repo.CreateMedia(r.Context(), db, current.ID, relPath, mimeType)
		writeJSON(w, http.StatusOK, map[string]string{"path": relPath})
	}
}

func Media(cfg config.Config) http.Handler {
	fs := http.FileServer(http.Dir(cfg.MediaDir))
	return http.StripPrefix("/media/", fs)
}

func randomName() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func extensionFor(mimeType, filename string) string {
	if ext := strings.ToLower(filepath.Ext(filename)); ext != "" {
		return ext
	}
	ext, _ := mime.ExtensionsByType(mimeType)
	if len(ext) == 0 {
		return ""
	}
	return ext[0]
}
