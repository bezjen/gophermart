//go:generate mockery --name=Authorizer --output=../mocks --case=underscore
package service

import (
	"context"
	"errors"
	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Authorizer interface {
	Register(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, error)
}

type JWTAuthorizer struct {
	secretKey []byte
	storage   repository.Repository
	logger    *logger.Logger
}

func NewAuthorizer(secretKey []byte, storage repository.Repository, logger *logger.Logger) *JWTAuthorizer {
	return &JWTAuthorizer{
		secretKey: secretKey,
		storage:   storage,
		logger:    logger,
	}
}

func (a *JWTAuthorizer) Register(ctx context.Context, login string, password string) (string, error) {
	err := validateCredentials(login, password)
	if err != nil {
		return "", err
	}

	_, err = a.storage.GetUserByLogin(ctx, login)
	if err != nil && errors.Is(err, repository.ErrUserExists) {
		return "", err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}

	userID, err := a.storage.CreateUser(ctx, login, string(passwordHash))
	if err != nil {
		return "", err
	}

	return a.generateToken(userID)
}

func (a *JWTAuthorizer) Login(ctx context.Context, login string, password string) (string, error) {
	err := validateCredentials(login, password)
	if err != nil {
		return "", err
	}

	user, err := a.storage.GetUserByLogin(ctx, login)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	return a.generateToken(user.ID)
}

func (a *JWTAuthorizer) generateToken(userID int) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &gophermartClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.secretKey)
}

func validateCredentials(login string, password string) error {
	if login == "" || password == "" {
		return errors.New("login/password cannot be empty") // TODO: add custom error types
	}
	return nil
}

type gophermartClaims struct {
	UserID int `json:"userID"`
	jwt.RegisteredClaims
}
