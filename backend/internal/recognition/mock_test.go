package recognition

import (
	"context"
	"testing"
	"time"

	"llm-homework/backend/internal/product"
)

func TestMockOCRClientExtractsPlausibleText(t *testing.T) {
	client := MockOCRClient{}

	text, err := client.ExtractText(context.Background(), []byte("irrelevant"))
	if err != nil {
		t.Fatalf("ExtractText() error = %v", err)
	}
	if text == "" {
		t.Fatal("ExtractText() returned empty text, want a plausible OCR result")
	}
}

func TestMockLLMClientParsesPlausibleDraft(t *testing.T) {
	client := MockLLMClient{}

	fields, err := client.ParseProduct(context.Background(), "irrelevant", "ru")
	if err != nil {
		t.Fatalf("ParseProduct() error = %v", err)
	}
	if fields.Name == nil || *fields.Name == "" {
		t.Fatal("ParseProduct() Name is nil/empty, want a plausible product name")
	}
	if fields.DateType == nil || *fields.DateType != product.DateTypeUseBy {
		t.Fatalf("ParseProduct() DateType = %v, want %v", fields.DateType, product.DateTypeUseBy)
	}
	if fields.ExpiryDate == nil || !fields.ExpiryDate.After(time.Now()) {
		t.Fatal("ParseProduct() ExpiryDate is nil or not in the future, want a plausible future date")
	}
}

func TestClientsForMockProviderReturnsMockClients(t *testing.T) {
	ocrClient, llmClient := Clients("mock")

	if _, ok := ocrClient.(MockOCRClient); !ok {
		t.Fatalf("Clients(\"mock\") ocr = %T, want MockOCRClient", ocrClient)
	}
	if _, ok := llmClient.(MockLLMClient); !ok {
		t.Fatalf("Clients(\"mock\") llm = %T, want MockLLMClient", llmClient)
	}
}

func TestClientsForUnknownProviderReturnsUnconfiguredClients(t *testing.T) {
	ocrClient, llmClient := Clients("")

	if _, ok := ocrClient.(UnconfiguredOCRClient); !ok {
		t.Fatalf("Clients(\"\") ocr = %T, want UnconfiguredOCRClient", ocrClient)
	}
	if _, ok := llmClient.(UnconfiguredLLMClient); !ok {
		t.Fatalf("Clients(\"\") llm = %T, want UnconfiguredLLMClient", llmClient)
	}
}
