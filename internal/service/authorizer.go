//go:generate mockery --name=Authorizer --output=../mocks --case=underscore
package service

import (
	"context"
	"errors"
	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var ErrParse = errors.New("failed to parse token")

type Authorizer interface {
	Register(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, error)
	ParseToken(tokenString string) (int, error)
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

func (a *JWTAuthorizer) ParseToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &gophermartClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			a.logger.Error("Unexpected signing method")
			return nil, ErrParse
		}
		return a.secretKey, nil
	})

	if err != nil {
		a.logger.Error("Failed to parse token", zap.Error(err), zap.String("token", tokenString))
		return 0, ErrParse
	}

	if claims, ok := token.Claims.(*gophermartClaims); ok && token.Valid {
		return claims.UserID, nil
	}

	a.logger.Error("Invalid token", zap.String("token", tokenString))
	return 0, ErrParse
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
