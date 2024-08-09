// api.go

package routes

import (
	"CyberDefenseEd/QuadDB/database"
	"CyberDefenseEd/QuadDB/util"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
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
	api.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	{
		api.GET("/docs/:db", func(c *gin.Context) {
			startTime := time.Now()

			dbName := c.Param("db")
			dbFile := filepath.Join(dataDir, dbName+".qdb")
			db := database.LoadDB(dbFile, aesKey)

			page := c.DefaultQuery("page", "1")
			size := c.Query("size")

			pageSize := 5
			if size != "" {
				newSize, err := strconv.Atoi(size)
				if err == nil && newSize > 0 {
					pageSize = newSize
				}
			}

			offset, err := strconv.Atoi(page)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
				return
			}
			if offset <= 0 {
				offset = 1
			}
			offset = (offset - 1) * pageSize

			documents, err := db.LoadDocumentsPaginated(offset, pageSize)
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
				if document.Data == nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Data field is required"})
					return
				}

				err := db.CreateDocument(document.Id, document.Data)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
			}

			databases[dbName] = database.LoadDB(dbFile, aesKey)

			endTime := time.Now()
			elapsedTime := endTime.Sub(startTime)

			c.JSON(http.StatusCreated, gin.H{"_resp": elapsedTime.String(), "message": "Documents created successfully"})
		})

		api.GET("/docs/:db/search", func(c *gin.Context) {
			startTime := time.Now()

			dbName := c.Param("db")
			db := databases[dbName] // Use the loaded database

			// Extract query parameters
			queryParams := c.Request.URL.Query()

			if len(queryParams) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "At least one field-value pair is required"})
				return
			}

			fieldValues := make(map[string]string)
			for field, values := range queryParams {
				if len(values) > 0 {
					fieldValues[field] = values[0] // Only consider the first value for each field
				}
			}

			matchingDocuments, err := db.FetchDocumentsByFieldValues(fieldValues)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			var allDocuments []database.Document
			for key, data := range matchingDocuments {
				document := database.Document{
					Id:   key,
					Data: data,
				}
				allDocuments = append(allDocuments, document)
			}

			elapsedTime := time.Since(startTime)

			c.JSON(http.StatusOK, gin.H{
				"_resp":     elapsedTime.String(),
				"_num":      len(allDocuments),
				"documents": allDocuments,
			})
		})

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

		api.GET("/docs/updates", func(c *gin.Context) {
			adminInfo := gin.H{
				"last_used_db":      database.LastUsedDB,
				"last_update_time":  database.LastUpdateTime.Format(time.RFC3339),
				"last_added_record": database.LastAddedRecord,
				"last_read_record":  database.LastReadRecord,
			}

			c.JSON(http.StatusOK, adminInfo)
		})

		api.GET("/docs/collections", func(c *gin.Context) {
			collections := make(map[string]int)

			for dbName, db := range databases {
				count, err := db.CountDocuments()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				collections[dbName] = count
			}

			c.JSON(http.StatusOK, collections)
		})

		router.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})
	}
}
