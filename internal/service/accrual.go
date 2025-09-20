package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/model"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	defaultWorkerCount = 5
)

type AccrualService interface {
	StartOrderProcessingWorker(ctx context.Context, interval time.Duration)
}

type AccrualRestService struct {
	client         *resty.Client
	accrualBaseURL string
	orderService   OrderService
	logger         *logger.Logger
	workerCount    int
	rateLimitUntil time.Time
	rateLimitMutex sync.Mutex
}

func NewAccrualRestService(accrualBaseURL string, orderService OrderService, logger *logger.Logger) *AccrualRestService {
	return &AccrualRestService{
		client:         resty.New().SetBaseURL(accrualBaseURL),
		accrualBaseURL: accrualBaseURL,
		orderService:   orderService,
		logger:         logger,
		workerCount:    defaultWorkerCount,
	}
}

func (s *AccrualRestService) StartOrderProcessingWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.processPendingOrders(ctx); err != nil {
				s.logger.Error("Error processing pending orders", zap.Error(err))
			}
		}
	}
}

func (s *AccrualRestService) processPendingOrders(ctx context.Context) error {
	orders, err := s.orderService.GetPendingOrders(ctx, s.workerCount)
	if err != nil {
		return fmt.Errorf("failed to get pending orders: %w", err)
	}

	if len(orders) == 0 {
		return nil
	}

	jobs := make(chan model.Order, len(orders))
	for _, order := range orders {
		jobs <- order
	}
	close(jobs)

	var wg sync.WaitGroup
	for i := 0; i < s.workerCount; i++ {
		wg.Add(1)
		go s.worker(ctx, &wg, jobs)
	}
	wg.Wait()

	return nil
}

func (s *AccrualRestService) worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan model.Order) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case order, ok := <-jobs:
			if !ok {
				return
			}

			if err := s.waitForRateLimit(ctx); err != nil {
				s.logger.Info("Worker shutting down while waiting for rate limit.", zap.Error(err))
				return
			}

			err := s.updateOrderStatus(ctx, order.Number)
			if err != nil {
				var rateLimitErr *RateLimitError
				if errors.As(err, &rateLimitErr) {
					pauseDuration := time.Duration(rateLimitErr.RetryAfter) * time.Second
					s.acquireRateLimitPause(pauseDuration)
					s.logger.Warn("Accrual service rate limit acquired",
						zap.Int("retryAfter", rateLimitErr.RetryAfter))
				} else {
					s.logger.Error("Failed to update order status",
						zap.Error(err),
						zap.String("order_number", order.Number),
					)
				}
			}
		}
	}
}

func (s *AccrualRestService) acquireRateLimitPause(d time.Duration) {
	s.rateLimitMutex.Lock()
	defer s.rateLimitMutex.Unlock()
	s.rateLimitUntil = time.Now().Add(d)
}

func (s *AccrualRestService) waitForRateLimit(ctx context.Context) error {
	s.rateLimitMutex.Lock()
	waitTime := time.Until(s.rateLimitUntil)
	s.rateLimitMutex.Unlock()

	if waitTime <= 0 {
		return nil
	}

	timer := time.NewTimer(waitTime)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (s *AccrualRestService) updateOrderStatus(ctx context.Context, orderNumber string) error {
	accrualResp, err := s.getOrderStatus(ctx, orderNumber)
	if err != nil {
		return err
	}

	if accrualResp == nil {
		return nil
	}

	order := model.Order{
		Number:  accrualResp.Order,
		Status:  mapAccrualOrderStatus(accrualResp.Status),
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
		return nil, fmt.Errorf("unexpected status code: %d", statusCode)
	}
}

type RateLimitError struct {
	RetryAfter int
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded, retry after: %d", e.RetryAfter)
}

func mapAccrualOrderStatus(accrualOrderStatus model.AccrualOrderStatus) model.OrderStatus {
	switch accrualOrderStatus {
	case "REGISTERED":
		return model.OrderStatusProcessing
	case "PROCESSING":
		return model.OrderStatusProcessing
	case "PROCESSED":
		return model.OrderStatusProcessed
	case "INVALID":
		return model.OrderStatusInvalid
	default:
		return model.OrderStatusNew
	}
}
