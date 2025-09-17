package model

type DbUser struct {
	ID           int    `json:"id"`
	Login        string `json:"login"`
	PasswordHash string `json:"password_hash"`
}
