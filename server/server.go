package backpaws

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	port      = 8080
	photoDir  = "./photos"             // Change this to your photo directory
	fileTypes = ".jpg,.jpeg,.png,.gif" // Supported file extensions
)

// Serve photo list and individual files
func StartServer() {
	// Ensure the photo directory exists
	if _, err := os.Stat(photoDir); os.IsNotExist(err) {
		log.Fatalf("Photo directory %s does not exist", photoDir)
	}

	http.HandleFunc("/", listPhotosHandler)
	http.HandleFunc("/photos/", servePhotoHandler)

	fmt.Printf("Starting server at http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// listPhotosHandler serves an HTML page listing all available photos
func listPhotosHandler(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(photoDir)
	if err != nil {
		http.Error(w, "Failed to read photo directory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintln(w, "<html><body>")
	fmt.Fprintln(w, "<h1>Photo Gallery</h1>")
	fmt.Fprintln(w, "<ul>")

	for _, file := range files {
		if file.IsDir() || !isValidPhoto(file.Name()) {
			continue
		}
		photoURL := "/photos/" + file.Name()
		fmt.Fprintf(w, `<li><a href="%s">%s</a></li>`, photoURL, file.Name())
	}

	fmt.Fprintln(w, "</ul>")
	fmt.Fprintln(w, "</body></html>")
}

// servePhotoHandler serves individual photo files
func servePhotoHandler(w http.ResponseWriter, r *http.Request) {
	photoName := strings.TrimPrefix(r.URL.Path, "/photos/")
	if photoName == "" {
		http.Error(w, "Photo not specified", http.StatusBadRequest)
		return
	}

	photoPath := filepath.Join(photoDir, photoName)
	if _, err := os.Stat(photoPath); os.IsNotExist(err) {
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, photoPath)
}

// isValidPhoto checks if a file has a valid photo extension
func isValidPhoto(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return strings.Contains(fileTypes, ext)
}
