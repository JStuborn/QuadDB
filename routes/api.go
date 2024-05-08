// api.go

package routes

import (
	"CyberDefenseEd/QuadDB/database"
	"CyberDefenseEd/QuadDB/util"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, dataDir string, aesKey []byte) {
	databases := make(map[string]*database.Database)

	dbFiles, err := filepath.Glob(filepath.Join(dataDir, "*.qdb"))
	if err != nil {
		util.Error("Failed to index qdb files.")
	}

	for _, dbFile := range dbFiles {
		dbName := strings.TrimSuffix(filepath.Base(dbFile), ".qdb")
		db := database.LoadDB(dbFile, aesKey)
		databases[dbName] = db
		util.Info(fmt.Sprintf("Imported Database - %s.qdb", dbName))
	}

	api := router.Group("/api/v1")
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
			db := database.LoadDB(dbFile, aesKey)

			documents, err := db.LoadDocuments()
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Database not found"})
				return
			}

			var allDocuments []database.Document
			for key, data := range documents {
				document := database.Document{
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
			db := database.LoadDB(dbFile, aesKey)

			var documents []database.Document
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
			db := database.LoadDB(dbFile, aesKey)

			key := c.Param("key")
			data, err := db.ReadDocument(key)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}

			database.LastReadRecord = data

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
			db := database.LoadDB(dbFile, aesKey)

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
			db := database.LoadDB(dbFile, aesKey)

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

			adminInfo := gin.H{
				"last_used_db":      database.LastUsedDB,
				"last_update_time":  database.LastUpdateTime.Format(time.RFC3339),
				"last_added_record": database.LastAddedRecord,
				"last_read_record":  database.LastReadRecord,
			}

			c.JSON(http.StatusOK, adminInfo)
		})
	}

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
}
