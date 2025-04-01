package queue

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"os"
// 	"path/filepath"
// 	"sync"
// 	"sync/atomic"
// 	"time"
// 	"goaz/internal/azure"
// 	"goaz/internal/config"
// 	"goaz/internal/models"
// 	"github.com/fsnotify/fsnotify"
// )

// var UploadedCount int64

// // * StartWorkers initializes parallel upload workers and retry workers
// func StartWorkers(ctx context.Context, client *azure.AzureClient, imageQueue, retryQueue chan models.ImageTask, wg *sync.WaitGroup, cfg config.Config) {
// 	//? Parallel upload workers
// 	// go azure.RetryOfflineUploads(ctx, client, cfg)
// 	wg.Add(cfg.MaxWorkers)
// 	for i := 0; i < cfg.MaxWorkers; i++ {
// 		// wg.Add(1)
// 		go worker(ctx, client, imageQueue, retryQueue, wg, cfg, i)
// 	}

// 	//! Retry upload workers
// 	for i := 0; i < 10; i++ {
// 		// wg.Add(1)
// 		go retryWorker(ctx, client, retryQueue, wg, cfg, i)
// 	}
// }

// // * worker handles normal image uploads from imageQueue
// func worker(ctx context.Context, client *azure.AzureClient, imageQueue, retryQueue chan models.ImageTask, wg *sync.WaitGroup, cfg config.Config, id int) {
// 	defer wg.Done()
// 	log.Printf("🚀 Workerrrr-%d started", id)

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			log.Printf("⚠️ Workerrrr-%d shutting down", id)
// 			return
// 		case task, ok := <-imageQueue:
// 			if !ok {
// 				return
// 			}
// 			success := azure.UploadImage(ctx, client, task, wg, retryQueue, cfg)
// 			if success {
// 				newCount := atomic.AddInt64(&UploadedCount, 1)
// 				log.Printf("🚀 Total uploads: %d", newCount)
// 			} else {
// 				log.Printf("🔄 Retrying %s", task.BlobName)
// 				retryQueue <- task
// 				// ------------ added this
// 			}
// 		}
// 	}
// }

// // ! retryWorker handles failed uploads from retryQueue
// func retryWorker(ctx context.Context, client *azure.AzureClient, retryQueue chan models.ImageTask, wg *sync.WaitGroup, cfg config.Config, id int) {
// 	// defer wg.Done()
// 	log.Printf("🔄 Retry Worker-%d started", id)

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			log.Printf("⚠️ Retry Worker-%d shutting down", id)
// 			return
// 		case task, ok := <-retryQueue:
// 			if !ok {
// 				return
// 			}
// 			//!! Add delay before retrying ------- RECHECK
// 			// time.Sleep(10 * time.Millisecond)
// 			// log.Printf("🔄🔄🔄🔄🔄🔄🔄🔄🔄🔄🔄 --------------- Retrying %s", task.BlobName)
// 			// azure.UploadImage(ctx, client, task, wg, retryQueue, cfg)
// 			log.Printf("🔄🔄🔄🔄🔄🔄🔄🔄🔄🔄🔄🔄 Retrying %s (Attempt %d)", task.BlobName, task.Retry)
// 			// task.Retry++
// 			if task.Retry <= cfg.RetryLimit {
// 				task.Retry++
// 				azure.UploadImage(ctx, client, task, wg, retryQueue, cfg)
// 			}
// 			// else {
// 			// 	log.Printf("❌❌❌❌ Retry limit reached for %s", task.BlobName)

// 			// }
// 		}
// 	}

// }

// // ! SimulateImageStream mimics an incoming high-speed image stream
// func SimulateImageStream(ctx context.Context, imageQueue chan models.ImageTask, cfg config.Config) {
// 	for i := 1; i <= 50; i++ {
// 		select {
// 		case <-ctx.Done():
// 			log.Println("⚠️ Image stream simulation stopped")
// 			return
// 		default:
// 			blobName := fmt.Sprintf("dddd_%d.png", i)
// 			log.Printf("📸 Adding image to queue: %s", blobName)
// 			imageQueue <- models.ImageTask{ImagePath: cfg.ImageSource, BlobName: blobName, Retry: 0}
// 			time.Sleep(40 * time.Millisecond) //! Simulate high-speed stream
// 		}
// 	}
// }

// func TakeImagesfromFolderAndSendToQueue(ctx context.Context, imageQueue chan models.ImageTask, cfg config.Config) {
// 	sourceDir := cfg.ImageSource

// 	//? Step 1: Upload all existing images in the folder
// 	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		if !info.IsDir() && isImageFile(path) {
// 			blobName := filepath.Base(path)
// 			log.Printf("📸 Found existing image: %s", blobName)
// 			imageQueue <- models.ImageTask{ImagePath: path, BlobName: blobName, Retry: 0}
// 			time.Sleep(10 * time.Millisecond) //! Simulate high-speed stream
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		log.Fatalf("❌ Error scanning folder: %v", err)
// 	}

// 	//? Step 2: Watch for new images in the folder
// 	watcher, err := fsnotify.NewWatcher()
// 	if err != nil {
// 		log.Fatalf("❌ Failed to create file watcher: %v", err)
// 	}
// 	defer watcher.Close()

// 	err = watcher.Add(sourceDir)
// 	if err != nil {
// 		log.Fatalf("❌ Failed to watch directory: %v", err)
// 	}

// 	log.Println("👀 Watching folder for new images...")

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			log.Println("⚠️ Stopping folder watch...")
// 			return
// 		case event, ok := <-watcher.Events:
// 			if !ok {
// 				return
// 			}

// 			// If a new file is created, add it to the queue
// 			if event.Op&(fsnotify.Create) != 0 {
// 				if isImageFile(event.Name) {
// 					blobName := filepath.Base(event.Name)
// 					log.Printf("📸 New image detected: %s", blobName)
// 					imageQueue <- models.ImageTask{ImagePath: event.Name, BlobName: blobName, Retry: 0}
// 					time.Sleep(40 * time.Millisecond) //! Simulate high-speed stream
// 				}
// 			}
// 		case err, ok := <-watcher.Errors:
// 			if ok {
// 				log.Printf("⚠️ File watcher error: %v", err)
// 			}
// 		}
// 	}
// }

// // * isImageFile checks if a file is a valid image
// func isImageFile(filename string) bool {
// 	ext := filepath.Ext(filename)
// 	switch ext {
// 	case ".jpg", ".jpeg", ".png", ".bmp", ".gif", ".tiff":
// 		return true
// 	}
// 	return false
// }
