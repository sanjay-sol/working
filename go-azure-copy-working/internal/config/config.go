// //

// package config

// // Config stores application configuration
// type Config struct {
// 	AzureAccountName  string
// 	AzureAccountKey   string
// 	Container         string
// 	RetryPath         string
// 	ImageSource       []string
// 	RetryInterval     int
// 	QueueSize         int
// 	MaxWorkers        int
// 	TotalRetryWorkers int
// 	LogLevel          string
// 	RetryLimit        int
// }

// // LoadConfig loads configuration from .env file or environment variables
// func LoadConfig() *Config {
// 	// if err := godotenv.Load(); err != nil {
// 	// 	log.Println("No .env file found, loading from environment variables")
// 	// }

// 	return &Config{
// 		AzureAccountName: "harvestedstorage2",
// 		AzureAccountKey:  "bihW5fxPa/VdaATbn5iBgj+yd6XBmn6LQaXEjgHiThbJ3RcW+M6TtQc5Ml3cfihXruNRQRjzYGpU+AStU/OnHA==",
// 		Container:        "storage-container2",
// 		RetryPath:        "/Users/sanjaysirangi/Desktop/go-azure-copy/retry",
// 		ImageSource: []string{
// 			"/Users/sanjaysirangi/Desktop/go-azure-copy/files",
// 			"/Users/sanjaysirangi/Desktop/go-azure-copy/files2",
// 			"/Users/sanjaysirangi/Desktop/go-azure-copy/files3",
// 			"/Users/sanjaysirangi/Desktop/go-azure-copy/files4",
// 		},
// 		RetryInterval:     10,
// 		QueueSize:         500,
// 		MaxWorkers:        100,
// 		TotalRetryWorkers: 10,
// 		LogLevel:          "info",
// 		RetryLimit:        3,
// 	}
// }

// func UpdateImageSource(newPath string) {
// 	cfg := LoadConfig()

// 	// Check if the path already exists to prevent duplicates
// 	for _, path := range cfg.ImageSource {
// 		if path == newPath {
// 			return
// 		}
// 	}

// 	cfg.ImageSource = append(cfg.ImageSource, newPath)
// }

// package config

// import "sync"

// var configInstance *Config
// var once sync.Once

// type Config struct {
// 	AzureAccountName  string
// 	AzureAccountKey   string
// 	Container         string
// 	RetryPath         string
// 	ImageSource       []string
// 	RetryInterval     int
// 	QueueSize         int
// 	MaxWorkers        int
// 	TotalRetryWorkers int
// 	LogLevel          string
// 	RetryLimit        int
// }

// // LoadConfig initializes and returns a singleton Config instance
// func LoadConfig() *Config {
// 	once.Do(func() {
// 		configInstance = &Config{
// 			AzureAccountName: "harvestedstorage2",
// 			AzureAccountKey:  "bihW5fxPa/VdaATbn5iBgj+yd6XBmn6LQaXEjgHiThbJ3RcW+M6TtQc5Ml3cfihXruNRQRjzYGpU+AStU/OnHA==",
// 			Container:        "storage-container2",
// 			RetryPath:        "/Users/sanjaysirangi/Desktop/go-azure-copy/retry",
// 			ImageSource: []string{
// 				"/Users/sanjaysirangi/Desktop/go-azure-copy/files",
// 				"/Users/sanjaysirangi/Desktop/go-azure-copy/files2",
// 				"/Users/sanjaysirangi/Desktop/go-azure-copy/files3",
// 				"/Users/sanjaysirangi/Desktop/go-azure-copy/files4",
// 			},
// 			RetryInterval:     10,
// 			QueueSize:         500,
// 			MaxWorkers:        100,
// 			TotalRetryWorkers: 10,
// 			LogLevel:          "info",
// 			RetryLimit:        3,
// 		}
// 	})
// 	return configInstance
// }

// // updateImageSource dynamically adds a new path to the ImageSource array
// func UpdateImageSource(newPath string) {
// 	cfg := LoadConfig()

// 	// Check if the path already exists to prevent duplicates
// 	for _, path := range cfg.ImageSource {
// 		if path == newPath {
// 			return
// 		}
// 	}

