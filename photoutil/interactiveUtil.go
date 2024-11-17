package photoutil

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type PhotoUtilInteractiveConfig struct {
	PhotoDir      string `json:"photoDir"`
	PhotoListPath string `json:"photoListPath"`
}

func DefaultInteractiveConfig() PhotoUtilInteractiveConfig {
	return PhotoUtilInteractiveConfig{
		PhotoDir:      "./photos",
		PhotoListPath: "./persist/photos.json",
	}
}

func UpdatePhotosInteractive(conf PhotoUtilInteractiveConfig) error {
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
		fmt.Printf("Enter description for photdo %d (%s): ", i+1, photo.Title)

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
