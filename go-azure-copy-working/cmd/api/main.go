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
// ##########################
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"goaz/internal/config"
	"goaz/services/uploader"

	"github.com/gorilla/mux"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins (change in production)
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
	router := mux.NewRouter()

	// Initialize Azure Client
	azureClient, err := uploader.NewAzureClient(*cfg)
	if err != nil {
		log.Fatalf("Failed to create Azure client: %v", err)
	}

	// Define API route
	router.HandleFunc("/api/blobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		containerId := vars["id"]
		if containerId == "" {
			http.Error(w, "Container ID is required", http.StatusBadRequest)
			return
		}
		blobs, err := uploader.ListBlobsAsJSON(context.Background(), azureClient, *cfg, containerId)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error listing blobs: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(blobs)
	}).Methods("GET", "OPTIONS")

	// router.HandleFunc()

	router.HandleFunc("/api/register/{jetsonId}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		jetsonId := vars["jetsonId"]
		if jetsonId == "" {
			http.Error(w, "Jetson ID is required", http.StatusBadRequest)
			return
		}
		// fmt.Printf("asdasdasd")
		// Create a container with the same name as Jetson ID
		err := uploader.CreateContainer(context.Background(), azureClient, jetsonId)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating containerr - Jetson already registered: %v", err), http.StatusInternalServerError)
			return
		}

		// Return success response
		response := map[string]string{
			"message":     "Container created successfully",
			"jetson":   jetsonId,
			"container": jetsonId,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}).Methods("POST", "OPTIONS")

	port := ":8080"
	log.Println("🚀 API server running on http://localhost" + port)

		// Wrap router with CORS middleware and start the server
		log.Fatal(http.ListenAndServe(port, enableCORS(router)))
	}
// package main

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"log"
// 	"net/http"
// 	"sync"

// 	"goaz/internal/config"
// 	"goaz/services/uploader"

// 	"github.com/gorilla/mux"
// )

// // Path to jetsons.json file
// const jetsonsFile = "/Users/sanjaysirangi/hrf/nextjs-frontend-template/public/jetsons.json"

// // Mutex to handle concurrent file writes
// var fileMutex sync.Mutex

// func enableCORS(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Access-Control-Allow-Origin", "*")
// 		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
// 		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

// 		if r.Method == "OPTIONS" {
// 			w.WriteHeader(http.StatusOK)
// 			return
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }

// // Struct for a Jetson
// type Jetson struct {
// 	Name      string `json:"name"`
// 	Container string `json:"container"`
// 	Config    struct {
// 		Path string `json:"path"`
// 	} `json:"config"`
// }
// type Config struct {
// 	Path string `json:"path"`
// }

// // Read jetsons.json
// func readJetsonsFile() ([]Jetson, error) {
// 	data, err := ioutil.ReadFile(jetsonsFile)
// 	if err != nil {
// 		return nil, fmt.Errorf("error reading Jetsons file: %v", err)
// 	}

// 	var jetsons []Jetson
// 	if err := json.Unmarshal(data, &jetsons); err != nil {
// 		return nil, fmt.Errorf("error parsing jetsons.json: %v", err)
// 	}

// 	return jetsons, nil
// }

// // Write updated Jetsons list to jetsons.json
// func writeJetsonsFile(jetsons []Jetson) error {
// 	fileMutex.Lock()
// 	defer fileMutex.Unlock()

// 	jsonData, err := json.MarshalIndent(jetsons, "", "  ")
// 	if err != nil {
// 		return err
// 	}

// 	return ioutil.WriteFile(jetsonsFile, jsonData, 0644)
// }

// func main() {
// 	cfg := config.LoadConfig()
// 	router := mux.NewRouter()

// 	azureClient, err := uploader.NewAzureClient(*cfg)
// 	if err != nil {
// 		log.Fatalf("Failed to create Azure client: %v", err)
// 	}

// 	// err2 := uploader.DeleteAllBlobsFromContainer(context.Background(), azureClient, "storage-container2")
// 	// if err2 != nil {
// 	// 	log.Fatalf("Error deleting blobs: %v", err)
// 	// }

// 	// API to list blobs
// 	router.HandleFunc("/api/blobs/{id}", func(w http.ResponseWriter, r *http.Request) {
// 		vars := mux.Vars(r)
// 		containerId := vars["id"]
// 		if containerId == "" {
// 			http.Error(w, "Container ID is required", http.StatusBadRequest)
// 			return
// 		}
// 		blobs, err := uploader.ListBlobsAsJSON(context.Background(), azureClient, *cfg, containerId)
// 		if err != nil {
// 			http.Error(w, fmt.Sprintf("Error listing blobs: %v", err), http.StatusInternalServerError)
// 			return
// 		}

// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusOK)
// 		w.Write(blobs)
// 	}).Methods("GET", "OPTIONS")

	// Register Jetson API
	// router.HandleFunc("/api/register/{jetsonId}", func(w http.ResponseWriter, r *http.Request) {
	// 	vars := mux.Vars(r)
	// 	jetsonId := vars["jetsonId"]
	// 	if jetsonId == "" {
	// 		http.Error(w, "Jetson ID is required", http.StatusBadRequest)
	// 		return
	// 	}

	// 	// Create Azure container for Jetson
	// 	err := uploader.CreateContainer(context.Background(), azureClient, jetsonId)
	// 	if err != nil {
	// 		http.Error(w, fmt.Sprintf("Error creating container - Jetson might already be registered: %v", err), http.StatusInternalServerError)
	// 		return
	// 	}

	// 	// Read and update the jetsons.json file
	// 	jetsons, err := readJetsonsFile()
	// 	if err != nil {
	// 		http.Error(w, fmt.Sprintf("Error reading Jetsons file: %v", err), http.StatusInternalServerError)
	// 		return
	// 	}

	// 	// Update the existing Jetson with the new container name
	// 	var jetsonPath string
	// 	for i, jetson := range jetsons {
	// 		if jetson.Name == jetsonId {
	// 			jetsons[i].Container = jetsonId
	// 			jetsonPath = jetson.Config.Path // Extracting the path
	// 			break
	// 		}
	// 	}

	// 	// Save updated jetsons.json
	// 	if err := writeJetsonsFile(jetsons); err != nil {
	// 		http.Error(w, fmt.Sprintf("Error updating Jetsons file: %v", err), http.StatusInternalServerError)
	// 		return
	// 	}

	// 	fmt.Println(jetsonPath)

	// 	if jetsonPath != "" {
	// 		config.UpdateImageSource(jetsonPath)
	// 	}

	// 	response := map[string]string{
	// 		"message":   "Container created successfully",
	// 		"jetson":    jetsonId,
	// 		"container": jetsonId,
	// 	}

	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.WriteHeader(http.StatusOK)
	// 	json.NewEncoder(w).Encode(response)
	// }).Methods("POST", "OPTIONS")
// 	router.HandleFunc("/api/register/{jetsonId}", func(w http.ResponseWriter, r *http.Request) {
// 		vars := mux.Vars(r)
// 		jetsonId := vars["jetsonId"]
// 		if jetsonId == "" {
// 			http.Error(w, "Jetson ID is required", http.StatusBadRequest)
// 			return
// 		}

// 		// Create Azure container for Jetson
// 		err := uploader.CreateContainer(context.Background(), azureClient, jetsonId)
// 		if err != nil {
// 			http.Error(w, fmt.Sprintf("Error creating container - Jetson might already be registered: %v", err), http.StatusInternalServerError)
// 			return
// 		}

// 		// Read and update the jetsons.json file
// 		jetsons, err := readJetsonsFile()
// 		if err != nil {
// 			http.Error(w, fmt.Sprintf("Error reading Jetsons file: %v", err), http.StatusInternalServerError)
// 			return
// 		}

// 		// Extract path from jetsons.json
// 		var jetsonPath string
// 		for i, jetson := range jetsons {
// 			if jetson.Name == jetsonId {
// 				jetsons[i].Container = jetsonId
// 				jetsonPath = jetson.Config.Path
// 				break
// 			}
// 		}

// 		// Save updated jetsons.json
// 		if err := writeJetsonsFile(jetsons); err != nil {
// 			http.Error(w, fmt.Sprintf("Error updating Jetsons file: %v", err), http.StatusInternalServerError)
// 			return
// 		}

// 		// Update ImageSource in image_sources.json
// 		// if jetsonPath != "" {
// 		// 	fmt.Println("Adding path to ImageSource:", jetsonPath)
// 		// 	config.UpdateImageSource(jetsonPath)
// 		// }

// 		response := map[string]string{
// 			"message":   "Container created successfully",
// 			"jetson":    jetsonId,
// 			"container": jetsonId,
// 		}

// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusOK)
// 		json.NewEncoder(w).Encode(response)
// 	}).Methods("POST", "OPTIONS")

// 	// Create Jetson API
// 	router.HandleFunc("/api/create-jetson", func(w http.ResponseWriter, r *http.Request) {
// 		jetsons, err := readJetsonsFile()
// 		if err != nil {
// 			http.Error(w, fmt.Sprintf("Error reading Jetsons file: %v", err), http.StatusInternalServerError)
// 			return
// 		}

// 		newJetsonName := fmt.Sprintf("finaltestjetson%d", len(jetsons)+1)
// 		newJetson := Jetson{
// 			Name:      newJetsonName,
// 			Container: "Not registered",
// 			Config: Config{
// 				Path: fmt.Sprintf("/Users/sanjaysirangi/Desktop/go-azure-copy/%s", newJetsonName),
// 			},
// 		}
// 		// newJetson.Config.Path = "/Users/sanjaysirangi/Desktop/go-azure-copy/files"

// 		jetsons = append(jetsons, newJetson)

// 		if err := writeJetsonsFile(jetsons); err != nil {
// 			http.Error(w, fmt.Sprintf("Error updating Jetsons file: %v", err), http.StatusInternalServerError)
// 			return
// 		}

// 		response := map[string]string{
// 			"message": "Jetson created successfully",
// 			"jetson":  newJetsonName,
// 		}

// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusOK)
// 		json.NewEncoder(w).Encode(response)
// 	}).Methods("POST", "OPTIONS")

// 	port := ":8080"
// 	log.Println("🚀 API server running on http://localhost" + port)
// 	log.Fatal(http.ListenAndServe(port, enableCORS(router)))
// }
