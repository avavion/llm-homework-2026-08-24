package recognition

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"llm-homework/backend/internal/product"
)

// ServiceAPI is the subset of Service the HTTP layer depends on.
type ServiceAPI interface {
	Recognize(ctx context.Context, accountID uuid.UUID, image []byte, locale, sourceReference string) (ProductDraft, error)
	Get(ctx context.Context, accountID, draftID uuid.UUID) (ProductDraft, error)
	Approve(ctx context.Context, accountID, draftID uuid.UUID, edited product.CreateInput) (product.Product, error)
	Reject(ctx context.Context, accountID, draftID uuid.UUID) (ProductDraft, error)
}

// AccountResolver mirrors product.AccountResolver so both modules share the
// same session-to-account boundary without importing each other.
type AccountResolver func(c *gin.Context) (uuid.UUID, bool)

const maxImageBytes = 10 << 20 // 10 MiB

type approveRequest struct {
	Name                  string    `json:"name"`
	DateType              string    `json:"date_type"`
	ExpiryDate            time.Time `json:"expiry_date"`
	Quantity              *float64  `json:"quantity"`
	Unit                  *string   `json:"unit"`
	ProductGroup          *string   `json:"product_group"`
	StorageLocation       *string   `json:"storage_location"`
	CountryCode           *string   `json:"country_code"`
	AlertThresholdMinutes *int      `json:"alert_threshold_minutes"`
}

type draftResponse struct {
	ID              uuid.UUID   `json:"id"`
	Status          DraftStatus `json:"status"`
	RawText         string      `json:"raw_text,omitempty"`
	Fields          DraftFields `json:"fields"`
	ApprovedProduct *uuid.UUID  `json:"approved_product_id,omitempty"`
}

func RegisterRoutes(router gin.IRouter, service ServiceAPI, resolveAccount AccountResolver) {
	router.POST("/v1/product-drafts/recognize", recognizeHandler(service, resolveAccount))
	router.GET("/v1/product-drafts/:id", getHandler(service, resolveAccount))
	router.POST("/v1/product-drafts/:id/approve", approveHandler(service, resolveAccount))
	router.POST("/v1/product-drafts/:id/reject", rejectHandler(service, resolveAccount))
}

func getHandler(service ServiceAPI, resolveAccount AccountResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID, ok := resolveAccount(c)
		if !ok {
			return
		}

		draftID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			writeError(c, http.StatusNotFound, ErrNotFound.Error())
			return
		}

		draft, err := service.Get(c.Request.Context(), accountID, draftID)
		if writeServiceError(c, err) {
			return
		}
		c.JSON(http.StatusOK, toDraftResponse(draft))
	}
}

func recognizeHandler(service ServiceAPI, resolveAccount AccountResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID, ok := resolveAccount(c)
		if !ok {
			return
		}

		file, header, err := c.Request.FormFile("image")
		if err != nil {
			writeError(c, http.StatusBadRequest, "an image file is required")
			return
		}
		defer file.Close()

		image, err := io.ReadAll(io.LimitReader(file, maxImageBytes+1))
		if err != nil || len(image) > maxImageBytes {
			writeError(c, http.StatusBadRequest, "invalid image upload")
			return
		}

		locale := c.PostForm("locale")
		draft, err := service.Recognize(c.Request.Context(), accountID, image, locale, header.Filename)
		if writeServiceError(c, err) {
			return
		}
		c.JSON(http.StatusCreated, toDraftResponse(draft))
	}
}

func approveHandler(service ServiceAPI, resolveAccount AccountResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID, ok := resolveAccount(c)
		if !ok {
			return
		}

		draftID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			writeError(c, http.StatusNotFound, ErrNotFound.Error())
			return
		}

		var request approveRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			writeError(c, http.StatusBadRequest, "invalid request")
			return
		}

		created, err := service.Approve(c.Request.Context(), accountID, draftID, product.CreateInput{
			Name:                  request.Name,
			DateType:              product.DateType(request.DateType),
			ExpiryDate:            request.ExpiryDate,
			Quantity:              request.Quantity,
			Unit:                  request.Unit,
			ProductGroup:          request.ProductGroup,
			StorageLocation:       request.StorageLocation,
			CountryCode:           request.CountryCode,
			AlertThresholdMinutes: request.AlertThresholdMinutes,
		})
		if writeServiceError(c, err) {
			return
		}
		c.JSON(http.StatusCreated, product.ToPublicProduct(created))
	}
}

func rejectHandler(service ServiceAPI, resolveAccount AccountResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID, ok := resolveAccount(c)
		if !ok {
			return
		}

		draftID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			writeError(c, http.StatusNotFound, ErrNotFound.Error())
			return
		}

		_, err = service.Reject(c.Request.Context(), accountID, draftID)
		if writeServiceError(c, err) {
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func writeServiceError(c *gin.Context, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, ErrNotFound):
		writeError(c, http.StatusNotFound, ErrNotFound.Error())
	case errors.Is(err, ErrAlreadyDecided):
		writeError(c, http.StatusConflict, ErrAlreadyDecided.Error())
	case errors.Is(err, ErrEmptyImage), errors.Is(err, ErrRecognitionUnavailable),
		errors.Is(err, product.ErrInvalidProduct), errors.Is(err, product.ErrInvalidDateType),
		errors.Is(err, product.ErrThresholdTooSmall):
		writeError(c, http.StatusBadRequest, err.Error())
	default:
		writeError(c, http.StatusInternalServerError, "internal server error")
	}
	return true
}

func writeError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}

func toDraftResponse(draft ProductDraft) draftResponse {
	return draftResponse{
		ID:              draft.ID,
		Status:          draft.Status,
		RawText:         draft.RawText,
		Fields:          draft.Fields,
		ApprovedProduct: draft.ApprovedProductID,
	}
}
