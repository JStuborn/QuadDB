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
	"time"

	"github.com/gin-gonic/gin"
)

type Document struct {
	Id   string          `json:"id"`
	Data json.RawMessage `json:"data"`
}

type Database struct {
	filename string
	aesKey   []byte
}

var (
	lastUsedDB      string
	lastUpdateTime  time.Time
	lastAddedRecord string
	lastReadRecord  json.RawMessage
)

func LoadDB(filename string, aesKey []byte) *Database {
	return &Database{
		filename: filename,
		aesKey:   aesKey,
	}
}

func (db *Database) loadDocuments() (map[string]json.RawMessage, error) {
	data, err := os.ReadFile(db.filename)
	if err != nil {
		// If file does not exist, return an empty map
		if os.IsNotExist(err) {
			return make(map[string]json.RawMessage), nil
		}
		return nil, err
	}

	decryptedData, err := db.decrypt(data)
	if err != nil {
		return nil, err
	}

	var documents map[string]json.RawMessage
	err = json.Unmarshal(decryptedData, &documents)
	if err != nil {
		return nil, err
	}

	return documents, nil
}

func (db *Database) saveDocuments(documents map[string]json.RawMessage) error {
	data, err := json.Marshal(documents)
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
	documents, err := db.loadDocuments()
	if err != nil {
		return err
	}

	if _, exists := documents[key]; exists {
		return fmt.Errorf("document with key '%s' already exists", key)
	}

	documents[key] = data

	// Update last added record
	lastAddedRecord = key

	err = db.saveDocuments(documents)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) ReadDocument(key string) (json.RawMessage, error) {
	documents, err := db.loadDocuments()
	if err != nil {
		return nil, err
	}

	data, exists := documents[key]
	if !exists {
		return nil, fmt.Errorf("document with key '%s' not found", key)
	}

	// Update last used database and last update time
	lastUsedDB = db.filename
	lastUpdateTime = time.Now()

	return data, nil
}

