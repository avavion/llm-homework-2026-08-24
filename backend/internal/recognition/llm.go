package recognition

import (
	"context"
	"errors"
)

// ErrLLMUnavailable mirrors ErrOCRUnavailable for the LLM parsing step.
var ErrLLMUnavailable = errors.New("recognition: LLM provider is unavailable")

// LLMClient turns raw OCR text into draft product fields for a given locale.
type LLMClient interface {
	ParseProduct(ctx context.Context, text, locale string) (DraftFields, error)
}

// UnconfiguredLLMClient is the default until a real provider is wired in.
type UnconfiguredLLMClient struct{}

func (UnconfiguredLLMClient) ParseProduct(context.Context, string, string) (DraftFields, error) {
	return DraftFields{}, ErrLLMUnavailable
}
