package capture

import (
	"context"

	"goaz/internal/config"
	"goaz/internal/models"
	"goaz/internal/queue"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TakeImagesfromFolderAndSendToQueue(ctx context.Context, imageQueue *queue.Queue, cfg config.Config) {
	sourceDir := cfg.ImageSource

	//? Step 1: Upload all existing images in the folder
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isImageFile(path) {
			blobName := filepath.Base(path)
			log.Printf("📸 Found existing image: %s", blobName)
			imageQueue.Push(models.ImageTask{FilePath: path, BlobName: blobName, RetryCount: 0})
			// imageQueue <- models.ImageTask{FilePath: path, BlobName: blobName, RetryCount: 0}
			time.Sleep(10 * time.Millisecond) //! Simulate high-speed stream
		}
		return nil
	})
	if err != nil {
		log.Fatalf("❌ Error scanning folder: %v", err)
	}

	//? Step 2: Watch for new images in the folder
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("❌ Failed to create file watcher: %v", err)
	}
	defer watcher.Close()

	err = watcher.Add(sourceDir)
	if err != nil {
		log.Fatalf("❌ Failed to watch directory: %v", err)
	}

	log.Println("👀 Watching folder for new images...")

	for {
		select {
		case <-ctx.Done():
			log.Println("⚠️ Stopping folder watch...")
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// If a new file is created, add it to the queue
			if event.Op&(fsnotify.Create) != 0 {
				if isImageFile(event.Name) {
					blobName := filepath.Base(event.Name)

					log.Printf("📸 New image detected: %s", blobName)
					imageQueue.Push(models.ImageTask{FilePath: event.Name, BlobName: blobName, RetryCount: 0})
					// imageQueue <- models.ImageTask{FilePath: event.Name, BlobName: blobName, RetryCount: 0}
					time.Sleep(40 * time.Millisecond) //! Simulate high-speed stream
				}
			}
		case err, ok := <-watcher.Errors:
			if ok {
				log.Printf("⚠️ File watcher error: %v", err)
			}
		}
	}
}

// * isImageFile checks if a file is a valid image

func isImageFile(filename string) bool {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".bmp", ".gif", ".tiff", ".raw":
		return true
	}
	return false
}

// func TakeImagesfromFolderAndSendToQueue(ctx context.Context, imageQueue *queue.Queue, cfg config.Config) {
// 	watcher, err := fsnotify.NewWatcher()
// 	if err != nil {
// 		log.Fatalf("❌ Failed to create file watcher: %v", err)
// 	}
// 	defer watcher.Close()

// 	// Process all existing images in each folder
// 	for _, sourceDir := range cfg.ImageSource {
// 		err := processExistingImages(sourceDir, imageQueue)
// 		if err != nil {
// 			log.Printf("❌ Error scanning folder %s: %v", sourceDir, err)
// 		}

// 		// Add each directory to the watcher
// 		err = watcher.Add(sourceDir)
// 		if err != nil {
// 			log.Fatalf("❌ Failed to watch directory %s: %v", sourceDir, err)
// 		}
// 	}

// 	log.Println("👀 Watching folders for new images...")

// 	// Watch for new files
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			log.Println("⚠️ Stopping folder watch...")
// 			return
// 		case event, ok := <-watcher.Events:
// 			if !ok {
// 				return
// 			}

// 			// Handle new file creation event
// 			if event.Op&fsnotify.Create != 0 && isImageFile(event.Name) {
// 				blobName := filepath.Base(event.Name)
// 				log.Printf("📸 New image detected: %s", blobName)
// 				imageQueue.Push(models.ImageTask{FilePath: event.Name, BlobName: blobName, RetryCount: 0})
// 				time.Sleep(40 * time.Millisecond) // Simulate high-speed stream
// 			}
// 		case err, ok := <-watcher.Errors:
// 			if ok {
// 				log.Printf("⚠️ File watcher error: %v", err)
// 			}
// 		}
// 	}
// }

// // processExistingImages uploads existing images in the folder
// func processExistingImages(sourceDir string, imageQueue *queue.Queue) error {
// 	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		if !info.IsDir() && isImageFile(path) {
// 			blobName := filepath.Base(path)
// 			log.Printf("📸 Found existing image: %s", blobName)
// 			imageQueue.Push(models.ImageTask{FilePath: path, BlobName: blobName, RetryCount: 0})
// 			time.Sleep(10 * time.Millisecond) // Simulate high-speed stream
// 		}
// 		return nil
// 	})
// }

// // isImageFile checks if a file is a valid image
// func isImageFile(filename string) bool {
// 	ext := filepath.Ext(filename)
// 	switch ext {
// 	case ".jpg", ".jpeg", ".png", ".bmp", ".gif", ".tiff":
// 		return true
// 	}
// 	return false
// }
