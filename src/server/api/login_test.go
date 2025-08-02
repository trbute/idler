package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleLogin_Validation(t *testing.T) {
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
				"password": "password",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email is invalid",
		},
		{
			name: "empty email",
			requestBody: map[string]string{
				"email":    "",
				"password": "password",
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
		{
			name: "whitespace only email",
			requestBody: map[string]string{
				"email":    "   ",
				"password": "password",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email is required",
		},
		{
			name: "no @ symbol",
			requestBody: map[string]string{
				"email":    "testexample.com",
				"password": "password",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email is invalid",
		},
		{
			name: "malformed JSON",
			requestBody: "invalid json",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Unable to decode parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			cfg.handleLogin(w, req)

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