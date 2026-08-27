package recognition

import (
	"context"
	"time"

	"llm-homework/backend/internal/product"
)

const mockDraftLifetime = 7 * 24 * time.Hour

// MockOCRClient is a canned stand-in for a real OCR provider, useful for
// exercising the recognize -> approve/reject flow without external services.
type MockOCRClient struct{}

func (MockOCRClient) ExtractText(context.Context, []byte) (string, error) {
	return "Молоко 3.2% пастеризованное\nГоден до: см. дату на упаковке", nil
}

// MockLLMClient is a canned stand-in for a real LLM provider; it always
// returns the same plausible draft regardless of input.
type MockLLMClient struct{}

func (MockLLMClient) ParseProduct(context.Context, string, string) (DraftFields, error) {
	name := "Молоко 3.2%"
	dateType := product.DateTypeUseBy
	expiryDate := time.Now().Add(mockDraftLifetime).Truncate(24 * time.Hour)
	quantity := 1.0
	unit := "л"

	return DraftFields{
		Name:       &name,
		DateType:   &dateType,
		ExpiryDate: &expiryDate,
		Quantity:   &quantity,
		Unit:       &unit,
	}, nil
}

// Clients selects the OCR/LLM client pair for the given provider name.
// Unknown or empty providers fall back to the unconfigured clients, which
// always report unavailability so the caller uses the manual form.
func Clients(provider string) (OCRClient, LLMClient) {
	if provider == "mock" {
		return MockOCRClient{}, MockLLMClient{}
	}
	return UnconfiguredOCRClient{}, UnconfiguredLLMClient{}
}
