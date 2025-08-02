package validation

import (
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
		errMsg  string
	}{
		{"valid email", "test@example.com", false, ""},
		{"valid email with subdomain", "user@mail.example.com", false, ""},
		{"empty email", "", true, "email is required"},
		{"whitespace only", "   ", true, "email is required"},
		{"no @ symbol", "testexample.com", true, "email is invalid"},
		{"no domain", "test@", true, "email is invalid"},
		{"no local part", "@example.com", true, "email is invalid"},
		{"no dot in domain", "test@example", true, "email is invalid"},
		{"multiple @", "test@@example.com", true, "email is invalid"},
		{"invalid format", "not-an-email", true, "email is invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("ValidateEmail() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{"valid password", "ValidPass123", false, ""},
		{"valid with special chars", "ValidPass123!", false, ""},
		{"empty password", "", true, "password is required"},
		{"too short", "Ab1", true, "password must be at least 8 characters"},
		{"no uppercase", "validpass123", true, "password must contain at least one uppercase letter, one lowercase letter, and one number"},
		{"no lowercase", "VALIDPASS123", true, "password must contain at least one uppercase letter, one lowercase letter, and one number"},
		{"no number", "ValidPassword", true, "password must contain at least one uppercase letter, one lowercase letter, and one number"},
		{"too long", "ThisPasswordIsWayTooLongAndExceedsTheMaximumAllowedLengthOf72Characters123", true, "password must be at most 72 characters"},
		{"exactly 72 chars", "ThisPasswordIsExactly72CharactersLongAndShouldBeAcceptedByValidat1", false, ""},
		{"exactly 8 chars", "ValidP1a", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("ValidatePassword() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateSurname(t *testing.T) {
	tests := []struct {
		name    string
		surname string
		wantErr bool
		errMsg  string
	}{
		{"valid surname", "testuser", false, ""},
		{"valid with numbers", "user123", false, ""},
		{"valid with underscore", "test_user", false, ""},
		{"empty surname", "", true, "surname is required"},
		{"whitespace only", "   ", true, "surname is required"},
		{"too short", "ab", true, "surname must be at least 3 characters"},
		{"too long", "thisusernameiswaytoolong", true, "surname must be at most 20 characters"},
		{"invalid characters", "test-user", true, "surname can only contain letters, numbers, and underscores"},
		{"with spaces", "test user", true, "surname can only contain letters, numbers, and underscores"},
		{"with special chars", "test@user", true, "surname can only contain letters, numbers, and underscores"},
		{"exactly 3 chars", "abc", false, ""},
		{"exactly 20 chars", "abcdefghijklmnopqrst", false, ""},
		{"trimmed whitespace", "  testuser  ", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSurname(tt.surname)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSurname() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("ValidateSurname() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateCharacterName(t *testing.T) {
	tests := []struct {
		name          string
		characterName string
		wantErr       bool
		errMsg        string
	}{
		{"valid name", "testchar", false, ""},
		{"valid with numbers", "char123", false, ""},
		{"valid with underscore", "test_char", false, ""},
		{"empty name", "", true, "name is required"},
		{"whitespace only", "   ", true, "name is required"},
		{"too short", "ab", true, "name must be at least 3 characters"},
		{"too long", "thischaracternameistoolong", true, "name must be at most 20 characters"},
		{"invalid characters", "test-char", true, "name can only contain letters, numbers, and underscores"},
		{"with spaces", "test char", true, "name can only contain letters, numbers, and underscores"},
		{"exactly 3 chars", "abc", false, ""},
		{"exactly 20 chars", "abcdefghijklmnopqrst", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCharacterName(tt.characterName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCharacterName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("ValidateCharacterName() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateTarget(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		wantErr bool
		errMsg  string
	}{
		{"valid IDLE", "IDLE", false, ""},
		{"valid STICKS", "STICKS", false, ""},
		{"valid ROCKS", "ROCKS", false, ""},
		{"valid BALSA TREE", "BALSA TREE", false, ""},
		{"valid SOAPSTONE DEPOSIT", "SOAPSTONE DEPOSIT", false, ""},
		{"valid with numbers", "BALSA TREE 1", false, ""},
		{"empty target", "", true, "target is required"},
		{"too long", "THISRESOURCENAMEISWAYTOOLONGANDEXCEEDSTHEMAXIMUMLIMIT", true, "target is invalid"},
		{"invalid lowercase", "sticks", true, "target is invalid"},
		{"invalid mixed case", "Balsa Tree", true, "target is invalid"},
		{"invalid characters", "STICKS@ROCKS", true, "target is invalid"},
		{"invalid underscore", "BALSA_TREE", true, "target is invalid"},
		{"invalid dash", "BALSA-TREE", true, "target is invalid"},
		{"exactly 50 chars", "ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWX", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTarget(tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("ValidateTarget() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateAmount(t *testing.T) {
	ten := 10
	zero := 0
	negative := -5

	tests := []struct {
		name    string
		amount  *int
		wantErr bool
		errMsg  string
	}{
		{"nil amount", nil, false, ""},
		{"valid positive", &ten, false, ""},
		{"zero amount", &zero, true, "amount must be greater than 0"},
		{"negative amount", &negative, true, "amount must be greater than 0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAmount(tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("ValidateAmount() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateItemName(t *testing.T) {
	tests := []struct {
		name     string
		itemName string
		wantErr  bool
		errMsg   string
	}{
		{"valid STICKS", "STICKS", false, ""},
		{"valid ROCKS", "ROCKS", false, ""},
		{"valid BALSA LOGS", "BALSA LOGS", false, ""},
		{"valid SOAPSTONE", "SOAPSTONE", false, ""},
		{"valid with numbers", "STICKS 2", false, ""},
		{"empty name", "", true, "item name is required"},
		{"too long", "THISITEMNAMEISWAYTOOLONGANDEXCEEDSTHEMAXIMUMLIMITOFCHARS", true, "item name is invalid"},
		{"invalid lowercase", "sticks", true, "item name is invalid"},
		{"invalid mixed case", "Balsa Logs", true, "item name is invalid"},
		{"invalid characters", "STICKS@ROCKS", true, "item name is invalid"},
		{"invalid underscore", "BALSA_LOGS", true, "item name is invalid"},
		{"invalid dash", "BALSA-LOGS", true, "item name is invalid"},
		{"exactly 50 chars", "ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWX", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateItemName(tt.itemName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateItemName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("ValidateItemName() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateQuantity(t *testing.T) {
	tests := []struct {
		name     string
		quantity int
		wantErr  bool
		errMsg   string
	}{
		{"valid quantity", 10, false, ""},
		{"zero quantity", 0, true, "quantity must be greater than 0"},
		{"negative quantity", -5, true, "quantity must be greater than 0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQuantity(tt.quantity)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQuantity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("ValidateQuantity() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{"normal string", "hello", 10, "hello"},
		{"with whitespace", "  hello  ", 10, "hello"},
		{"too long", "this is a very long string", 10, "this is a "},
		{"exact length", "exactly10c", 10, "exactly10c"},
		{"empty string", "", 10, ""},
		{"whitespace only", "   ", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input, tt.maxLength)
			if result != tt.expected {
				t.Errorf("SanitizeString() = %v, want %v", result, tt.expected)
			}
		})
	}
}