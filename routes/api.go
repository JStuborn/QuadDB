// api.go

package routes

import (
	"CyberDefenseEd/QuadDB/util"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
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
	Id   string          `json:"id"`
	Data json.RawMessage `json:"data"`
}

// Database represents a document-based database.
type Database struct {
	documents map[string]json.RawMessage
	mu        sync.RWMutex // Ensure that this line exists
	filename  string       // Filename to store the database
	aesKey    []byte       // AES encryption key
}

// NewDatabase creates a new Database instance.
func NewDatabase(filename string, aesKey []byte) (*Database, error) {
	util.Info(fmt.Sprintf("Found database '%s' attempting to decrypt...", filename))

	db := &Database{
		documents: make(map[string]json.RawMessage),
		filename:  filename,
		aesKey:    aesKey,
	}

	// Load existing documents from file
	err := db.loadDocuments()
	if err != nil {
		util.Error(fmt.Sprintf("Failed to decrypt database '%s': invalid password", db.filename))
		return nil, err
	}

	// Save the documents to file immediately
	err = db.saveDocuments()
	if err != nil {
		return nil, err
	}

	util.Info(fmt.Sprintf("Database '%s' loaded", filename))

	return db, nil
}

func (db *Database) loadDocuments() error {
	data, err := os.ReadFile(db.filename)
	if err != nil {
		// If file does not exist, it's fine, just return
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Decrypt data
	decryptedData, err := db.decrypt(data)
	if err != nil {
		return err
	}

	err = json.Unmarshal(decryptedData, &db.documents)
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

	// Encrypt data
	encryptedData, err := db.encrypt(data)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.filename, encryptedData, 0644)
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

// Encrypt data using AES
func (db *Database) encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(db.aesKey)
	if err != nil {
		return nil, err
	}

	// Pad data if necessary
	data = padData(data, block.BlockSize())

	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], data)

	return ciphertext, nil
}

// Decrypt data using AES
func (db *Database) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(db.aesKey)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// Unpad data
	ciphertext = db.unpadData(ciphertext)

	return ciphertext, nil
}

// Pad data to be a multiple of blockSize
func padData(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func (db *Database) unpadData(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])
	if unpadding > length {
		return nil
	}
	return data[:(length - unpadding)]
}

func SetupRoutes(router *gin.Engine, dataDir string, aesKey []byte) {
	util.Info(fmt.Sprintf("Scanning directory %s", dataDir))

	databases := make(map[string]*Database)

	// Create a new database file if it doesn't exist
	for _, dbName := range []string{"db"} { // Add more database names as needed
		dbFile := filepath.Join(dataDir, dbName+".qdb")
		db, _ := NewDatabase(dbFile, aesKey)
		databases[dbName] = db
	}

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
			for id, data := range db.documents {
				document := Document{
					Id:   id,
					Data: data,
				}
				allDocuments = append(allDocuments, document)
			}

			c.JSON(http.StatusFound, allDocuments)
		})

		api.POST("/documents/:db", func(c *gin.Context) {
			dbName := c.Param("db")
			db, exists := databases[dbName]
			if !exists {
				// Create a new database if it doesn't exist
				dbFile := filepath.Join(dataDir, dbName+".qdb")
				newDB, err := NewDatabase(dbFile, []byte(aesKey))
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				databases[dbName] = newDB
				db = newDB
			}

			var documents []Document
			if err := c.ShouldBindJSON(&documents); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			for _, document := range documents {
				if document.Id == "" || document.Data == nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Key and Data fields are required"})
					return
				}

				err := db.CreateDocument(document.Id, document.Data)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
			}

			c.JSON(http.StatusCreated, gin.H{"message": "Documents created successfully"})
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

			c.JSON(http.StatusFound, gin.H{"data": data})
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
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusAccepted, gin.H{"message": "Document updated successfully"})
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

			c.JSON(http.StatusAccepted, gin.H{"message": "Document deleted successfully"})
		})
	}

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
}
