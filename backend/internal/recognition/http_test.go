package recognition

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
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

func multipartImageRequest(accountID uuid.UUID) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("image", "photo.jpg")
	_, _ = part.Write([]byte("fake-image-bytes"))
	_ = writer.WriteField("locale", "en")
	_ = writer.Close()

	request := httptest.NewRequest(http.MethodPost, "/v1/product-drafts/recognize", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("X-Test-Account", accountID.String())
	return request
}

func TestRecognizeEndpointCreatesPendingDraft(t *testing.T) {
	name := "Milk"
	router := newTestRouter(NewService(newMemoryRepository(), fakeOCR{text: "milk"}, fakeLLM{fields: DraftFields{Name: &name}}))
	accountID := uuid.New()

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, multipartImageRequest(accountID))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
}

func TestApproveAndForeignOwnerAndDoubleApprove(t *testing.T) {
	repository := newMemoryRepository()
	router := newTestRouter(NewService(repository, fakeOCR{}, fakeLLM{}))
	owner := uuid.New()
	stranger := uuid.New()

	draft, _ := repository.Create(context.Background(), owner, DraftFields{}, "text", "photo.jpg")

	edited := map[string]any{
		"name": "Milk", "date_type": "use_by", "expiry_date": time.Now().Add(24 * time.Hour),
	}
	encoded, _ := json.Marshal(edited)

	foreign := httptest.NewRequest(http.MethodPost, "/v1/product-drafts/"+draft.ID.String()+"/approve", bytes.NewReader(encoded))
	foreign.Header.Set("Content-Type", "application/json")
	foreign.Header.Set("X-Test-Account", stranger.String())
	foreignRecorder := httptest.NewRecorder()
	router.ServeHTTP(foreignRecorder, foreign)
	if foreignRecorder.Code != http.StatusNotFound {
		t.Fatalf("foreign approve status = %d, want %d", foreignRecorder.Code, http.StatusNotFound)
	}

	approve := httptest.NewRequest(http.MethodPost, "/v1/product-drafts/"+draft.ID.String()+"/approve", bytes.NewReader(encoded))
	approve.Header.Set("Content-Type", "application/json")
	approve.Header.Set("X-Test-Account", owner.String())
	approveRecorder := httptest.NewRecorder()
	router.ServeHTTP(approveRecorder, approve)
	if approveRecorder.Code != http.StatusCreated {
		t.Fatalf("approve status = %d, want %d, body=%s", approveRecorder.Code, http.StatusCreated, approveRecorder.Body.String())
	}

	repeat := httptest.NewRequest(http.MethodPost, "/v1/product-drafts/"+draft.ID.String()+"/approve", bytes.NewReader(encoded))
	repeat.Header.Set("Content-Type", "application/json")
	repeat.Header.Set("X-Test-Account", owner.String())
	repeatRecorder := httptest.NewRecorder()
	router.ServeHTTP(repeatRecorder, repeat)
	if repeatRecorder.Code != http.StatusConflict {
		t.Fatalf("repeat approve status = %d, want %d", repeatRecorder.Code, http.StatusConflict)
	}
}

func TestRejectReturnsNoContentAndCreatesNoProduct(t *testing.T) {
	repository := newMemoryRepository()
	router := newTestRouter(NewService(repository, fakeOCR{}, fakeLLM{}))
	owner := uuid.New()

	draft, _ := repository.Create(context.Background(), owner, DraftFields{}, "text", "photo.jpg")

	request := httptest.NewRequest(http.MethodPost, "/v1/product-drafts/"+draft.ID.String()+"/reject", nil)
	request.Header.Set("X-Test-Account", owner.String())
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if repository.productCount(owner) != 0 {
		t.Fatal("reject must not create a product")
	}
}
