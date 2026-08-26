package notification

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"llm-homework/backend/internal/product"
	"llm-homework/backend/internal/regulation"
)

type fakeProductSource struct {
	candidates []Candidate
}

func (source fakeProductSource) ActiveWithCountry(context.Context) ([]Candidate, error) {
	return source.candidates, nil
}

type fakeRuleLookup struct {
	rule regulation.Rule
	has  bool
}

func (lookup fakeRuleLookup) RuleFor(string, product.DateType) (regulation.Rule, bool) {
	return lookup.rule, lookup.has
}

type fakeDeliveryLog struct {
	claimed map[string]bool
}

func newFakeDeliveryLog() *fakeDeliveryLog {
	return &fakeDeliveryLog{claimed: map[string]bool{}}
}

func (store *fakeDeliveryLog) Claim(_ context.Context, productID uuid.UUID, scheduledFor time.Time, channel string) (bool, error) {
	key := productID.String() + scheduledFor.String() + channel
	if store.claimed[key] {
		return false, nil
	}
	store.claimed[key] = true
	return true, nil
}

type fakeSender struct {
	Calls int
}

func (sender *fakeSender) SendExpiryReminder(context.Context, string, product.Product) error {
	sender.Calls++
	return nil
}

func enabledRule() regulation.Rule {
	return regulation.Rule{
		DateType: product.DateTypeUseBy, Status: regulation.StatusEnabled,
		ExpiryTimezone: "UTC", ExpiryInstantRule: regulation.ExpiryAtEndOfDay,
		PostExpiryStatus: regulation.PostExpiryExpired,
	}
}

func TestThresholdBelowSixtyMinutesIsRejected(t *testing.T) {
	_, err := NextNotificationAt(time.Now(), 59*time.Minute)
	if err != ErrThresholdTooSmall {
		t.Fatalf("err = %v, want ErrThresholdTooSmall", err)
	}
	if _, err := NextNotificationAt(time.Now(), 60*time.Minute); err != nil {
		t.Fatalf("60 minutes should be accepted, got %v", err)
	}
}

func TestSchedulerDoesNotSendSameReminderTwice(t *testing.T) {
	country := "ZZ"
	threshold := 60
	expiry := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	item := product.Product{
		ID: uuid.New(), Name: "Milk", DateType: product.DateTypeUseBy, ExpiryDate: expiry,
		CountryCode: &country, AlertThresholdMinutes: &threshold, LifecycleStatus: product.LifecycleActive,
	}
	sender := &fakeSender{}
	service := NewService(
		fakeProductSource{candidates: []Candidate{{Product: item, RecipientEmail: "user@example.com"}}},
		fakeRuleLookup{rule: enabledRule(), has: true},
		newFakeDeliveryLog(),
		sender,
	)

	now := expiry.Add(23*time.Hour + 59*time.Minute + 1*time.Second) // one second past the alert instant
	if err := service.DeliverDue(context.Background(), now); err != nil {
		t.Fatalf("first DeliverDue err = %v", err)
	}
	if err := service.DeliverDue(context.Background(), now); err != nil {
		t.Fatalf("second DeliverDue err = %v", err)
	}
	if sender.Calls != 1 {
		t.Fatalf("sender.Calls = %d, want 1", sender.Calls)
	}
}

func TestUnverifiedRuleProducesNoDelivery(t *testing.T) {
	country := "ZZ"
	threshold := 60
	item := product.Product{
		ID: uuid.New(), Name: "Milk", DateType: product.DateTypeUseBy, ExpiryDate: time.Now().Add(-48 * time.Hour),
		CountryCode: &country, AlertThresholdMinutes: &threshold, LifecycleStatus: product.LifecycleActive,
	}
	sender := &fakeSender{}
	service := NewService(
		fakeProductSource{candidates: []Candidate{{Product: item, RecipientEmail: "user@example.com"}}},
		fakeRuleLookup{rule: regulation.Rule{Status: regulation.StatusResearchRequired}, has: true},
		newFakeDeliveryLog(),
		sender,
	)

	if err := service.DeliverDue(context.Background(), time.Now()); err != nil {
		t.Fatalf("DeliverDue err = %v", err)
	}
	if sender.Calls != 0 {
		t.Fatalf("sender.Calls = %d, want 0 for a research_required rule", sender.Calls)
	}
}
