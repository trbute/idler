package api

import (
	"testing"

	"github.com/trbute/idler/server/internal/validation"
)

func TestCharacterValidationIntegration(t *testing.T) {
	tests := []struct {
		name          string
		characterName string
		target        string
		amount        int
		hasAmount     bool
		wantErr       bool
		expectedError string
	}{
		{
			name:          "valid character and target",
			characterName: "testchar",
			target:        "STICKS",
			hasAmount:     false,
			wantErr:       false,
		},
		{
			name:          "valid with amount",
			characterName: "testchar",
			target:        "BALSA TREE",
			amount:        10,
			hasAmount:     true,
			wantErr:       false,
		},
		{
			name:          "invalid character name",
			characterName: "test-char",
			target:        "STICKS",
			hasAmount:     false,
			wantErr:       true,
			expectedError: "name can only contain letters, numbers, and underscores",
		},
		{
			name:          "character name too short",
			characterName: "ab",
			target:        "STICKS",
			hasAmount:     false,
			wantErr:       true,
			expectedError: "name must be at least 3 characters",
		},
		{
			name:          "invalid target format",
			characterName: "testchar",
			target:        "sticks",
			hasAmount:     false,
			wantErr:       true,
			expectedError: "target is invalid",
		},
		{
			name:          "invalid amount",
			characterName: "testchar",
			target:        "STICKS",
			amount:        -5,
			hasAmount:     true,
			wantErr:       true,
			expectedError: "amount must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validation.ValidateCharacterName(tt.characterName); err != nil {
				if !tt.wantErr {
					t.Errorf("Unexpected character name validation error: %v", err)
					return
				}
				if err.Error() != tt.expectedError {
					t.Errorf("Expected error %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err := validation.ValidateTarget(tt.target); err != nil {
				if !tt.wantErr {
					t.Errorf("Unexpected target validation error: %v", err)
					return
				}
				if err.Error() != tt.expectedError {
					t.Errorf("Expected error %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if tt.hasAmount {
				if err := validation.ValidateAmount(&tt.amount); err != nil {
					if !tt.wantErr {
						t.Errorf("Unexpected amount validation error: %v", err)
						return
					}
					if err.Error() != tt.expectedError {
						t.Errorf("Expected error %q, got %q", tt.expectedError, err.Error())
					}
					return
				}
			}

			if tt.wantErr {
				t.Errorf("Expected validation error but got none")
			}
		})
	}
}