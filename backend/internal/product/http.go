package product

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ServiceAPI is the subset of Service that HTTP handlers depend on, so tests
// can substitute a fake without a database.
type ServiceAPI interface {
	Create(ctx context.Context, accountID uuid.UUID, input CreateInput) (Product, error)
	Get(ctx context.Context, accountID, productID uuid.UUID) (Product, error)
	List(ctx context.Context, accountID uuid.UUID) ([]Product, error)
	Complete(ctx context.Context, accountID, productID uuid.UUID, status LifecycleStatus) (Product, error)
}

// AccountResolver extracts the authenticated account ID from a request,
// letting handlers stay agnostic of the session middleware implementation.
type AccountResolver func(c *gin.Context) (uuid.UUID, bool)

// DisplayStatusResolver keeps the HTTP package independent from the regulation
// package, which already depends on product domain types.
type DisplayStatusResolver interface {
	ResolveDisplayStatus(Product) DisplayStatus
}

// DisplayStatusFunc adapts a function for use as a DisplayStatusResolver.
type DisplayStatusFunc func(Product) DisplayStatus

func (resolver DisplayStatusFunc) ResolveDisplayStatus(item Product) DisplayStatus {
	return resolver(item)
}

type createRequest struct {
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

type PublicProduct struct {
	ID                    uuid.UUID  `json:"id"`
	Name                  string     `json:"name"`
	DateType              DateType   `json:"date_type"`
	ExpiryDate            time.Time  `json:"expiry_date"`
	Quantity              *float64   `json:"quantity,omitempty"`
	Unit                  *string    `json:"unit,omitempty"`
	ProductGroup          *string    `json:"product_group,omitempty"`
	StorageLocation       *string    `json:"storage_location,omitempty"`
	CountryCode           *string    `json:"country_code,omitempty"`
	AlertThresholdMinutes *int       `json:"alert_threshold_minutes,omitempty"`
	Status                string     `json:"status"`
	DisplayStatus         DisplayStatus `json:"display_status"`
	CompletedAt           *time.Time `json:"completed_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

func RegisterRoutes(router gin.IRouter, service ServiceAPI, resolveAccount AccountResolver, displayResolvers ...DisplayStatusResolver) {
	displayResolver := defaultDisplayStatusResolver
	if len(displayResolvers) > 0 && displayResolvers[0] != nil {
		displayResolver = displayResolvers[0]
	}
	router.POST("/v1/products", createHandler(service, resolveAccount, displayResolver))
	router.GET("/v1/products", listHandler(service, resolveAccount, displayResolver))
	router.GET("/v1/products/:id", getHandler(service, resolveAccount, displayResolver))
	router.POST("/v1/products/:id/use", completeHandler(service, resolveAccount, displayResolver, LifecycleUsed))
	router.POST("/v1/products/:id/discard", completeHandler(service, resolveAccount, displayResolver, LifecycleDiscarded))
}

func createHandler(service ServiceAPI, resolveAccount AccountResolver, displayResolver DisplayStatusResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID, ok := resolveAccount(c)
		if !ok {
			return
		}

		var request createRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			writeError(c, http.StatusBadRequest, "invalid request")
			return
		}

		created, err := service.Create(c.Request.Context(), accountID, CreateInput{
			Name:                  request.Name,
			DateType:              DateType(request.DateType),
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
		c.JSON(http.StatusCreated, ToPublicProduct(created, displayResolver))
	}
}

func listHandler(service ServiceAPI, resolveAccount AccountResolver, displayResolver DisplayStatusResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID, ok := resolveAccount(c)
		if !ok {
			return
		}

		items, err := service.List(c.Request.Context(), accountID)
		if writeServiceError(c, err) {
			return
		}

		results := make([]PublicProduct, 0, len(items))
		for _, item := range items {
			results = append(results, ToPublicProduct(item, displayResolver))
		}
		c.JSON(http.StatusOK, results)
	}
}

func getHandler(service ServiceAPI, resolveAccount AccountResolver, displayResolver DisplayStatusResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID, ok := resolveAccount(c)
		if !ok {
			return
		}

		productID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			writeError(c, http.StatusNotFound, ErrNotFound.Error())
			return
		}

		found, err := service.Get(c.Request.Context(), accountID, productID)
		if writeServiceError(c, err) {
			return
		}
		c.JSON(http.StatusOK, ToPublicProduct(found, displayResolver))
	}
}

func completeHandler(service ServiceAPI, resolveAccount AccountResolver, displayResolver DisplayStatusResolver, status LifecycleStatus) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID, ok := resolveAccount(c)
		if !ok {
			return
		}

		productID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			writeError(c, http.StatusNotFound, ErrNotFound.Error())
			return
		}

		completed, err := service.Complete(c.Request.Context(), accountID, productID, status)
		if writeServiceError(c, err) {
			return
		}
		c.JSON(http.StatusOK, ToPublicProduct(completed, displayResolver))
	}
}

func writeServiceError(c *gin.Context, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, ErrNotFound):
		writeError(c, http.StatusNotFound, ErrNotFound.Error())
	case errors.Is(err, ErrAlreadyCompleted):
		writeError(c, http.StatusConflict, ErrAlreadyCompleted.Error())
	case errors.Is(err, ErrInvalidProduct), errors.Is(err, ErrInvalidDateType),
		errors.Is(err, ErrThresholdTooSmall), errors.Is(err, ErrInvalidStatus):
		writeError(c, http.StatusBadRequest, err.Error())
	default:
		writeError(c, http.StatusInternalServerError, "internal server error")
	}
	return true
}

func writeError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}

var defaultDisplayStatusResolver DisplayStatusResolver = DisplayStatusFunc(func(item Product) DisplayStatus {
	switch item.LifecycleStatus {
	case LifecycleUsed:
		return DisplayStatusUsed
	case LifecycleDiscarded:
		return DisplayStatusDiscarded
	default:
		return DisplayStatusResearchRequired
	}
})

func ToPublicProduct(item Product, displayResolvers ...DisplayStatusResolver) PublicProduct {
	displayResolver := defaultDisplayStatusResolver
	if len(displayResolvers) > 0 && displayResolvers[0] != nil {
		displayResolver = displayResolvers[0]
	}

	return PublicProduct{
		ID:                    item.ID,
		Name:                  item.Name,
		DateType:              item.DateType,
		ExpiryDate:            item.ExpiryDate,
		Quantity:              item.Quantity,
		Unit:                  item.Unit,
		ProductGroup:          item.ProductGroup,
		StorageLocation:       item.StorageLocation,
		CountryCode:           item.CountryCode,
		AlertThresholdMinutes: item.AlertThresholdMinutes,
		Status:                string(item.LifecycleStatus),
		DisplayStatus:         displayResolver.ResolveDisplayStatus(item),
		CompletedAt:           item.CompletedAt,
		CreatedAt:             item.CreatedAt,
		UpdatedAt:             item.UpdatedAt,
	}
}
