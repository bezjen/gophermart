package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/bezjen/gophermart/internal/logger"
	"github.com/bezjen/gophermart/internal/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bezjen/gophermart/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthHandler_HandleRegister(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]string
		mockSetup      func(authorizer *mocks.Authorizer)
		expectedStatus int
		expectedHeader string
	}{
		{
			name: "successful registration",
			requestBody: map[string]string{
				"login":    "testuser",
				"password": "testpass",
			},
			mockSetup: func(m *mocks.Authorizer) {
				m.On("Register", mock.Anything, "testuser", "testpass").
					Return("valid-token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "Bearer valid-token",
		},
		{
			name: "user already exists",
			requestBody: map[string]string{
				"login":    "existinguser",
				"password": "testpass",
			},
			mockSetup: func(m *mocks.Authorizer) {
				m.On("Register", mock.Anything, "existinguser", "testpass").
					Return("", repository.ErrUserExists)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := &mocks.Authorizer{}
			tt.mockSetup(mockAuth)

			testLogger, _ := logger.NewLogger("debug")
			handler := NewAuthHandler(testLogger, mockAuth)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
			rr := httptest.NewRecorder()

			handler.HandleRegister(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedHeader, rr.Header().Get("Authorization"))
			}

			mockAuth.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_HandleLogin(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]string
		mockSetup      func(*mocks.Authorizer)
		expectedStatus int
		expectedHeader string
	}{
		{
			name: "successful login",
			requestBody: map[string]string{
				"login":    "testuser",
				"password": "testpass",
			},
			mockSetup: func(m *mocks.Authorizer) {
				m.On("Login", mock.Anything, "testuser", "testpass").
					Return("valid-token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "Bearer valid-token",
		},
		{
			name: "invalid credentials",
			requestBody: map[string]string{
				"login":    "testuser",
				"password": "wrongpass",
			},
			mockSetup: func(m *mocks.Authorizer) {
				m.On("Login", mock.Anything, "testuser", "wrongpass").
					Return("", errors.New("invalid credentials"))
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := &mocks.Authorizer{}
			tt.mockSetup(mockAuth)

			testLogger, _ := logger.NewLogger("debug")
			handler := NewAuthHandler(testLogger, mockAuth)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
			rr := httptest.NewRecorder()

			handler.HandleLogin(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedHeader, rr.Header().Get("Authorization"))
			}

			mockAuth.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_EdgeCases(t *testing.T) {
	t.Run("empty request body", func(t *testing.T) {
		mockAuth := &mocks.Authorizer{}
		testLogger, _ := logger.NewLogger("debug")
		handler := NewAuthHandler(testLogger, mockAuth)

		req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte{}))
		rr := httptest.NewRecorder()

		handler.HandleRegister(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("malformed JSON", func(t *testing.T) {
		mockAuth := &mocks.Authorizer{}
		testLogger, _ := logger.NewLogger("debug")
		handler := NewAuthHandler(testLogger, mockAuth)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte("{invalid json")))
		rr := httptest.NewRecorder()

		handler.HandleLogin(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestNewAuthHandler(t *testing.T) {
	mockAuth := &mocks.Authorizer{}
	testLogger, _ := logger.NewLogger("debug")

	handler := NewAuthHandler(testLogger, mockAuth)

	assert.NotNil(t, handler)
	assert.Equal(t, testLogger, handler.logger)
	assert.Equal(t, mockAuth, handler.authorizer)
}
