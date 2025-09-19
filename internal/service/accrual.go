package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/bezjen/gophermart/internal/model"
	"github.com/go-resty/resty/v2"
	"net/http"
	"strconv"
	"time"
)

type AccrualService interface {
	StartOrderProcessingWorker(ctx context.Context, interval time.Duration)
}

type AccrualRestService struct {
	client         *resty.Client
	accrualBaseURL string
	orderService   OrderService
}

func NewAccrualRestService(accrualBaseURL string, orderService OrderService) *AccrualRestService {
	return &AccrualRestService{
		client:         resty.New().SetBaseURL(accrualBaseURL),
		accrualBaseURL: accrualBaseURL,
		orderService:   orderService,
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
	orders, err := s.orderService.GetPendingOrders(ctx)
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
		if errors.Is(err, &RateLimitError{}) {
			var rateLimitErr *RateLimitError
			if errors.As(err, &rateLimitErr) {
				time.Sleep(time.Duration(rateLimitErr.RetryAfter) * time.Second)
			}
			return nil
		}
	}

	if accrualResp == nil {
		return nil
	}

	order := model.Order{
		Number:  accrualResp.Order,
		Status:  accrualResp.Status,
		Accrual: accrualResp.Accrual,
	}

	return s.orderService.UpdateOrderWithBalance(ctx, order)
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
		retryAfter, err := strconv.Atoi(resp.Header().Get("Retry-After"))
		if err != nil {
			return nil, err
		}
		return nil, &RateLimitError{RetryAfter: retryAfter}

	case http.StatusInternalServerError:
		return nil, fmt.Errorf("accrual service internal error")

	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

type RateLimitError struct {
	RetryAfter int
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded, retry after: %s", e.RetryAfter)
}
