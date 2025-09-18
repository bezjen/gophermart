package service

import (
	"context"
	"errors"
	"github.com/bezjen/gophermart/internal/mocks"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"

	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/model"
	"github.com/bezjen/gophermart/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestJWTAuthorizer_Register(t *testing.T) {
	tests := []struct {
		name          string
		login         string
		password      string
		mockSetup     func(*mocks.Repository)
		expectedError error
	}{
		{
			name:     "successful registration",
			login:    "testuser",
			password: "password123",
			mockSetup: func(mockRepo *mocks.Repository) {
				mockRepo.On("GetUserByLogin", mock.Anything, "testuser").Return(nil, nil)
				mockRepo.On("CreateUser", mock.Anything, "testuser", mock.Anything).Return(1, nil)
			},
			expectedError: nil,
		},
		{
			name:     "empty credentials",
			login:    "",
			password: "",
			mockSetup: func(mockRepo *mocks.Repository) {
			},
			expectedError: ErrEmptyCredentials,
		},
		{
			name:     "user already exists",
			login:    "existinguser",
			password: "password123",
			mockSetup: func(mockRepo *mocks.Repository) {
				mockRepo.On("GetUserByLogin", mock.Anything, "existinguser").Return(nil, repository.ErrUserExists)
			},
			expectedError: repository.ErrUserExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.Repository{}
			testLogger, _ := logger.NewLogger("debug")
			authorizer := NewAuthorizer([]byte("secret"), mockRepo, testLogger)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			token, err := authorizer.Register(context.Background(), tt.login, tt.password)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestJWTAuthorizer_Login(t *testing.T) {
	tests := []struct {
		name          string
		login         string
		password      string
		mockSetup     func(*mocks.Repository)
		expectedError error
	}{
		{
			name:     "successful login",
			login:    "testuser",
			password: "correctpassword",
			mockSetup: func(mockRepo *mocks.Repository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), 14)
				user := &model.DbUser{ID: 1, Login: "testuser", PasswordHash: string(hashedPassword)}
				mockRepo.On("GetUserByLogin", mock.Anything, "testuser").Return(user, nil)
			},
			expectedError: nil,
		},
		{
			name:     "invalid credentials - wrong password",
			login:    "testuser",
			password: "wrongpassword",
			mockSetup: func(mockRepo *mocks.Repository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), 14)
				user := &model.DbUser{ID: 1, Login: "testuser", PasswordHash: string(hashedPassword)}
				mockRepo.On("GetUserByLogin", mock.Anything, "testuser").Return(user, nil)
			},
			expectedError: ErrInvalidCredentials,
		},
		{
			name:     "user not found",
			login:    "nonexistent",
			password: "password",
			mockSetup: func(mockRepo *mocks.Repository) {
				mockRepo.On("GetUserByLogin", mock.Anything, "nonexistent").Return((*model.DbUser)(nil), errors.New("user not found"))
			},
			expectedError: ErrInvalidCredentials,
		},
		{
			name:     "database error on get user",
			login:    "testuser",
			password: "password",
			mockSetup: func(mockRepo *mocks.Repository) {
				mockRepo.On("GetUserByLogin", mock.Anything, "testuser").Return((*model.DbUser)(nil), errors.New("db connection error"))
			},
			expectedError: ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.Repository{}
			testLogger, _ := logger.NewLogger("debug")
			authorizer := NewAuthorizer([]byte("secret"), mockRepo, testLogger)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			token, err := authorizer.Login(context.Background(), tt.login, tt.password)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestJWTAuthorizer_ParseToken(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")

	tests := []struct {
		name           string
		secretKey      []byte
		tokenSetup     func(*JWTAuthorizer) string
		expectedUserID int
		expectedError  error
	}{
		{
			name:      "valid token",
			secretKey: []byte("secret"),
			tokenSetup: func(a *JWTAuthorizer) string {
				token, _ := a.generateToken(123)
				return token
			},
			expectedUserID: 123,
			expectedError:  nil,
		},
		{
			name:      "invalid token - wrong signature",
			secretKey: []byte("secret"),
			tokenSetup: func(a *JWTAuthorizer) string {
				wrongAuthorizer := NewAuthorizer([]byte("wrong-secret"), nil, testLogger)
				token, _ := wrongAuthorizer.generateToken(123)
				return token
			},
			expectedUserID: 0,
			expectedError:  ErrParse,
		},
		{
			name:      "expired token",
			secretKey: []byte("secret"),
			tokenSetup: func(a *JWTAuthorizer) string {
				expirationTime := time.Now().Add(-1 * time.Hour)
				claims := &gophermartClaims{
					UserID: 123,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(expirationTime),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("secret"))
				return tokenString
			},
			expectedUserID: 0,
			expectedError:  ErrParse,
		},
		{
			name:           "malformed token",
			secretKey:      []byte("secret"),
			tokenSetup:     func(a *JWTAuthorizer) string { return "malformed.token.string" },
			expectedUserID: 0,
			expectedError:  ErrParse,
		},
		{
			name:           "empty token",
			secretKey:      []byte("secret"),
			tokenSetup:     func(a *JWTAuthorizer) string { return "" },
			expectedUserID: 0,
			expectedError:  ErrParse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authorizer := NewAuthorizer(tt.secretKey, nil, testLogger)
			tokenString := tt.tokenSetup(authorizer)

			userID, err := authorizer.ParseToken(tokenString)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedUserID, userID)
		})
	}
}
