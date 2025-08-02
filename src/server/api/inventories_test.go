package api

import (
	"testing"

	"github.com/trbute/idler/server/internal/validation"
)

func TestInventoryValidationIntegration(t *testing.T) {
	tests := []struct {
		name          string
		characterName string
		itemName      string
		quantity      int
		wantErr       bool
		expectedError string
	}{
		{
			name:          "valid character and item",
			characterName: "testchar",
			itemName:      "STICKS",
			quantity:      5,
			wantErr:       false,
		},
		{
			name:          "valid balsa logs",
			characterName: "testchar",
			itemName:      "BALSA LOGS",
			quantity:      10,
			wantErr:       false,
		},
		{
			name:          "invalid character name",
			characterName: "test-char",
			itemName:      "STICKS",
			quantity:      5,
			wantErr:       true,
			expectedError: "name can only contain letters, numbers, and underscores",
		},
		{
			name:          "character name too short",
			characterName: "ab",
			itemName:      "STICKS",
			quantity:      5,
			wantErr:       true,
			expectedError: "name must be at least 3 characters",
		},
		{
			name:          "empty item name",
			characterName: "testchar",
			itemName:      "",
			quantity:      5,
			wantErr:       true,
			expectedError: "item name is required",
		},
		{
			name:          "invalid item name format (lowercase)",
			characterName: "testchar",
			itemName:      "sticks",
			quantity:      5,
			wantErr:       true,
			expectedError: "item name is invalid",
		},
		{
			name:          "invalid item name format (mixed case)",
			characterName: "testchar",
			itemName:      "Balsa Logs",
			quantity:      5,
			wantErr:       true,
			expectedError: "item name is invalid",
		},
		{
			name:          "invalid item name characters",
			characterName: "testchar",
			itemName:      "STICKS@ROCKS",
			quantity:      5,
			wantErr:       true,
			expectedError: "item name is invalid",
		},
		{
			name:          "invalid quantity",
			characterName: "testchar",
			itemName:      "STICKS",
			quantity:      0,
			wantErr:       true,
			expectedError: "quantity must be greater than 0",
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

			if err := validation.ValidateItemName(tt.itemName); err != nil {
				if !tt.wantErr {
					t.Errorf("Unexpected item name validation error: %v", err)
					return
				}
				if err.Error() != tt.expectedError {
					t.Errorf("Expected error %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err := validation.ValidateQuantity(tt.quantity); err != nil {
				if !tt.wantErr {
					t.Errorf("Unexpected quantity validation error: %v", err)
					return
				}
				if err.Error() != tt.expectedError {
					t.Errorf("Expected error %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if tt.wantErr {
				t.Errorf("Expected validation error but got none")
			}
		})
	}
}