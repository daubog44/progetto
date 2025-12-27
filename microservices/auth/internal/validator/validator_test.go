package validator

import (
	"testing"
)

func TestValidateRegister(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		username string
		wantErr  bool
	}{
		{
			name:     "valid request",
			email:    "test@example.com",
			password: "Password123!",
			username: "testuser",
			wantErr:  false,
		},
		{
			name:     "invalid email",
			email:    "invalid-email",
			password: "Password123!",
			username: "testuser",
			wantErr:  true,
		},
		{
			name:     "short password",
			email:    "test@example.com",
			password: "Pass1!",
			username: "testuser",
			wantErr:  true,
		},
		{
			name:     "no uppercase in password",
			email:    "test@example.com",
			password: "password123!",
			username: "testuser",
			wantErr:  true,
		},
		{
			name:     "no digit in password",
			email:    "test@example.com",
			password: "Password!",
			username: "testuser",
			wantErr:  true,
		},
		{
			name:     "no special char in password",
			email:    "test@example.com",
			password: "Password123",
			username: "testuser",
			wantErr:  true,
		},
		{
			name:     "short username",
			email:    "test@example.com",
			password: "Password123!",
			username: "tu",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateRegister(tt.email, tt.password, tt.username); (err != nil) != tt.wantErr {
				t.Errorf("ValidateRegister() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
