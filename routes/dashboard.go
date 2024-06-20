package routes

import (
	"encoding/json"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var users map[string]string

func init() {
	users = make(map[string]string)
	file, err := os.Open("./config/users.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&users); err != nil {
		panic(err)
	}
}

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

func authMiddleware(c *gin.Context) {
	// if _, exists := c.Get("authenticatedUser"); !exists {
	// 	c.Redirect(http.StatusFound, "/login")
	// 	c.Abort()
	// 	return
	// }
	c.Next()
}

func loginHandler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		RenderTemplate(c.Writer, "login.html", "Login", nil)
		return
	}

	var creds struct {
		Username string `form:"username" json:"username" binding:"required"`
		Password string `form:"password" json:"password" binding:"required"`
	}

	if err := c.ShouldBind(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	hashedPassword, exists := users[creds.Username]
	if !exists || bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(creds.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	c.Set("authenticatedUser", creds.Username)
	c.Redirect(http.StatusFound, "/")
}

func SetupDashboardRoutes(router *gin.Engine, dataDir string, aesKey []byte) {
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

	router.GET("/login", loginHandler)
	router.POST("/login", loginHandler)

	router.GET("/", authMiddleware, func(c *gin.Context) {
		RenderTemplate(c.Writer, "home.html", "Dashboard", gin.H{})
	})
}
