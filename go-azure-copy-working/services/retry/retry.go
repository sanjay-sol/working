package retry

import (
	"context"
	"goaz/internal/config"
	"log"
	"os"
	"path/filepath"
	"time"

	"goaz/internal/models"
	"goaz/services/uploader"
)

func RetryOfflineUploads(ctx context.Context, azureClient *models.AzureClient, cfg config.Config) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("🔥 Panic in RetryOfflineUploads: %v", r)
		}
	}()

	if ctx.Err() != nil {
		log.Printf("❌ Context canceled, stopping RetryOfflineUploads: %v", ctx.Err())
		return
	}

	retryDir := cfg.RetryPath
	files, err := os.ReadDir(retryDir)
	if err != nil {
		log.Printf("❌ Error reading retry folder: %v", err)
		return
	}

	log.Printf("🔄 Found %d files to retry uploading", len(files))

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if ctx.Err() != nil {
			log.Printf("❌ Context canceled inside loop, stopping RetryOfflineUploads: %v", ctx.Err())
			return
		}

		filePath := filepath.Join(retryDir, file.Name())
		task := models.ImageTask{
			ID:         "123",
			FilePath:   filePath,
			BlobName:   file.Name(),
			RetryCount: 1,
			Timestamp:  time.Now(),
		}

		// Add timeout to prevent blocking
		ctxUpload, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		success := uploader.UploadImage(ctxUpload, azureClient, task, nil, nil, cfg)
		if success {
			if err := os.Remove(filePath); err != nil {
				log.Printf("⚠️ Failed to delete %s after upload: %v", filePath, err)
			} else {
				log.Printf("🗑️ Deleted %s after successful upload", filePath)
			}
		} else {
			log.Printf("❌ Upload failed for %s, keeping for future retry", filePath)
		}

		time.Sleep(20 * time.Millisecond)
	}

}
