// package main

// import (
// 	"context"
// 	"goaz/internal/config"
// 	"goaz/services/uploader"
// 	"log"
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

// 	// err2 := uploader.ListBlobs(ctx, client)

// 	// if err2 != nil {
// 	// 	log.Fatalf("❌ Failed to list blobs: %v", err2)
// 	// }

// 	err3 := uploader.DownloadBlobs(ctx, client, "blobs")
// 	if err3 != nil {
// 		log.Fatalf("Error downloading blobs: %v", err)
// 	}

// }

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"goaz/internal/config"
	"goaz/services/uploader"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins (change "*" to a specific domain if needed)
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests (OPTIONS method)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	cfg := config.LoadConfig()

	// Initialize Azure Client
	azureClient, err := uploader.NewAzureClient(*cfg)
	if err != nil {
		log.Fatalf("Failed to create Azure client: %v", err)
	}

	// err2 := uploader.DeleteAllBlobsFromContainer(context.Background(), azureClient)
	// if err2 != nil {
	// 	log.Fatalf("Error deleting blobs: %v", err)
	// }

	// Define API route
	http.HandleFunc("/api/blobs", func(w http.ResponseWriter, r *http.Request) {
		blobs, err := uploader.ListBlobsAsJSON(context.Background(), azureClient, *cfg)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error listing blobs: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(blobs)
	})

	port := ":8080"
	log.Println("🚀 API server running on http://localhost" + port)

	// Wrap server with CORS middleware
	log.Fatal(http.ListenAndServe(port, enableCORS(http.DefaultServeMux)))
}
