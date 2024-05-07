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
	"strings"
	"sync"
	"time"

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
	db := &Database{
		documents: make(map[string]json.RawMessage),
		filename:  filename,
		aesKey:    aesKey,
	}

	err := db.loadDocuments()
	if err != nil {
		util.Error(fmt.Sprintf("Failed to decrypt database '%s': invalid password", db.filename))
		return nil, err
	}

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
		// If file does not exist, it's fine (i hope), just return
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

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

func (db *Database) saveDocuments() error {
	data, err := json.Marshal(db.documents)
	if err != nil {
		return err
	}

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

func (db *Database) CreateDocument(key string, data json.RawMessage) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.documents[key]; exists {
		return fmt.Errorf("document with key '%s' already exists", key)
	}

	db.documents[key] = data

	err := db.saveDocuments()
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) ReadDocument(key string) (json.RawMessage, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	data, exists := db.documents[key]
	if !exists {
		return nil, fmt.Errorf("document with key '%s' not found", key)
	}

	return data, nil
}

func (db *Database) UpdateDocument(key string, data json.RawMessage) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.documents[key]; !exists {
		return fmt.Errorf("document with key '%s' not found", key)
	}

	db.documents[key] = data

	err := db.saveDocuments()
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) DeleteDocument(key string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.documents[key]; !exists {
		return fmt.Errorf("document with key '%s' not found", key)
	}

	delete(db.documents, key)

	err := db.saveDocuments()
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(db.aesKey)
	if err != nil {
		return nil, err
	}

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

	// Unpad data and hope it doesnt break
	ciphertext = db.unpadData(ciphertext)

	return ciphertext, nil
}

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
	databases := make(map[string]*Database)

	dbFiles, err := filepath.Glob(filepath.Join(dataDir, "*.qdb"))
	if err != nil {
		util.Error("Failed to index qdb files.")
	}

	for _, dbFile := range dbFiles {
		dbName := strings.TrimSuffix(filepath.Base(dbFile), ".qdb")
		db, err := NewDatabase(dbFile, aesKey)
		if err != nil {
			util.Error(fmt.Sprintf("Failed to load database '%s'", dbFile))
		}
		databases[dbName] = db
	}

	api := router.Group("/api")
	{
		api.GET("/documents/:db", func(c *gin.Context) {
			dbName := c.Param("db")
			db, exists := databases[dbName]
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Database not found"})
				return
			}

			startTime := time.Now()

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

			endTime := time.Now()
			elapsedTime := endTime.Sub(startTime)

			numDocs := len(allDocuments)

			c.JSON(http.StatusOK, gin.H{
				"_resp":     elapsedTime.String(),
				"_num":      numDocs,
				"documents": allDocuments,
			})
		})

		api.POST("/documents/:db", func(c *gin.Context) {
			dbName := c.Param("db")
			db, exists := databases[dbName]
			if !exists {
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
