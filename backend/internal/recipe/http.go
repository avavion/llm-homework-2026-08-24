package recipe

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"llm-homework/backend/internal/product"
)

// ProductLister is the subset of product.Service the recipe endpoint needs to
// read a user's own products.
type ProductLister interface {
	List(ctx context.Context, accountID uuid.UUID) ([]product.Product, error)
}

type AccountResolver func(c *gin.Context) (uuid.UUID, bool)

type recipeResponse struct {
	Kind        Kind        `json:"kind"`
	ProductName string      `json:"product_name,omitempty"`
	GroupName   string      `json:"group_name,omitempty"`
	ProductIDs  []uuid.UUID `json:"product_ids"`
}

func RegisterRoutes(router gin.IRouter, service *Service, products ProductLister, resolveAccount AccountResolver) {
	router.GET("/v1/recipes", func(c *gin.Context) {
		accountID, ok := resolveAccount(c)
		if !ok {
			return
		}

		items, err := products.List(c.Request.Context(), accountID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		suggestions := service.Suggest(items)
		results := make([]recipeResponse, 0, len(suggestions))
		for _, suggestion := range suggestions {
			results = append(results, recipeResponse{Kind: suggestion.Kind, ProductName: suggestion.ProductName, GroupName: suggestion.GroupName, ProductIDs: suggestion.ProductIDs})
		}
		c.JSON(http.StatusOK, results)
	})
}
