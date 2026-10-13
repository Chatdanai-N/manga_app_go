package router

import (
	"manga_app/config"
	"manga_app/handlers"
	"manga_app/middleware"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func SetupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	// ใช้ gin.New() แทน gin.Default() เพื่อปิด Built-in Logger ของ Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Attach Custom JSON Logger และ Recovery Middleware
	r.Use(middleware.StructuredLogger())
	r.Use(gin.Recovery())

	// initialize Handler
	mangaLoginHandler := handlers.NewLoginHandler(db, cfg)
	mangaSystemParameterHandler := handlers.NewMangaSystemParameterHandler(db)

	// Health Check Endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP"})
	})

	apiV1 := r.Group("/api/v1")
	{

		apiV1.POST("/login", mangaLoginHandler.PostLogin)

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

// ยังไม่เสร็จ
func jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid format token"})
		}

	}
}
