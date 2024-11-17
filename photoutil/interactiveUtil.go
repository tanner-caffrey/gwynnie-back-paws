package photoutil

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func DefaultInteractiveConfig() PhotoUtilConfig {
	return PhotoUtilConfig{
		PhotoDir:      "./photos",
		PhotoListPath: "./photos/photos.json",
	}
}

func AddPhotosInteractive(conf PhotoUtilConfig) error {
	// Retrieve the existing photo list
	photoList, err := GetPhotoList(conf.PhotoListPath)
	if err != nil {
		return fmt.Errorf("failed to get photo list: %w", err)
	}

	// Create a scanner for reading user input
	scanner := bufio.NewScanner(os.Stdin)
	var photos []Photo

	// Interactive loop to get photo URLs
	fmt.Println("Enter photo URLs one by one. Type 'done' when finished:")
	for {
		fmt.Printf("Enter photo URL: ")
		if !scanner.Scan() {
			return fmt.Errorf("failed to read input: %w", scanner.Err())
		}
		url := scanner.Text()

		// Exit the loop if the user types "done"
		if url == "done" {
			break
		}

		// Download the photo
		filename, err := downloadPhoto(url, conf.PhotoDir)
		if err != nil {
			fmt.Printf("Error downloading photo from %s: %v\n", url, err)
			continue
		}

		// Create a new Photo object and add it to the list
		photo := Photo{Filename: filename}
		photos = append(photos, photo)
	}

	// Ask for title and description for each photo
	for i, photo := range photos {
		fmt.Printf("Enter title for photo (%s): ", photo.Filename)
		if scanner.Scan() {
			photo.Title = scanner.Text()
		} else {
			return fmt.Errorf("failed to read title: %w", scanner.Err())
		}

		fmt.Printf("Enter description for photo %d (%s): ", i+1, photo.Title)
		if scanner.Scan() {
			photo.Description = scanner.Text()
		} else {
			return fmt.Errorf("failed to read description: %w", scanner.Err())
		}

		// Update or insert the photo into the photo list
		UpdateOrInsertPhoto(&photoList, &photo)
	}

	// Write the updated photo list back to the file
	err = WritePhotoList(conf.PhotoListPath, photoList)
	if err != nil {
		return fmt.Errorf("failed to write photo list: %w", err)
	}

	fmt.Println("Photo list updated successfully!")
	return nil
}

func UpdatePhotosInteractive(conf PhotoUtilConfig) error {
	// Retrieve the existing photo list
	photoList, err := GetPhotoList(conf.PhotoListPath)
	if err != nil {
		return fmt.Errorf("failed to get photo list: %w", err)
	}

	// Retrieve the photos from the directory
	photos, err := GetPhotosFromDir(conf.PhotoDir)
	if err != nil {
		return fmt.Errorf("failed to get photos from directory: %w", err)
	}

	// Create a scanner for reading user input
	scanner := bufio.NewScanner(os.Stdin)

	// Ask for a description for each photo and update the photo list
	for i, photo := range photos {
		// exists := false
		// for _, p := range photoList.Photos {
		// 	if p.Filename == photo.Filename {
		// 		exists = true
		// 		break
		// 	}
		// }
		// if exists {
		// 	continue
		// }
		fmt.Printf("Enter description for photo %d (%s): ", i+1, photo.Title)

		// Read the entire line of input, including spaces
		if scanner.Scan() {
			photo.Description = scanner.Text()
		} else {
			return fmt.Errorf("failed to read description: %w", scanner.Err())
		}

		// Update or insert the photo into the photo list
		UpdateOrInsertPhoto(&photoList, &photo)
	}

	// Write the updated photo list back to the file
	err = WritePhotoList(conf.PhotoListPath, photoList)
	if err != nil {
		return fmt.Errorf("failed to write photo list: %w", err)
	}

	fmt.Println("Photo list updated successfully!")
	return nil
}

func GetPhotosFromDir(path string) ([]Photo, error) {
	// Check if the directory exists and is accessible
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	var photos []Photo

	// Iterate through the files in the directory
	for _, file := range files {
		// Skip directories
		if file.IsDir() {
			continue
		}

		// Check if the file has a valid image extension
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" {
			photo := Photo{
				Filename: file.Name(),
				Title:    strings.TrimSuffix(file.Name(), ext), // Use the filename (without extension) as a default title
				// Description can be left empty, as it's not available from the filesystem
			}
			photos = append(photos, photo)
		}
	}

	// Return an error if no photos are found
	if len(photos) == 0 {
		return nil, errors.New("no photos found in the directory")
	}

	return photos, nil
}

func downloadPhoto(url, dir string) (string, error) {
	// Get the response from the URL
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download photo from %s: %w", url, err)
	}
	defer resp.Body.Close()

	// Check if the response is successful
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download photo from %s: HTTP %d", url, resp.StatusCode)
	}

	// Get the filename from the URL
	filename := filepath.Base(url)
	if filename == "" {
		return "", fmt.Errorf("invalid URL, cannot determine filename: %s", url)
	}

	// Create the file in the specified directory
	filePath := filepath.Join(dir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	// Copy the photo content to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save photo to %s: %w", filePath, err)
	}

	return filename, nil
}
