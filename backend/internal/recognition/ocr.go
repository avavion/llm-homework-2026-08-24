package recognition

import (
	"context"
	"errors"
)

// ErrOCRUnavailable is a stable, secret-free failure the service maps to a
// generic client response telling the user to fall back to the manual form.
var ErrOCRUnavailable = errors.New("recognition: OCR provider is unavailable")

// OCRClient extracts raw text from a photographed package. Real providers
// (a cloud OCR API, a local engine) implement this without changing the
// service that consumes it.
type OCRClient interface {
	ExtractText(ctx context.Context, image []byte) (string, error)
}

// UnconfiguredOCRClient is the default until a real provider is wired in; it
// always reports unavailability so the client is told to use the manual form.
type UnconfiguredOCRClient struct{}

func (UnconfiguredOCRClient) ExtractText(context.Context, []byte) (string, error) {
	return "", ErrOCRUnavailable
}
