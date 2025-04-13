package main

import (
	"context"

	"goaz/internal/config"
	"goaz/internal/queue"
	"goaz/services/retry"
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

	//* Create retry queue
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

	//* Start retry workers
	workers.StartRetryWorkers(ctx, client, retryQueue, &wg, *cfg)

	//* Retry offline uploads periodically
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("⚠️ Retry loop stopping due to context cancellation")
				return
			case <-ticker.C:
				log.Println("♻️ Checking offline uploads for retry...")
				retry.RetryOfflineUploads(ctx, client, *cfg)
			}
		}
	}()

	//* Graceful shutdown handling
	// go func() {
	// 	<-sigChan
	// 	log.Println("⚠️ Shutdown signal received. Stopping retry workers...")
	// 	cancel()
	// 	retryQueue.Close()
	// 	// wg.Wait()
	// 	// os.Exit(0)
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

	//* Wait for all retry workers to complete
	wg.Wait()

	log.Println("✅ Retry process completed!")
}
