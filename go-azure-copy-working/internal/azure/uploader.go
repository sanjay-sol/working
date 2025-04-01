package azure

// import (
// 	"context"
// 	"fmt"
// 	"io"
// 	"log"
// 	"os"
// 	"path/filepath"
// 	"sync"
// 	"time"

// 	"goaz/internal/config"
// 	"goaz/internal/models"
// 	"goaz/internal/storage"

// 	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
// )

// type AzureClient struct {
// 	Client        *azblob.Client
// 	ContainerName string
// }

// func NewAzureClient(cfg config.Config) (*AzureClient, error) {
// 	cred, err := azblob.NewSharedKeyCredential(cfg.AzureAccountName, cfg.AzureAccountKey)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
// 	}

// 	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", cfg.AzureAccountName)
// 	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create Azure blob client: %w", err)
// 	}

// 	return &AzureClient{Client: client, ContainerName: cfg.Container}, nil
// }

// func UploadImage(ctx context.Context, azureClient *AzureClient, task models.ImageTask, wg *sync.WaitGroup, retryQueue chan models.ImageTask, cfg config.Config) bool {

// 	file, err := os.Open(task.ImagePath)
// 	if err != nil {
// 		log.Printf("Error opening %s: %v", task.BlobName, err)
// 		return false
// 	}
// 	defer file.Close()

// 	data, err := io.ReadAll(file)
// 	if err != nil {
// 		log.Printf("Failed to read file %s: %v", task.BlobName, err)
// 		return false
// 	}

// 	start := time.Now()
// 	resp, err := azureClient.Client.UploadBuffer(ctx, azureClient.ContainerName, task.BlobName, data, nil)
// 	elapsed := time.Since(start)

// 	if err != nil {
// 		if task.Retry < cfg.RetryLimit {
// 			log.Printf("🔄 Upload failed for %s (Attempt %d): %v", task.BlobName, task.Retry, err)
// 			task.Retry++
// 			// time.Sleep(time.Duration(task.Retry) * time.Second) // Exponential backoff
// 			time.Sleep(10 * time.Millisecond) // Linear backoff
// 			retryQueue <- task
// 		} else {
// 			log.Printf("❌❌❌ Retry limit reached, saving locally %s", task.BlobName)
// 			storage.SaveLocally2(data, task.BlobName, cfg.RetryFiles)
// 		}
// 		return false
// 	}

// 	log.Printf("✅ Successfully uploaded %s in %v (ETag: %s)", task.BlobName, elapsed, *resp.ETag)
// 	return true
// }
// func RetryOfflineUploads(ctx context.Context, azureClient *AzureClient, cfg config.Config) {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			log.Printf("🔥 Panic in RetryOfflineUploads: %v", r)
// 		}
// 	}()

// 	if ctx.Err() != nil {
// 		log.Printf("❌ Context canceled, stopping RetryOfflineUploads: %v", ctx.Err())
// 		return
// 	}

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

// 		if ctx.Err() != nil {
// 			log.Printf("❌ Context canceled inside loop, stopping RetryOfflineUploads: %v", ctx.Err())
// 			return
// 		}

// 		filePath := filepath.Join(retryDir, file.Name())
// 		task := models.ImageTask{
// 			ImagePath: filePath,
// 			BlobName:  file.Name(),
// 			Retry:     1,
// 		}

// 		// Add timeout to prevent blocking
// 		ctxUpload, cancel := context.WithTimeout(ctx, 10*time.Second)
// 		defer cancel()

// 		success := UploadImage(ctxUpload, azureClient, task, nil, nil, cfg)
// 		if success {
// 			if err := os.Remove(filePath); err != nil {
// 				log.Printf("⚠️ Failed to delete %s after upload: %v", filePath, err)
// 			} else {
// 				log.Printf("🗑️ Deleted %s after successful upload", filePath)
// 			}
// 		} else {
// 			log.Printf("❌ Upload failed for %s, keeping for future retry", filePath)
// 		}

// 		time.Sleep(20 * time.Millisecond)
// 	}
// }
