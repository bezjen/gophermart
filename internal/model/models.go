package model

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
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
