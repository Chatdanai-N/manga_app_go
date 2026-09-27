package router

import (
	"manga_app/handlers"
	"manga_app/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	// ใช้ gin.New() แทน gin.Default() เพื่อปิด Built-in Logger ของ Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Attach Custom JSON Logger และ Recovery Middleware
	r.Use(middleware.StructuredLogger())
	r.Use(gin.Recovery())

	// initialize Handler
	mangaSystemParameterHandler := handlers.NewMangaSystemParameterHandler(db)

	// Health Check Endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP"})
	})

	apiV1 := r.Group("/api/v1")
	{

		// Parameter Group
		params := apiV1.Group("/parameter")
		{
			params.GET("", mangaSystemParameterHandler.GetMangaSystemParameters)
			params.GET("/:id", mangaSystemParameterHandler.GetMangaSystemParameterById)
			params.POST("", mangaSystemParameterHandler.CreateMangaSystemParameter)
			params.PUT("", mangaSystemParameterHandler.PutMangaSystemParameter)
			params.PATCH("/:id", mangaSystemParameterHandler.PatchMangaSystemParameter)
			params.DELETE("/:id", mangaSystemParameterHandler.DeleteMangaSystemParameterById)
		}

	}
	return r
}
