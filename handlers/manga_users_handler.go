package handlers

import (
	"errors"
	"log/slog"
	"manga_app/config"
	"manga_app/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SignupRequest struct {
	Username        string `json:"username" binding:"required"`
	Password        string `json:"password" binding:"required,len_password,has_upper,has_lower,has_number,has_special"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
	Email           string `json:"email" binding:"required,is_valid_email"`
}

type LoginHandler struct {
	DB     *gorm.DB
	Config *config.Config
}

func NewLoginHandler(db *gorm.DB, cfg *config.Config) *LoginHandler {
	return &LoginHandler{DB: db, Config: cfg}
}

func (h *LoginHandler) PostLogin(c *gin.Context) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// query check username/password
	var param models.MangaUsers

	if err := h.DB.Where("username = ?", request.Username).First(&param).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(c.Request.Context(), "username not found")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		slog.WarnContext(c.Request.Context(), "error finding user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(param.PasswordHash), []byte(request.Password)); err != nil {
		slog.WarnContext(c.Request.Context(), err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid password"})
		return
	}

	//set expire date token 2 hour
	expirationTime := time.Now().Add(2 * time.Hour)

	claims := &Claims{
		Username: request.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "manga_app",
		},
	}

	// get JWTSecret with config
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.Config.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
	}

	// response Token to client
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func (h *LoginHandler) PostSignup(c *gin.Context) {
	var request SignupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid Body Request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error register": err.Error()})
		return
	}

	// Generate password hash
	passwordByte, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "error hash password", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	param := models.MangaUsers{
		Username:     request.Username,
		PasswordHash: string(passwordByte),
		Email:        request.Email,
		CreatedDate:  time.Now(),
	}

	if err := h.DB.Create(&param).Error; err != nil {
		slog.WarnContext(c.Request.Context(), "error create user", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": param})

}
