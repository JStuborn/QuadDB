package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"
)

// Document represents a document in the database.
type Document struct {
	Key  string          `json:"key"`
	Data json.RawMessage `json:"data"`
}

// Database represents a document-based database.
type Database struct {
	documents map[string]json.RawMessage
	mu        sync.RWMutex
	filename  string // Filename to store the database
}

// NewDatabase creates a new instance of a Database.
func NewDatabase(filename string) (*Database, error) {
	db := &Database{
		documents: make(map[string]json.RawMessage),
		filename:  filename,
	}

	// Load existing documents from file
	err := db.loadDocuments()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// loadDocuments loads existing documents from the database file.
func (db *Database) loadDocuments() error {
	data, err := os.ReadFile(db.filename)
	if err != nil {
		// If file does not exist, it's fine, just return
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	err = json.Unmarshal(data, &db.documents)
	if err != nil {
		return err
	}

	return nil
}

// saveDocuments saves all documents to the database file.
func (db *Database) saveDocuments() error {
	data, err := json.Marshal(db.documents)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.filename, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// CreateDocument adds a new document to the database.
func (db *Database) CreateDocument(key string, data json.RawMessage) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.documents[key]; exists {
		return fmt.Errorf("document with key '%s' already exists", key)
	}

	db.documents[key] = data

	// Save documents to file
	err := db.saveDocuments()
	if err != nil {
		return err
	}

	return nil
}

// ReadDocument retrieves a document from the database.
func (db *Database) ReadDocument(key string) (json.RawMessage, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	data, exists := db.documents[key]
	if !exists {
		return nil, fmt.Errorf("document with key '%s' not found", key)
	}

	return data, nil
}

// UpdateDocument updates an existing document in the database.
func (db *Database) UpdateDocument(key string, data json.RawMessage) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.documents[key]; !exists {
		return fmt.Errorf("document with key '%s' not found", key)
	}

	db.documents[key] = data

	// Save documents to file
	err := db.saveDocuments()
	if err != nil {
		return err
	}

	return nil
}

// DeleteDocument deletes a document from the database.
func (db *Database) DeleteDocument(key string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.documents[key]; !exists {
		return fmt.Errorf("document with key '%s' not found", key)
	}

	delete(db.documents, key)

	// Save documents to file
	err := db.saveDocuments()
	if err != nil {
		return err
	}

	return nil
}

// Compress data using gzip
func compress(data json.RawMessage) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := gzip.NewWriter(buf)
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	writer.Close()
	return buf.Bytes(), nil
}

// Decompress data using gzip
func decompress(compressedData []byte) (json.RawMessage, error) {
	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	decompressedData, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(decompressedData), nil
}

func main() {
	// Data directory
	dataDir := "data"
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		panic(err)
	}

	// Initialize databases
	databases := make(map[string]*Database)

	// Initialize Gin router
	router := gin.Default()

	// REST API endpoints
	api := router.Group("/api")
	{
		api.GET("/documents/:db", func(c *gin.Context) {
			dbName := c.Param("db")
			db, exists := databases[dbName]
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Database not found"})
				return
			}

			db.mu.RLock()
			defer db.mu.RUnlock()

			var allDocuments []Document
			for key, data := range db.documents {
				document := Document{
					Key:  key,
					Data: data,
				}
				allDocuments = append(allDocuments, document)
			}

			c.JSON(http.StatusOK, allDocuments)
		})

		api.POST("/documents/:db", func(c *gin.Context) {
			dbName := c.Param("db")
			db, exists := databases[dbName]
			if !exists {
				// Create a new database if it doesn't exist
				dbFile := filepath.Join(dataDir, dbName+".qdb")
				db, err = NewDatabase(dbFile)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				databases[dbName] = db
			}

			var document Document
			if err := c.ShouldBindJSON(&document); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if document.Key == "" || document.Data == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Key and Data fields are required"})
				return
			}

			err := db.CreateDocument(document.Key, document.Data)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, gin.H{"message": "Document created successfully"})
		})

		api.GET("/documents/:db/:key", func(c *gin.Context) {
			dbName := c.Param("db")
			db, exists := databases[dbName]
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Database not found"})
				return
			}

			key := c.Param("key")
			data, err := db.ReadDocument(key)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"data": data})
		})

		api.PUT("/documents/:db/:key", func(c *gin.Context) {
			dbName := c.Param("db")
			db, exists := databases[dbName]
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Database not found"})
				return
			}

			key := c.Param("key")
			var newData json.RawMessage
			if err := c.ShouldBindJSON(&newData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			err := db.UpdateDocument(key, newData)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Document updated successfully"})
		})

		api.DELETE("/documents/:db/:key", func(c *gin.Context) {
			dbName := c.Param("db")
			db, exists := databases[dbName]
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Database not found"})
				return
			}

			key := c.Param("key")
			err := db.DeleteDocument(key)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Document deleted successfully"})
		})
	}

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Start server
	router.Run(":8080")
}
