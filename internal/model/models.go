package model

import "time"

type APIUser struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type DBUser struct {
	ID           int    `json:"id"`
	Login        string `json:"login"`
	PasswordHash string `json:"password_hash"`
}

type Order struct {
	ID         int         `json:"-"`
	Number     string      `json:"number"`
	Status     OrderStatus `json:"status"`
	Accrual    float64     `json:"accrual,omitempty"`
	UploadedAt time.Time   `json:"uploaded_at"`
	UserID     int         `json:"-"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Withdrawal struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
}

type AccrualResponse struct {
	Order   string             `json:"order"`
	Status  AccrualOrderStatus `json:"status"`
	Accrual float64            `json:"accrual"`
}

type AccrualOrderStatus string

const (
	AccrualOrderStatusRegistered AccrualOrderStatus = "REGISTERED"
	AccrualOrderStatusProcessing AccrualOrderStatus = "PROCESSING"
	AccrualOrderStatusInvalid    AccrualOrderStatus = "INVALID"
	AccrualOrderStatusProcessed  AccrualOrderStatus = "PROCESSED"
)

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)
