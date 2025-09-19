package service

import (
	"context"
	"fmt"
	"github.com/bezjen/gophermart/internal/model"
	"github.com/bezjen/gophermart/internal/repository"
	"github.com/go-resty/resty/v2"
	"net/http"
	"time"
)

type AccrualService interface {
	StartOrderProcessingWorker(ctx context.Context, interval time.Duration)
}

type AccrualRestService struct {
	client         *resty.Client
	accrualBaseURL string
	storage        repository.Repository
}

func NewAccrualRestService(accrualBaseURL string, storage repository.Repository) *AccrualRestService {
	return &AccrualRestService{
		client:         resty.New().SetBaseURL(accrualBaseURL),
		accrualBaseURL: accrualBaseURL,
		storage:        storage,
	}
}

func (s *AccrualRestService) StartOrderProcessingWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval) // TODO: add batching
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.processPendingOrders(ctx); err != nil {
				fmt.Printf("Error processing pending orders: %v\n", err)
			}
		}
	}
}

func (s *AccrualRestService) processPendingOrders(ctx context.Context) error {
	orders, err := s.storage.GetPendingOrders(ctx)
	if err != nil {
		return fmt.Errorf("get pending orders: %w", err)
	}

	for _, order := range orders {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err = s.updateOrderStatus(ctx, order.Number); err != nil {
				fmt.Printf("Error updating order %s: %v\n", order.Number, err)
				continue
			}

			time.Sleep(100 * time.Millisecond)
		}
	}

	return nil
}

func (s *AccrualRestService) updateOrderStatus(ctx context.Context, orderNumber string) error {
	accrualResp, err := s.getOrderStatus(ctx, orderNumber)
	if err != nil {
		return err // TODO: handle rate limit error
	}

	if accrualResp == nil { // TODO: move to orderService; check current order status (?)
		return nil
	}

	order := model.Order{
		Number:  accrualResp.Order,
		Status:  accrualResp.Status,
		Accrual: accrualResp.Accrual,
	}

	return s.storage.UpdateOrderWithBalance(ctx, order)
}

func (s *AccrualRestService) getOrderStatus(ctx context.Context, orderNumber string) (*model.AccrualResponse, error) {
	var accrualResp model.AccrualResponse
	resp, err := s.client.R().SetContext(ctx).SetResult(&accrualResp).Get("/api/orders/" + orderNumber)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}

	statusCode := resp.StatusCode()
	switch statusCode {
	case http.StatusOK:
		return &accrualResp, nil

	case http.StatusNoContent:
		return nil, nil

	case http.StatusTooManyRequests:
		retryAfter := resp.Header().Get("Retry-After")
		return nil, &RateLimitError{RetryAfter: retryAfter}

	case http.StatusInternalServerError:
		return nil, fmt.Errorf("accrual service internal error")

	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

type RateLimitError struct {
	RetryAfter string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded, retry after: %s", e.RetryAfter)
}
