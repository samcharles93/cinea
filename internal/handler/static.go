package handler

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

func NewStaticHandler(webDist embed.FS) http.Handler {
	// Get the dist subdirectory from the embedded files
	dist, err := fs.Sub(webDist, "web/dist")
	if err != nil {
		panic(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Clean the path to prevent directory traversal
		cleanPath := path.Clean(r.URL.Path)

		// Remove leading slash
		cleanPath = strings.TrimPrefix(cleanPath, "/")

		// Check if the path is for an asset
		if strings.HasPrefix(cleanPath, "assets/") {
			// Serve the asset directly
			http.FileServer(http.FS(dist)).ServeHTTP(w, r)
			return
		}

		// For all other paths, serve index.html
		file, err := dist.Open("index.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		indexData, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexData)
	})
}
