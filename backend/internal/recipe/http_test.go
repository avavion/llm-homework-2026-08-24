package recipe

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"llm-homework/backend/internal/product"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type fakeProductLister struct {
	items []product.Product
}

func (lister fakeProductLister) List(context.Context, uuid.UUID) ([]product.Product, error) {
	return lister.items, nil
}

func TestRecipesEndpointReturnsSuggestionsForCaller(t *testing.T) {
	item := product.Product{ID: uuid.New(), Name: "Milk", LifecycleStatus: product.LifecycleActive}
	router := gin.New()
	resolveAccount := func(c *gin.Context) (uuid.UUID, bool) { return uuid.New(), true }
	RegisterRoutes(router, NewService(fakeRules{}), fakeProductLister{items: []product.Product{item}}, resolveAccount)

	request := httptest.NewRequest(http.MethodGet, "/v1/recipes", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var body []recipeResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(body) != 1 || body[0].Kind != KindUseUp || body[0].ProductName != "Milk" {
		t.Fatalf("body = %+v", body)
	}
}
