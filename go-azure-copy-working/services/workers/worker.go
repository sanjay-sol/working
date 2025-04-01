// package workers

// import (
// 	"context"
// 	"goaz/internal/config"
// 	"goaz/internal/models"
// 	"goaz/internal/queue"
// 	"goaz/services/uploader"
// 	"log"
// 	"sync"
// 	"sync/atomic"
// )

// var UploadedCount int64

// // * StartWorkers initializes parallel upload workers and retry workers
// func StartWorkers(ctx context.Context, client *models.AzureClient, imageQueue, retryQueue *queue.Queue, wg *sync.WaitGroup, cfg config.Config) {
// 	//? Parallel upload workers
// 	// go azure.RetryOfflineUploads(ctx, client, cfg)
// 	wg.Add(cfg.MaxWorkers)
// 	for i := 0; i < cfg.MaxWorkers; i++ {
// 		// wg.Add(1)
// 		go worker(ctx, client, imageQueue, retryQueue, wg, cfg, i)
// 	}

// 	// //! Retry upload workers
// 	// for i := 0; i < 10; i++ {
// 	// 	// wg.Add(1)
// 	// 	go retryWorker(ctx, client, retryQueue, wg, cfg, i)
// 	// }
// }

// // * worker handles normal image uploads from imageQueue
// func worker(ctx context.Context, client *models.AzureClient, imageQueue, retryQueue *queue.Queue, wg *sync.WaitGroup, cfg config.Config, id int) {
// 	defer wg.Done()
// 	log.Printf("🚀 Workerrrr-%d started", id)

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			log.Printf("⚠️ Workerrrr-%d shutting down", id)
// 			return
// 		case task, ok := imageQueue.Pop()
// 			if !ok {
// 				return
// 			}
// 			success := uploader.UploadImage(ctx, client, task, wg, retryQueue, cfg)
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

// func StartRetryWorkers(ctx context.Context, client *models.AzureClient, retryQueue chan models.ImageTask, wg *sync.WaitGroup, cfg config.Config) {

// 	wg.Add(cfg.MaxWorkers)
// 	for i := 0; i < cfg.MaxWorkers; i++ {
// 		go retryWorker(ctx, client, retryQueue, wg, cfg, i)
// 	}
// }

// // ! retryWorker handles failed uploads from retryQueue
// func retryWorker(ctx context.Context, client *models.AzureClient, retryQueue chan models.ImageTask, wg *sync.WaitGroup, cfg config.Config, id int) {
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
// 			log.Printf("🔄🔄🔄🔄🔄🔄🔄🔄🔄🔄🔄🔄 Retrying %s (Attempt %d)", task.BlobName, task.RetryCount)
// 			// task.Retry++
// 			if task.RetryCount <= cfg.RetryLimit {
// 				task.RetryCount++
// 				uploader.UploadImage(ctx, client, task, wg, retryQueue, cfg)
// 			}
// 			// else {
// 			// 	log.Printf("❌❌❌❌ Retry limit reached for %s", task.BlobName)

// 			// }
// 		}
// 	}

// }

package workers

