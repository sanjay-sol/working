package storage

// import (
// 	"database/sql"
// 	"os"
// 	"path/filepath"

// 	_ "github.com/mattn/go-sqlite3"
// )

// // Store manages image storage and indexing.
// type Store struct {
// 	db       *sql.DB
// 	basePath string
// }

// // NewStore initializes storage with SQLite.
// func NewStore(dbPath, basePath string) (*Store, error) {
// 	db, err := sql.Open("sqlite3", dbPath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS images (
// 		id INTEGER PRIMARY KEY AUTOINCREMENT,
// 		file_path TEXT UNIQUE,
// 		status TEXT
// 	)`)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &Store{db: db, basePath: basePath}, nil
// }

// // SaveImage stores image metadata.
// func (s *Store) SaveImage(filePath string, status string) error {
// 	_, err := s.db.Exec("INSERT OR IGNORE INTO images (file_path, status) VALUES (?, ?)", filePath, status)
// 	return err
// }

// // GetPendingImages retrieves unuploaded images.
// func (s *Store) GetPendingImages() ([]string, error) {
// 	rows, err := s.db.Query("SELECT file_path FROM images WHERE status = 'pending'")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var paths []string
// 	for rows.Next() {
// 		var path string
// 		if err := rows.Scan(&path); err != nil {
// 			return nil, err
// 		}
// 		paths = append(paths, path)
// 	}
// 	return paths, nil
// }

// // MarkUploaded updates image status after upload.
// func (s *Store) MarkUploaded(filePath string) error {
// 	_, err := s.db.Exec("UPDATE images SET status = 'uploaded' WHERE file_path = ?", filePath)
// 	return err
// }

// // SaveFile physically saves an image.
// func (s *Store) SaveFile(fileName string, data []byte) (string, error) {
// 	filePath := filepath.Join(s.basePath, fileName)
// 	err := os.WriteFile(filePath, data, 0644)
// 	if err != nil {
// 		return "", err
// 	}
// 	return filePath, nil
// }
