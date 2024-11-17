package backpaws

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tanner-caffrey/gwynnie-back-paws/photoutil"
)

const (
	port      = 8080
	photoDir  = "./photos"             // Directory to store photos
	staticDir = "./static"             // Directory for static files
	fileTypes = ".jpg,.jpeg,.png,.gif" // Supported file extensions
)

func StartServer() {
	// Ensure the photo directory exists
	if _, err := os.Stat(photoDir); os.IsNotExist(err) {
		log.Fatalf("Photo directory %s does not exist", photoDir)
	}

	// Set up handlers
	http.HandleFunc("/", listPhotosHandler)
	http.HandleFunc("/photos/", servePhotoHandler)
	http.HandleFunc("/upload", uploadPhotoHandler)

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
	fmt.Fprintln(w, `<a href="/upload">Upload a Photo</a><br><br>`)
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

// uploadPhotoHandler serves the upload page and processes uploads
func uploadPhotoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Serve the static HTML file
		http.ServeFile(w, r, filepath.Join(staticDir, "upload.html"))
	} else if r.Method == http.MethodPost {
		// Process the uploaded photo
		err := r.ParseMultipartForm(10 << 20) // 10 MB max file size
		if err != nil {
			http.Error(w, "Unable to process form data", http.StatusInternalServerError)
			return
		}

		// Retrieve title, description, and photo
		title := r.FormValue("title")
		description := r.FormValue("description")
		file, handler, err := r.FormFile("photo")
		if err != nil {
			http.Error(w, "Failed to retrieve file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Validate the file extension
		if !isValidPhoto(handler.Filename) {
			http.Error(w, "Invalid file type", http.StatusBadRequest)
			return
		}

		// Save the file to the photo directory
		photoPath := filepath.Join(photoDir, handler.Filename)
		out, err := os.Create(photoPath)
		if err != nil {
			http.Error(w, "Failed to save photo", http.StatusInternalServerError)
			return
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "Failed to save photo", http.StatusInternalServerError)
			return
		}

		// Update the photo list using photoutil
		photoListPath := filepath.Join(photoDir, "photo_list.json") // Adjust path if needed
		photoList, err := photoutil.GetPhotoList(photoListPath)
		if err != nil && !os.IsNotExist(err) {
			http.Error(w, "Failed to retrieve photo list", http.StatusInternalServerError)
			return
		}

		// Create a new photo entry
		newPhoto := photoutil.Photo{
			Filename:    handler.Filename,
			Title:       title,
			Description: description,
		}

		// Update or insert the new photo
		photoutil.UpdateOrInsertPhoto(&photoList, &newPhoto)

		// Save the updated photo list
		err = photoutil.WritePhotoList(photoListPath, photoList)
		if err != nil {
			http.Error(w, "Failed to update photo list", http.StatusInternalServerError)
			return
		}

		// Log the photo metadata
		log.Printf("Uploaded photo: %s, Title: %s, Description: %s\n", handler.Filename, title, description)

		// Redirect back to the gallery
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// isValidPhoto checks if a file has a valid photo extension
func isValidPhoto(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return strings.Contains(fileTypes, ext)
}
