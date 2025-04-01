package storage

import (
	"log"
	"os"
	"path/filepath"
)

// func SaveLocally(imagePath string, cfg config.Config) {
// 	if _, err := os.Stat(cfg.); os.IsNotExist(err) {
// 		_ = os.Mkdir(cfg.LocalStore, os.ModePerm)
// 	}

// 	destPath := fmt.Sprintf("%s/%s", cfg.LocalStore, imagePath)
// 	srcFile, err := os.Open(imagePath)
// 	if err != nil {
// 		log.Printf(" Failed to open file for local save: %v\n", err)
// 		return
// 	}
// 	defer srcFile.Close()

// 	destFile, err := os.Create(destPath)
// 	if err != nil {
// 		log.Printf("Failed to create local copy: %v\n", err)
// 		return
// 	}
// 	defer destFile.Close()

// 	_, _ = io.Copy(destFile, srcFile)
// 	log.Printf(" Saved failed upload locally: %s\n", destPath)
// }

func SaveLocally2(fileData []byte, fileName, retryFolder string) bool {

	if err := os.MkdirAll(retryFolder, os.ModePerm); err != nil {
		log.Printf("❌ Failed to create retry folder %s: %v", retryFolder, err)
		return false
	}

	destPath := filepath.Join(retryFolder, fileName)

	if err := os.WriteFile(destPath, fileData, os.ModePerm); err != nil {
		log.Printf("❌ Error saving file %s: %v", destPath, err)
		return false
	}

	log.Printf("💾 Saved the failed image locally: %s", destPath)
	return true
}

// // ! RetryOfflineUploads re-attempts failed uploads
// func RetryOfflineUploads(client *azblob.Client, imageQueue chan models.ImageTask, wg *sync.WaitGroup, cfg config.Config) {
// 	// Scan the retry folder for failed uploads
// 	retryDir := cfg.RetryFiles
// 	files, err := os.ReadDir(retryDir)
// 	if err != nil {
// 		log.Printf("❌ Error reading retry folder: %v", err)
// 		return
// 	}

// 	log.Printf("🔄 Found %d files to retry uploading", len(files))

// 	for _, file := range files {
// 		if file.IsDir() {
// 			continue
// 		}

// 		filePath := filepath.Join(retryDir, file.Name())
// 		imageQueue <- models.ImageTask{
// 			ImagePath: filePath,
// 			BlobName:  file.Name(),
// 			Retry:     1, // Start retry count from 1
// 		}
// 		log.Printf("🔄 Queued %s for re-upload", file.Name())
// 		time.Sleep(20 * time.Millisecond) // Throttle re-queueing
// 	}
// }
