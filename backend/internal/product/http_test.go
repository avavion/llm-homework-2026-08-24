package product

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestRouter wires the product routes behind a resolver that trusts an
// "X-Test-Account" header, so handler tests can exercise account scoping
// without a real session cookie.
func newTestRouter(service ServiceAPI) *gin.Engine {
	router := gin.New()
	resolveAccount := func(c *gin.Context) (uuid.UUID, bool) {
		accountID, err := uuid.Parse(c.GetHeader("X-Test-Account"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return uuid.UUID{}, false
		}
		return accountID, true
	}
	RegisterRoutes(router, service, resolveAccount)
	return router
}

func performAs(router *gin.Engine, accountID uuid.UUID, method, path string, body any) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body != nil {
		encoded, _ := json.Marshal(body)
		reader = bytes.NewReader(encoded)
	} else {
		reader = bytes.NewReader(nil)
	}
	request := httptest.NewRequest(method, path, reader)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Test-Account", accountID.String())
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestCreateReturnsBadRequestForMissingFields(t *testing.T) {
	router := newTestRouter(NewService(newMemoryRepository()))
	accountID := uuid.New()

	rr := performAs(router, accountID, http.MethodPost, "/v1/products", map[string]any{})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestCreateListAndGetRoundTrip(t *testing.T) {
	router := newTestRouter(NewService(newMemoryRepository()))
	accountID := uuid.New()

	created := performAs(router, accountID, http.MethodPost, "/v1/products", map[string]any{
		"name": "Milk", "date_type": "use_by", "expiry_date": time.Now().Add(24 * time.Hour),
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d, body=%s", created.Code, http.StatusCreated, created.Body.String())
	}
	var createdBody response
	if err := json.Unmarshal(created.Body.Bytes(), &createdBody); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	list := performAs(router, accountID, http.MethodGet, "/v1/products", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", list.Code, http.StatusOK)
	}
	var listBody []response
	if err := json.Unmarshal(list.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(listBody) != 1 {
		t.Fatalf("list length = %d, want 1", len(listBody))
	}

	get := performAs(router, accountID, http.MethodGet, "/v1/products/"+createdBody.ID.String(), nil)
	if get.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", get.Code, http.StatusOK)
	}
}

func TestForeignProductIsNotVisible(t *testing.T) {
	router := newTestRouter(NewService(newMemoryRepository()))
	owner := uuid.New()
	stranger := uuid.New()

	created := performAs(router, owner, http.MethodPost, "/v1/products", map[string]any{
		"name": "Milk", "date_type": "use_by", "expiry_date": time.Now().Add(24 * time.Hour),
	})
	var createdBody response
	_ = json.Unmarshal(created.Body.Bytes(), &createdBody)

	rr := performAs(router, stranger, http.MethodGet, "/v1/products/"+createdBody.ID.String(), nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestUseAndDiscardProduceDistinctStatuses(t *testing.T) {
	router := newTestRouter(NewService(newMemoryRepository()))
	accountID := uuid.New()

	first := performAs(router, accountID, http.MethodPost, "/v1/products", map[string]any{
		"name": "Milk", "date_type": "use_by", "expiry_date": time.Now().Add(24 * time.Hour),
	})
	var firstBody response
	_ = json.Unmarshal(first.Body.Bytes(), &firstBody)

	used := performAs(router, accountID, http.MethodPost, "/v1/products/"+firstBody.ID.String()+"/use", nil)
	if used.Code != http.StatusOK {
		t.Fatalf("use status = %d, want %d", used.Code, http.StatusOK)
	}
	var usedBody response
	_ = json.Unmarshal(used.Body.Bytes(), &usedBody)
	if usedBody.Status != "used" || usedBody.CompletedAt == nil {
		t.Fatalf("used body = %+v", usedBody)
	}

	repeat := performAs(router, accountID, http.MethodPost, "/v1/products/"+firstBody.ID.String()+"/discard", nil)
	if repeat.Code != http.StatusConflict {
		t.Fatalf("repeat completion status = %d, want %d", repeat.Code, http.StatusConflict)
	}
}
