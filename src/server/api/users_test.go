package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleCreateUser_Validation(t *testing.T) {
	cfg := &ApiConfig{}

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "invalid email",
			requestBody: map[string]string{
				"email":    "invalid-email",
				"password": "ValidPass123",
				"surname":  "testuser",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email is invalid",
		},
		{
			name: "weak password",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "weak",
				"surname":  "testuser",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "password must be at least 8 characters",
		},
		{
			name: "password missing complexity",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "weakpassword",
				"surname":  "testuser",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "password must contain at least one uppercase letter, one lowercase letter, and one number",
		},
		{
			name: "surname too short",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "ValidPass123",
				"surname":  "ab",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "surname must be at least 3 characters",
		},
		{
			name: "surname invalid characters",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "ValidPass123",
				"surname":  "test-user",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "surname can only contain letters, numbers, and underscores",
		},
		{
			name: "empty email",
			requestBody: map[string]string{
				"email":    "",
				"password": "ValidPass123",
				"surname":  "testuser",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email is required",
		},
		{
			name: "empty password",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "",
				"surname":  "testuser",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "password is required",
		},
		{
			name: "empty surname",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "ValidPass123",
				"surname":  "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "surname is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			cfg.handleCreateUser(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]string
			json.Unmarshal(w.Body.Bytes(), &response)

			if response["error"] != tt.expectedError {
				t.Errorf("Expected error message %q, got %q", tt.expectedError, response["error"])
			}
		})
	}
}

func TestHandleUpdateUser_Validation(t *testing.T) {
	cfg := &ApiConfig{}

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "invalid email",
			requestBody: map[string]string{
				"email":    "invalid-email",
				"password": "ValidPass123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email is invalid",
		},
		{
			name: "weak password",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "weak",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "password must be at least 8 characters",
		},
		{
			name: "empty email",
			requestBody: map[string]string{
				"email":    "",
				"password": "ValidPass123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email is required",
		},
		{
			name: "empty password",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "password is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/api/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer fake-token")

			w := httptest.NewRecorder()
			cfg.handleUpdateUser(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]string
			json.Unmarshal(w.Body.Bytes(), &response)

			if response["error"] != tt.expectedError {
				t.Errorf("Expected error message %q, got %q", tt.expectedError, response["error"])
			}
		})
	}
}