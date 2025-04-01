package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"goaz/internal/config"
	"goaz/internal/queue"
	"goaz/internal/storage"
)

func main() {
	cfg := config.LoadConfig()

	store, err := storage.NewStore(cfg.DBPath, cfg.StoragePath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	q := queue.NewQueue()

	go watchFolder(cfg.StoragePath, q, store)

	// Simulate processing (this would be handled by the uploader service)
	for {
		task, ok := q.Pop()
		if ok {
			fmt.Printf("Processing image: %s\n", task.FilePath)
		}
		time.Sleep(1 * time.Second)
	}
}

// watchFolder detects new images.
func watchFolder(path string, q *queue.Queue, store *storage.Store) {
	for {
		files, err := os.ReadDir(path)
		if err != nil {
			log.Println("Failed to read directory:", err)
			continue
		}

		for _, file := range files {
			if filepath.Ext(file.Name()) == ".png" {
				filePath := filepath.Join(path, file.Name())
				q.Push(queue.Task{FilePath: filePath})
				store.SaveImage(filePath, "pending")
				fmt.Println("New image detected:", filePath)
			}
		}

		time.Sleep(2 * time.Second)
	}
}
