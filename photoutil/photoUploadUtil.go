package photoutil

import (
	"encoding/json"
	"fmt"
	"os"
)

type PhotoList struct {
	Path   string  `json:"path"`
	Photos []Photo `json:"photos"`
}

type Photo struct {
	Filename    string `json:"filename"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type PhotoUtilConfig struct {
	PhotoDir      string `json:"photoDir"`
	PhotoListPath string `json:"photoListPath"`
}

func GetPhotoList(path string) (PhotoList, error) {
	f, err := os.Open(path)
	if err != nil {
		return PhotoList{}, fmt.Errorf("cannot open photo list %s: %w", path, err)
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	var photoList PhotoList
	if err = decoder.Decode(&photoList); err != nil {
		return PhotoList{}, fmt.Errorf("failed to decode photo list from %s: %w", path, err)
	}

	return photoList, nil
}

func WritePhotoList(path string, photoList PhotoList) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot write photo list %s: %w", path, err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "	")
	if err := encoder.Encode(photoList); err != nil {
		return fmt.Errorf("failed to encode photo list to %s: %w", path, err)
	}

	return nil
}

func UpdateOrInsertPhoto(photoList *PhotoList, newPhoto *Photo) {
	for i, photo := range photoList.Photos {
		if photo.Filename == newPhoto.Filename {
			photoList.Photos[i] = *newPhoto
			return
		}
	}
	photoList.Photos = append(photoList.Photos, *newPhoto)
}

func DeletePhoto(photoList *PhotoList, photoToDelete Photo) error {
	for i, photo := range photoList.Photos {
		if photo.Filename == photoToDelete.Filename {
			photoList.Photos = append(photoList.Photos[:i], photoList.Photos[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("photo with filename %s not found", photoToDelete.Filename)
}

func UpdateOrInsertPhotoList(photoList *PhotoList, photos []Photo) {
	for _, photo := range photos {
		UpdateOrInsertPhoto(photoList, &photo)
	}
}
