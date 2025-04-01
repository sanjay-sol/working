// package main

// import (
// 	"context"
// 	"log"
// 	"os"
// 	"os/signal"
// 	"sync"
// 	"sync/atomic"
// 	"syscall"
// 	"time"

// 	"goaz/internal/config"
// 	// "goaz/internal/models"
// 	"goaz/internal/queue"
// 	"goaz/services/capture"
// 	"goaz/services/retry"
// 	"goaz/services/uploader"
// 	"goaz/services/workers"
// 	// "github.com/Azure/azure-sdk-for-go/services"
// )

// func main() {
// 	cfg := config.LoadConfig()

// 	//* Create context with cancellation for graceful shutdown
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	//* Initialize Azure client
// 	client, err := uploader.NewAzureClient(*cfg)
// 	if err != nil {
// 		log.Fatalf("❌ Failed to create Azure client: %v", err)
// 	}

// 	//* Handle OS signals for clean shutdown
// 	sigChan := make(chan os.Signal, 1)
// 	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

// 	//* Image processing channels
// 	// imageQueue := make(chan models.ImageTask, 500)
// 	// retryQueue := make(chan models.ImageTask, 500)

// 	imageQueue := queue.NewQueue()
// 	retryQueue := queue.NewQueue()

// 	var wg sync.WaitGroup
// 	go func() {
// 		ticker := time.NewTicker(10 * time.Second)
// 		// defer ticker.Stop()

// 		for {
// 			select {
// 			case <-ctx.Done():
// 				log.Println("⚠️ Retry loop stopping due to context cancellation")
// 				return
// 			case <-ticker.C:
// 				log.Println("♻️♻️ Checking offline uploads for retry...")
// 				retry.RetryOfflineUploads(ctx, client, *cfg)
// 			}
// 		}
// 	}()

// 	// go func() {
// 	// 	for {
// 	// 		time.Sleep(10 * time.Second)
// 	// 		log.Println("♻️♻️ Checking offline uploads for retry...")
// 	// 		azure.RetryOfflineUploads(ctx, client, cfg)
// 	// 		// time.Sleep(5 * time.Second)
// 	// 		// log.Printf("📊 Total Images Uploaded So Far: %d", atomic.LoadInt64(&queue.UploadedCount))
// 	// 	}
// 	// }()

// 	//* Start workers
// 	workers.StartWorkers(ctx, client, imageQueue, retryQueue, &wg, *cfg)

// 	go func() {
// 		for {
// 			time.Sleep(5 * time.Second)
// 			log.Printf("📊 Total Images Uploaded So Far: %d", atomic.LoadInt64(&workers.UploadedCount))
// 		}
// 	}()

// 	//* Simulate high-speed image stream
// 	go func() {
// 		capture.TakeImagesfromFolderAndSendToQueue(ctx, imageQueue, *cfg)
// 		// close(imageQueue)
// 	}()

// 	// go queue.SimulateImageStream(ctx, imageQueue, cfg)

// 	//! Handle graceful shutdown on interrupt signal ---------- RECHECK AGAIIN
// 	go func() {
// 		<-sigChan
// 		log.Println("⚠️ Shutdown signal received. Closing queues and stopping workers...")
// 		cancel()
// 		close(imageQueue)
// 		close(retryQueue)
// 	}()

// 	//* Wait for all workers to complete
// 	wg.Wait()

// 	log.Println("✅ All uploads completed!")
// }

package main

import (
	"context"

	"goaz/internal/config"
	"goaz/internal/queue"
	"goaz/services/capture"
	"goaz/services/uploader"
	"goaz/services/workers"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	cfg := config.LoadConfig()

	//* Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//* Initialize Azure client
	client, err := uploader.NewAzureClient(*cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create Azure client: %v", err)
	}

	//* Handle OS signals for clean shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	//* Create queues
	imageQueue := queue.NewQueue()
	retryQueue := queue.NewQueue()

	var wg sync.WaitGroup
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop() // Ensure ticker stops

		for {
			select {
			case <-ctx.Done():
				log.Println("⚠️ Stopping upload progress ticker...")
				return
			case <-ticker.C:
				log.Printf("📊 Total Images Uploaded So Far: %d", atomic.LoadInt64(&workers.UploadedCount))
			}
		}
	}()

	//* Start upload workers
	workers.StartWorkers(ctx, client, imageQueue, retryQueue, &wg, *cfg)

	//* Monitor uploaded count
	go func() {
		for {
			time.Sleep(5 * time.Second)
			log.Printf("📊 Total Images Uploaded So Far: %d", atomic.LoadInt64(&workers.UploadedCount))
		}
	}()

	//* Start image capture service
	go capture.TakeImagesfromFolderAndSendToQueue(ctx, imageQueue, *cfg)

	//* Graceful shutdown handling
	// go func() {
	// 	<-sigChan
	// 	log.Println("⚠️ Shutdown signal received. Closing queues and stopping workers...")
	// 	cancel()
	// 	imageQueue.Close()
	// 	retryQueue.Close()

	// }()
	go func() {
		<-sigChan
		log.Println("⚠️ Shutdown signal received. Closing queues and stopping workers...")
		cancel() // Cancel the context (stops all workers)

		time.Sleep(1 * time.Second) // Small delay to allow cleanup

		wg.Wait() // Wait for all workers to finish
		log.Println("✅ All uploads completed! Exiting...")
		os.Exit(0)
	}()

	//* Wait for all workers to complete
	wg.Wait()

	log.Println("✅ All uploads completed!")
}
