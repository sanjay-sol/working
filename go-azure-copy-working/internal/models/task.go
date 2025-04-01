// package models

// type ImageTask struct {
// 	ImagePath         string
// 	BlobName          string
// 	Retry             int
// }

package models

import "time"

// Task represents an image upload task
type ImageTask struct {
	ID         string    `json:"id"`
	FilePath   string    `json:"file_path"`
	BlobName   string    `json:"blob_name"`
	RetryCount int       `json:"retry_count"`
	Timestamp  time.Time `json:"timestamp"`
}
