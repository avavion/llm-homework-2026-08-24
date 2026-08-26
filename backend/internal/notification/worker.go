package notification

import (
	"context"
	"log"
	"time"
)

// Worker periodically triggers Service.DeliverDue until its context is
// cancelled, so the API process can run reminder delivery in the background.
type Worker struct {
	service  *Service
	interval time.Duration
}

func NewWorker(service *Service, interval time.Duration) *Worker {
	return &Worker{service: service, interval: interval}
}

func (worker *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(worker.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case tick := <-ticker.C:
			if err := worker.service.DeliverDue(ctx, tick); err != nil {
				log.Printf("notification worker: deliver due reminders: %v", err)
			}
		}
	}
}