import (
	"context"

	"goaz/internal/config"
	"goaz/internal/models"
	"goaz/internal/queue"
	"goaz/services/uploader"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

var UploadedCount int64

// StartWorkers initializes parallel upload and retry workers
func StartWorkers(ctx context.Context, client *models.AzureClient, imageQueue, retryQueue *queue.Queue, wg *sync.WaitGroup, cfg config.Config) {
	wg.Add(cfg.MaxWorkers)

	for i := 0; i < cfg.MaxWorkers; i++ {
		go worker(ctx, client, imageQueue, retryQueue, wg, cfg, i)
	}

	// // Start retry workers
	// wg.Add(cfg.TotalRetryWorkers)
	// for i := 0; i < cfg.TotalRetryWorkers; i++ {
	// 	go retryWorker(ctx, client, retryQueue, wg, cfg, i)
	// }
}

func StartRetryWorkers(ctx context.Context, client *models.AzureClient, retryQueue *queue.Queue, wg *sync.WaitGroup, cfg config.Config) {
	wg.Add(cfg.TotalRetryWorkers)

	for i := 0; i < cfg.TotalRetryWorkers; i++ {
		go retryWorker(ctx, client, retryQueue, wg, cfg, i)
	}
}

// worker handles image uploads
func worker(ctx context.Context, client *models.AzureClient, imageQueue, retryQueue *queue.Queue, wg *sync.WaitGroup, cfg config.Config, id int) {
	defer wg.Done()
	log.Printf("🚀 Worker-%d started", id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("⚠️ Worker-%d shutting down", id)
			return
		default:
			task, ok := imageQueue.PopWithTimeout(100 * time.Millisecond)
			if !ok {
				time.Sleep(10 * time.Millisecond) // Prevent busy looping
				continue
			}

			success := uploader.UploadImage(ctx, client, task, wg, retryQueue, cfg)
			if success {
				newCount := atomic.AddInt64(&UploadedCount, 1)
				log.Printf("✅ Worker-%d uploaded: %s (Total: %d)", id, task.BlobName, newCount)
			} else {
				log.Printf("🔄 Worker-%d retrying %s", id, task.BlobName)
				retryQueue.Push(task)
			}
		}
	}
}

// // retryWorker handles failed uploads
// func retryWorker(ctx context.Context, client *models.AzureClient, retryQueue *queue.Queue, wg *sync.WaitGroup, cfg config.Config, id int) {
// 	defer wg.Done()
// 	log.Printf("🔄 Retry Worker-%d started", id)

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			log.Printf("⚠️ Retry Worker-%d shutting down", id)
// 			return
// 		default:
// 			task, ok := retryQueue.Pop()
// 			if !ok {
// 				time.Sleep(50 * time.Millisecond)
// 				continue
// 			}

// 			if task.RetryCount >= cfg.RetryLimit {
// 				log.Printf("❌ Retry Worker-%d: Retry limit reached for %s", id, task.BlobName)
// 				continue
// 			}

// 			task.RetryCount++
// 			time.Sleep(500 * time.Millisecond) // Add delay before retrying
// 			success := uploader.UploadImage(ctx, client, task, wg, retryQueue, cfg)
// 			if success {
// 				log.Printf("✅ Retry Worker-%d successfully retried: %s", id, task.BlobName)
// 			} else {
// 				log.Printf("🔄 Retry Worker-%d failed again, re-queuing %s (Attempt %d)", id, task.BlobName, task.RetryCount)
// 				retryQueue.Push(task)
// 			}
// 		}
// 	}
// }

func retryWorker(ctx context.Context, client *models.AzureClient, retryQueue *queue.Queue, wg *sync.WaitGroup, cfg config.Config, id int) {
	defer wg.Done()
	log.Printf("🔄 Retry Worker-%d started", id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("⚠️ Retry Worker-%d shutting down", id)
			return
		default:
			task, ok := retryQueue.PopWithTimeout(100 * time.Millisecond)

			if !ok {
				time.Sleep(50 * time.Millisecond) // Prevent busy looping
				continue
			}

			if task.RetryCount >= cfg.RetryLimit {
				log.Printf("❌ Retry Worker-%d: Retry limit reached for %s", id, task.BlobName)
				continue
			}

			task.RetryCount++
			time.Sleep(500 * time.Millisecond) // Add delay before retrying
			success := uploader.UploadImage(ctx, client, task, wg, retryQueue, cfg)
			if success {
				log.Printf("✅ Retry Worker-%d successfully retried: %s", id, task.BlobName)
			} else {
				log.Printf("🔄 Retry Worker-%d failed again, re-queuing %s (Attempt %d)", id, task.BlobName, task.RetryCount)
				retryQueue.Push(task)
			}
		}
	}
}
