package uploader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"goaz/internal/config"
	"goaz/internal/models"
	"goaz/internal/queue"
	"goaz/internal/storage"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
)

func NewAzureClient(cfg config.Config) (*models.AzureClient, error) {
	cred, err := azblob.NewSharedKeyCredential(cfg.AzureAccountName, cfg.AzureAccountKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", cfg.AzureAccountName)
	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure blob client: %w", err)
	}

	return &models.AzureClient{Client: client, ContainerName: cfg.Container, Credential: cred}, nil
}

func CreateContainer(ctx context.Context, azureClient *models.AzureClient, containerName string) error {
	// Get a reference to the container
	_, err := azureClient.Client.CreateContainer(ctx, containerName, &container.CreateOptions{})
	if err != nil {
		// Check if the error is due to the container already existing
		// if containerErr, ok := err.(*container.StorageError); ok && containerErr.ErrorCode == container.ErrorCodeContainerAlreadyExists {
		// 	log.Printf("ℹ️ Container %s already exists", containerName)
		// 	return nil // No need to return an error
		// }
		return fmt.Errorf("failed to create container %s: %w", containerName, err)
	}

	log.Printf("✅ Successfully created container: %s", containerName)
	return nil
}

func UploadImage(ctx context.Context, azureClient *models.AzureClient, task models.ImageTask, wg *sync.WaitGroup, retryQueue *queue.Queue, cfg config.Config) bool {
	file, err := os.Open(task.FilePath)
	if err != nil {
		log.Printf("Error opening %s: %v", task.BlobName, err)
		return false
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Failed to read file %s: %v", task.BlobName, err)
		return false
	}

	// Detect content type
	contentType := "application/octet-stream" // Default
	if strings.HasSuffix(task.BlobName, ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(task.BlobName, ".jpg") || strings.HasSuffix(task.BlobName, ".jpeg") {
		contentType = "image/jpeg"
	} else if strings.HasSuffix(task.BlobName, ".raw") {
		contentType = "application/octet-stream"
	}

	// ✅ Fix: Use correct `BlobHTTPHeaders` from `azblob/blob`
	options := &azblob.UploadBufferOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: &contentType,
		},
	}

	start := time.Now()
	// containerName := strings.Split(task.BlobName, "_")[0]
	resp, err := azureClient.Client.UploadBuffer(ctx, cfg.Container, task.BlobName, data, options)
	elapsed := time.Since(start)

	if err != nil {
		if task.RetryCount < cfg.RetryLimit {
			log.Printf("🔄 Upload failed for %s (Attempt %d): %v", task.BlobName, task.RetryCount, err)
			task.RetryCount++
			time.Sleep(10 * time.Millisecond) // Linear backoff
			retryQueue.Push(task)
		} else {
			log.Printf("❌ Retry limit reached, saving locally: %s", task.BlobName)
			storage.SaveLocally2(data, task.BlobName, cfg.RetryPath)
		}
		return false
	}

	log.Printf("✅ Uploaded %s (Content-Type: %s) in %v (ETag: %s)", task.BlobName, contentType, elapsed, *resp.ETag)
	return true
}

func ListBlobs(ctx context.Context, azureClient *models.AzureClient) error {
	pager := azureClient.Client.NewListBlobsFlatPager(azureClient.ContainerName, nil)
	total := 0
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list blobs: %w", err)
		}

		for _, blob := range resp.Segment.BlobItems {
			total++
			log.Printf("📝 Blob: %s | Size: %d bytes", *blob.Name, *blob.Properties.ContentLength)
		}
	}

	log.Println("✅ Finished listing blobs. toatal blobs:", total)
	return nil
}

