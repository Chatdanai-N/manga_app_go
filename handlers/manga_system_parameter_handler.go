package handlers

import (
	"errors"
	"log/slog"
	"manga_app/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RequestMangaParameter struct {
	ParameterId uint64 `json:"parameterId"`
	GroupCode   string `json:"groupCode" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Value       string `json:"value" binding:"required"`
	OrderNumber uint64 `json:"orderNumber" binding:"required"`
}

type RequestPatchMangaSystemParameter struct {
	Value string `json:"value" binding:"required"`
}

type MangaSystemParameterHandler struct {
	DB *gorm.DB
}

func NewMangaSystemParameterHandler(db *gorm.DB) *MangaSystemParameterHandler {
	return &MangaSystemParameterHandler{DB: db}
}

func (h *MangaSystemParameterHandler) GetMangaSystemParameters(c *gin.Context) {
	var mangaSystemParameters []models.MangaSystemParameter
	if err := h.DB.Find(&mangaSystemParameters).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mangaSystemParameters)
}

func (h *MangaSystemParameterHandler) GetMangaSystemParameterById(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid parameter id format", slog.String("id", idStr), slog.String("err", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var param models.MangaSystemParameter

	if err := h.DB.First(&param, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(c.Request.Context(), "manga system parameter not found", slog.Uint64("id", param.ParameterId), slog.String("error", err.Error()))
			c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
			return
		}

		slog.ErrorContext(c.Request.Context(), "failed to query manga system parameter", slog.Uint64("id", param.ParameterId), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, param)

}

func (h *MangaSystemParameterHandler) CreateMangaSystemParameter(c *gin.Context) {

	// 1. Struct สำหรับรับ JSON Request Body
	var requestCreated RequestMangaParameter

	if err := c.ShouldBindJSON(&requestCreated); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid Body Request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Map เข้า Model
	param := models.MangaSystemParameter{
		GroupCode:   requestCreated.GroupCode,
		Code:        requestCreated.Code,
		Value:       requestCreated.Value,
		OrderNumber: requestCreated.OrderNumber,
	}

	// 3. สั่ง Insert ลง Database
	if err := h.DB.Create(&param).Error; err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to create manga system parameter", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create record"})
		return
	}
	c.JSON(http.StatusCreated, param)
}

func (h *MangaSystemParameterHandler) PutMangaSystemParameter(c *gin.Context) {

	// 1. Struct สำหรับรับ JSON Request Body
	var updateRequest RequestMangaParameter

	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid Body Request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var param models.MangaSystemParameter
	if err := h.DB.First(&param, updateRequest.ParameterId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(c.Request.Context(), "manga system parameter not found", slog.Uint64("id", updateRequest.ParameterId), slog.String("error", err.Error()))
			c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "failed to query manga system parameter", slog.Uint64("id", updateRequest.ParameterId), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	param.GroupCode = updateRequest.GroupCode
	param.Code = updateRequest.Code
	param.Value = updateRequest.Value
	param.OrderNumber = updateRequest.OrderNumber

	if err := h.DB.Save(&param).Error; err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to update manga system parameter", slog.Uint64("id", param.ParameterId), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update record"})
		return
	}

	c.JSON(http.StatusOK, param)
}

func (h *MangaSystemParameterHandler) PatchMangaSystemParameter(c *gin.Context) {

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "Invalid parameter id format", slog.String("id", idStr), slog.String("err", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var requestPatch RequestPatchMangaSystemParameter
	if err := c.ShouldBindJSON(&requestPatch); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid Body Request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var param models.MangaSystemParameter
	if err := h.DB.First(&param, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(c.Request.Context(), "manga system parameter not found", slog.Uint64("id", param.ParameterId), slog.String("error", err.Error()))
			c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "failed to query manga system parameter", slog.Uint64("id", param.ParameterId), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var updateParam models.MangaSystemParameter
	updateParam.Value = requestPatch.Value

	// ใช้ db.Model().Updates() (แนะนำสำหรับ Partial Update)
	// หากต้องการอัปเดต เฉพาะบางฟิลด์ที่ส่งมา ( Partial Update / PATCH) นิยมใช้ .Updates()
	// อัปเดตเฉพาะ column "value" ของ parameter_id นั้นๆ
	if err := h.DB.Model(&param).Updates(&updateParam).Error; err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to update manga system parameter", slog.Uint64("id", param.ParameterId), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update record"})
		return
	}

	c.JSON(http.StatusOK, param)
}

func (h *MangaSystemParameterHandler) DeleteMangaSystemParameterById(c *gin.Context) {

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid parameter id format", slog.String("id", idStr), slog.String("err", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	result := h.DB.Delete(&models.MangaSystemParameter{}, id)
	if result.Error != nil {
		slog.ErrorContext(c.Request.Context(), "failed to delete manga system parameter by id", slog.Uint64("id", id), slog.String("error", result.Error.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if result.RowsAffected == 0 {
		slog.WarnContext(c.Request.Context(), "not found manga system parameter by id", slog.Uint64("id", id))
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully deleted record"})
}
