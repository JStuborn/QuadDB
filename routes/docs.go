package routes

import (
	docs "CyberDefenseEd/QuadDB/docs"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterSwaggerRoutes(router *gin.Engine) {
	docs.SwaggerInfo.BasePath = "/api/v1/"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
