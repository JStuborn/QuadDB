package routes

import (
	"net/http"
	"text/template"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func RenderTemplate(w http.ResponseWriter, templateName string, title string, data interface{}) {
	tmpl, err := template.ParseFiles(
		"./dashboard/templates/base.html",
		"./dashboard/templates/pages/"+templateName,
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, map[string]interface{}{
		"title": title,
		"data":  data,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func SetupDashboardRoutes(router *gin.Engine, dataDir string, aesKey []byte) {
	// die
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "http://localhost:9010"
		},
		MaxAge: 12 * time.Hour,
	}))

	router.GET("/", func(c *gin.Context) {
		RenderTemplate(c.Writer, "home.html", "Dashboard", gin.H{})
	})
}
