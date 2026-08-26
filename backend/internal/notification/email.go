package notification

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"llm-homework/backend/internal/product"
)

// EmailSender delivers a single expiry reminder. Implementations must never
// leak provider credentials or secrets through their error values.
type EmailSender interface {
	SendExpiryReminder(ctx context.Context, recipient string, item product.Product) error
}

type SentMessage struct {
	Recipient   string
	ProductID   uuid.UUID
	ProductName string
	SentAt      time.Time
}

// DevSender is the MVP's development mail sender: it never talks to a real
// provider, records what would have been sent, and only logs safe fields.
type DevSender struct {
	mu   sync.Mutex
	Sent []SentMessage
}

func NewDevSender() *DevSender {
	return &DevSender{}
}

func (sender *DevSender) SendExpiryReminder(_ context.Context, recipient string, item product.Product) error {
	sender.mu.Lock()
	defer sender.mu.Unlock()

	sender.Sent = append(sender.Sent, SentMessage{
		Recipient:   recipient,
		ProductID:   item.ID,
		ProductName: item.Name,
		SentAt:      time.Now(),
	})
	log.Printf("dev-mail: expiry reminder queued for product %q", item.Name)
	return nil
}
