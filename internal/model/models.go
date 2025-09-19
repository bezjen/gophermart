package model

import "time"

type ApiUser struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type DbUser struct {
	ID           int    `json:"id"`
	Login        string `json:"login"`
	PasswordHash string `json:"password_hash"`
}

type Order struct {
	ID         int       `json:"-"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
	UserID     int       `json:"-"`
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
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}