// 	cfg.ImageSource = append(cfg.ImageSource, newPath)
// }

package config

import (
	"os"
	"strconv"
)

type Config struct {
	AzureAccountName  string `json:"azure_account_name"`
	AzureAccountKey   string `json:"azure_account_key"`
	Container         string `json:"container"`
	RetryPath         string `json:"retry_path"`
	ImageSource       string `json:"image_source"`
	RetryInterval     int    `json:"retry_interval"`
	QueueSize         int    `json:"queue_size"`
	MaxWorkers        int    `json:"max_workers"`
	TotalRetryWorkers int    `json:"total_retry_workers"`
	LogLevel          string `json:"log_level"`
	RetryLimit        int    `json:"retry_limit"`
}

func getEnvOrDefault(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getEnvIntOrDefault(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func LoadConfig() *Config {
	return &Config{
		AzureAccountName:  getEnvOrDefault("AZURE_ACCOUNT_NAME", "harvestedstorage2"),
		AzureAccountKey:   getEnvOrDefault("AZURE_ACCOUNT_KEY", "bihW5fxPa/VdaATbn5iBgj+yd6XBmn6LQaXEjgHiThbJ3RcW+M6TtQc5Ml3cfihXruNRQRjzYGpU+AStU/OnHA=="),
		Container:         getEnvOrDefault("AZURE_CONTAINER", "storage-container"),
		RetryPath:         getEnvOrDefault("RETRY_PATH", "/tmp/retry"),
		ImageSource:       getEnvOrDefault("IMAGE_SOURCE", "/Users/sanjaysirangi/Desktop/mimic/images"),
		RetryInterval:     getEnvIntOrDefault("RETRY_INTERVAL", 10),
		QueueSize:         getEnvIntOrDefault("QUEUE_SIZE", 500),
		MaxWorkers:        getEnvIntOrDefault("MAX_WORKERS", 100),
		TotalRetryWorkers: getEnvIntOrDefault("TOTAL_RETRY_WORKERS", 10),
		LogLevel:          getEnvOrDefault("LOG_LEVEL", "info"),
		RetryLimit:        getEnvIntOrDefault("RETRY_LIMIT", 3),
	}
}


// // Load image sources from `image_sources.json`
// func loadImageSources() []string {
// 	var imgConfig ImageSourceConfig

// 	file, err := os.Open(imageSourceFile)
// 	if err != nil {
// 		fmt.Println("No image_sources.json found, using defaults...")
// 		return []string{}
// 	}
// 	defer file.Close()

// 	if err := json.NewDecoder(file).Decode(&imgConfig); err != nil {
// 		fmt.Println("Error reading image_sources.json:", err)
// 		return []string{}
// 	}

// 	return imgConfig.Paths
// }

// // Append new path to `image_sources.json` and reload config
// func UpdateImageSource(newPath string) {
// 	// Load current paths
// 	paths := loadImageSources()

// 	// Check if the path already exists
// 	for _, path := range paths {
// 		if path == newPath {
// 			fmt.Println("Path already exists:", newPath)
// 			return
// 		}
// 	}

// 	// Append and save
// 	paths = append(paths, newPath)
// 	saveImageSources(paths)

// 	// Create the directory if it doesn't exist
// 	err := os.MkdirAll(newPath, os.ModePerm)
// 	if err != nil {
// 		fmt.Println("Error creating directory:", err)
// 		return
// 	}
// 	fmt.Println("Directory created at:", newPath)

// 	// Reload ImageSource in config
// 	configInstance.ImageSource = paths
// 	fmt.Println("Updated ImageSource:", configInstance.ImageSource)
// }

// // Save updated image paths back to `image_sources.json`
// func saveImageSources(paths []string) {
// 	file, err := os.Create(imageSourceFile)
// 	if err != nil {
// 		fmt.Println("Error saving image sources:", err)
// 		return
// 	}
// 	defer file.Close()

// 	data := ImageSourceConfig{Paths: paths}
// 	encoder := json.NewEncoder(file)
// 	encoder.SetIndent("", "  ")
// 	if err := encoder.Encode(data); err != nil {
// 		fmt.Println("Error encoding JSON:", err)
// 	}
// }