func (db *Database) UpdateDocument(key string, data json.RawMessage) error {
	documents, err := db.loadDocuments()
	if err != nil {
		return err
	}

	if _, exists := documents[key]; !exists {
		return fmt.Errorf("document with key '%s' not found", key)
	}

	documents[key] = data

	// Update last update time
	lastUpdateTime = time.Now()

	err = db.saveDocuments(documents)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) DeleteDocument(key string) error {
	documents, err := db.loadDocuments()
	if err != nil {
		return err
	}

	if _, exists := documents[key]; !exists {
		return fmt.Errorf("document with key '%s' not found", key)
	}

	delete(documents, key)

	err = db.saveDocuments(documents)
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

	// Unpad data
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
		db := LoadDB(dbFile, aesKey)
		databases[dbName] = db
	}

	api := router.Group("/api")
	{

		// @Summary Get documents by database
		// @Description Retrieve documents from a specific database
		// @ID get-documents
		// @Produce json
		// @Param db path string true "Database name"
		// @Success 200 {object} gin.H
		// @Router /api/documents/{db} [get]
		api.GET("/docs/:db", func(c *gin.Context) {
			startTime := time.Now()

			dbName := c.Param("db")
			dbFile := filepath.Join(dataDir, dbName+".qdb")
			db := LoadDB(dbFile, aesKey)

			documents, err := db.loadDocuments()
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Database not found"})
				return
			}

			var allDocuments []Document
			for key, data := range documents {
				document := Document{
					Id:   key,
					Data: data,
				}
				allDocuments = append(allDocuments, document)
			}

			endTime := time.Now()
			elapsedTime := endTime.Sub(startTime)

			c.JSON(http.StatusOK, gin.H{"_resp": elapsedTime.String(), "_num": len(allDocuments), "documents": allDocuments})
		})

		// @Summary Create documents
		// @Description Create documents in a specific database
		// @ID create-documents
		// @Accept json
		// @Produce json
		// @Param db path string true "Database name"
		// @Success 201 {object} gin.H
		// @Router /api/docs/{db} [post]
		api.POST("/docs/:db", func(c *gin.Context) {
			startTime := time.Now()

			dbName := c.Param("db")
			dbFile := filepath.Join(dataDir, dbName+".qdb")
			db := LoadDB(dbFile, aesKey)

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

			endTime := time.Now()
			elapsedTime := endTime.Sub(startTime)

			c.JSON(http.StatusCreated, gin.H{"_resp": elapsedTime.String(), "message": "Documents created successfully"})
		})

		// @Summary Get document by key
		// @Description Retrieve a document by its key from a specific database
		// @ID get-document-by-key
		// @Produce json
		// @Param db path string true "Database name"
		// @Param key path string true "Document key"
		// @Success 200 {object} gin.H
		// @Router /api/docs/{db}/{key} [get]
		api.GET("/docs/:db/:key", func(c *gin.Context) {
			startTime := time.Now()

			dbName := c.Param("db")
			dbFile := filepath.Join(dataDir, dbName+".qdb")
			db := LoadDB(dbFile, aesKey)

			key := c.Param("key")
			data, err := db.ReadDocument(key)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}

			lastReadRecord = data

			endTime := time.Now()
			elapsedTime := endTime.Sub(startTime)

			c.JSON(http.StatusOK, gin.H{"_resp": elapsedTime.String(), "data": data})
		})

		// @Summary Update document by key
		// @Description Update a document by its key in a specific database
		// @ID update-document-by-key
		// @Accept json
		// @Produce json
		// @Param db path string true "Database name"
		// @Param key path string true "Document key"
		// @Success 200 {object} gin.H
		// @Router /api/docs/{db}/{key} [put]
		api.PUT("/docs/:db/:key", func(c *gin.Context) {
			startTime := time.Now()

			dbName := c.Param("db")
			dbFile := filepath.Join(dataDir, dbName+".qdb")
			db := LoadDB(dbFile, aesKey)

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

			endTime := time.Now()
			elapsedTime := endTime.Sub(startTime)

			c.JSON(http.StatusOK, gin.H{"_resp": elapsedTime.String(), "message": "Document updated successfully"})
		})

		// @Summary Delete document by key
		// @Description Delete a document by its key from a specific database
		// @ID delete-document-by-key
		// @Produce json
		// @Param db path string true "Database name"
		// @Param key path string true "Document key"
		// @Success 200 {object} gin.H
		// @Router /api/docs/{db}/{key} [delete]
		api.DELETE("/docs/:db/:key", func(c *gin.Context) {
			startTime := time.Now()

			dbName := c.Param("db")
			dbFile := filepath.Join(dataDir, dbName+".qdb")
			db := LoadDB(dbFile, aesKey)

			key := c.Param("key")
			err := db.DeleteDocument(key)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			endTime := time.Now()
			elapsedTime := endTime.Sub(startTime)

			c.JSON(http.StatusOK, gin.H{"_resp": elapsedTime.String(), "message": "Document deleted successfully"})
		})

		// @Summary Get admin information
		// @Description Retrieve information about the last used database, last update time, last added record, and last read record
		// @ID get-admin-info
		// @Produce json
		// @Success 200 {object} gin.H
		// @Router /api/docs/updates [get]

		// all stored in memory non-persistant though restarts for security. not encrypted
		api.GET("/docs/updates", func(c *gin.Context) {
			lastUsedDB := lastUsedDB
			lastUpdateTime := lastUpdateTime.Format(time.RFC3339)
			lastAddedRecord := lastAddedRecord

			adminInfo := gin.H{
				"last_used_db":      lastUsedDB,
				"last_update_time":  lastUpdateTime,
				"last_added_record": lastAddedRecord,
				"last_read_record":  lastReadRecord,
			}

			c.JSON(http.StatusOK, adminInfo)
		})
	}

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
}