func DownloadBlobs(ctx context.Context, azureClient *models.AzureClient, downloadDir string) error {
	pager := azureClient.Client.NewListBlobsFlatPager(azureClient.ContainerName, nil)

	// Create download directory if it doesn't exist
	if err := os.MkdirAll(downloadDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create download directory: %w", err)
	}

	total := 0
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list blobs: %w", err)
		}

		for _, blob := range resp.Segment.BlobItems {
			blobName := *blob.Name
			localFilePath := filepath.Join(downloadDir, blobName)

			// Ensure parent directories exist
			if err := os.MkdirAll(filepath.Dir(localFilePath), os.ModePerm); err != nil {
				log.Printf("⚠️ Failed to create directory for %s: %v", localFilePath, err)
				continue
			}

			// Download blob
			if err := downloadBlob(ctx, azureClient, blobName, localFilePath); err != nil {
				log.Printf("❌ Failed to download %s: %v", blobName, err)
				continue
			}

			log.Printf("✅ Downloaded: %s -> %s", blobName, localFilePath)
			total++
		}
	}

	log.Printf("📥 Finished downloading blobs. Total files: %d", total)
	return nil
}

func downloadBlob(ctx context.Context, azureClient *models.AzureClient, blobName, localFilePath string) error {
	// Create local file
	file, err := os.Create(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", localFilePath, err)
	}
	defer file.Close()

	// Download blob
	resp, err := azureClient.Client.DownloadStream(ctx, azureClient.ContainerName, blobName, nil)
	if err != nil {
		return fmt.Errorf("failed to get blob response for %s: %w", blobName, err)
	}
	defer resp.Body.Close()

	// Copy data from blob to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write blob data to file %s: %w", localFilePath, err)
	}

	return nil
}

func GenerateSASURL(accountName, containerName, blobName string, cred *azblob.SharedKeyCredential) (string, error) {
	expiryTime := time.Now().UTC().Add(1 * time.Hour) // 1 hour expiry

	permissions := sas.BlobPermissions{Read: true} // Ensure "Read" is enabled

	sasValues := sas.BlobSignatureValues{
		Version:       "2023-11-03", // Use latest SAS version
		ContainerName: containerName,
		BlobName:      blobName,
		Permissions:   permissions.String(),
		ExpiryTime:    expiryTime,
	}

	// Sign with shared key
	sasQuery, err := sasValues.SignWithSharedKey(cred)
	if err != nil {
		return "", fmt.Errorf("failed to generate SAS token: %w", err)
	}

	// Generate valid SAS URL
	sasURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s?%s",
		accountName, containerName, blobName, sasQuery.Encode())

	return sasURL, nil
}

// List all blobs and return them as JSON with signed URLs
func ListBlobsAsJSON(ctx context.Context, azureClient *models.AzureClient, cfg config.Config, containerName string) ([]byte, error) {
	pager := azureClient.Client.NewListBlobsFlatPager(containerName, nil)
	var blobs []map[string]interface{}

	// Use the stored credential instead of calling GetSharedKeyCredential()
	cred := azureClient.Credential

	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list blobs: %w", err)
		}

		for _, blob := range resp.Segment.BlobItems {
			sasURL, err := GenerateSASURL(cfg.AzureAccountName, containerName, *blob.Name, cred)
			if err != nil {
				log.Printf("⚠️ Could not generate SAS URL for %s: %v", *blob.Name, err)
				continue
			}

			blobs = append(blobs, map[string]interface{}{
				"name": *blob.Name,
				"size": *blob.Properties.ContentLength,
				"url":  sasURL, // Use the signed URL
			})
		}
	}

	data, err := json.Marshal(blobs)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Total blobs retrieved: %d", len(blobs))
	return data, nil
}

func DeleteAllBlobsFromContainer(ctx context.Context, azureClient *models.AzureClient, container string) error {
	pager := azureClient.Client.NewListBlobsFlatPager(container, nil)
	totalDeleted := 0

	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list blobs: %w", err)
		}

		for _, blob := range resp.Segment.BlobItems {
			blobName := *blob.Name

			// Attempt to delete the blob
			_, err := azureClient.Client.DeleteBlob(ctx, container, blobName, nil)
			if err != nil {
				log.Printf("❌ Failed to delete %s: %v", blobName, err)
				continue
			}

			log.Printf("🗑️ Deleted: %s", blobName)
			totalDeleted++
		}
	}

	log.Printf("✅ Finished deleting blobs. Total deleted: %d", totalDeleted)
	return nil
}
