package router

import (
	"log"
	"manga_app/config"
	"manga_app/handlers"
	"manga_app/middleware"
	"manga_app/validatorutil"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
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

	// registration custom validation
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {

		if err := SetUpCustomValidators(v); err != nil {
			log.Fatalf("failed to set up custom validators %v", err)
		}
	}

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
		apiV1.POST("/signup", mangaLoginHandler.PostSignup)
		// Parameter Group
		params := apiV1.Group("/parameter")
		params.Use(middleware.AuthMiddleware(cfg.JWTSecret))
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

func SetUpCustomValidators(v *validator.Validate) error {

	if err := v.RegisterValidation("is_valid_email", validatorutil.IsValidEmail); err != nil {
		return err
	}

	if err := v.RegisterValidation("len_password", validatorutil.LenPassword); err != nil {
		return err
	}

	if err := v.RegisterValidation("has_upper", validatorutil.HasUpper); err != nil {
		return err
	}

	if err := v.RegisterValidation("has_lower", validatorutil.HasLower); err != nil {
		return err
	}

	if err := v.RegisterValidation("has_number", validatorutil.HasUpper); err != nil {
		return err
	}

	if err := v.RegisterValidation("has_special", validatorutil.HasSpacial); err != nil {

	}

	return nil
}
