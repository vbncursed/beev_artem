package auth_service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"
)

type ValidateSuite struct {
	suite.Suite
}

func (s *ValidateSuite) TestValidateEmail_Valid() {
	tests := []struct {
		name  string
		email string
	}{
		{
			name:  "valid email",
			email: "user@example.com",
		},
		{
			name:  "valid email with subdomain",
			email: "user@mail.example.com",
		},
		{
			name:  "valid email with plus",
			email: "user+tag@example.com",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := validateEmail(tt.email)
			assert.NilError(s.T(), err)
		})
	}
}

func (s *ValidateSuite) TestValidateEmail_Invalid() {
	tests := []struct {
		name  string
		email string
	}{
		{
			name:  "empty email",
			email: "",
		},
		{
			name:  "email too short",
			email: "a@b",
		},
		{
			name:  "email too long",
			email: strings.Repeat("a", 250) + "@example.com",
		},
		{
			name:  "email without @",
			email: "userexample.com",
		},
		{
			name:  "email with multiple @",
			email: "user@@example.com",
		},
		{
			name:  "email without domain",
			email: "user@",
		},
		{
			name:  "email without local part",
			email: "@example.com",
		},
		{
			name:  "email with invalid format",
			email: "user @example.com",
		},
		{
			name:  "email with domain too long",
			email: "user@" + strings.Repeat("a", 254) + ".com",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := validateEmail(tt.email)
			assert.ErrorIs(s.T(), err, ErrInvalidEmail)
		})
	}
}

func (s *ValidateSuite) TestValidatePassword_Valid() {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "valid password",
			password: "Password123!",
		},
		{
			name:     "valid password with symbol",
			password: "Password123@",
		},
		{
			name:     "valid password with punctuation",
			password: "Password123.",
		},
		{
			name:     "password with unicode special",
			password: "Password123€",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := validatePassword(tt.password)
			assert.NilError(s.T(), err)
		})
	}
}

func (s *ValidateSuite) TestValidatePassword_Invalid() {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "password too short",
			password: "Pass1!",
		},
		{
			name:     "password without uppercase",
			password: "password123!",
		},
		{
			name:     "password without lowercase",
			password: "PASSWORD123!",
		},
		{
			name:     "password without digit",
			password: "Password!",
		},
		{
			name:     "password without special character",
			password: "Password123",
		},
		{
			name:     "password with only uppercase and digit",
			password: "PASSWORD123",
		},
		{
			name:     "password with only lowercase and digit",
			password: "password123",
		},
		{
			name:     "password with only letters and special",
			password: "Password!",
		},
		{
			name:     "password with all requirements but too short",
			password: "Pass1!",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := validatePassword(tt.password)
			assert.ErrorIs(s.T(), err, ErrInvalidPassword)
		})
	}
}

func (s *ValidateSuite) TestValidateAuthInput_Valid() {
	err := validateAuthInput("user@example.com", "Password123!")
	assert.NilError(s.T(), err)
}

func (s *ValidateSuite) TestValidateAuthInput_InvalidEmail() {
	err := validateAuthInput("invalid-email", "Password123!")
	assert.ErrorIs(s.T(), err, ErrInvalidEmail)
}

func (s *ValidateSuite) TestValidateAuthInput_InvalidPassword() {
	err := validateAuthInput("user@example.com", "weak")
	assert.ErrorIs(s.T(), err, ErrInvalidPassword)
}

func (s *ValidateSuite) TestValidateAuthInput_BothInvalid() {
	err := validateAuthInput("invalid-email", "weak")
	assert.ErrorIs(s.T(), err, ErrInvalidEmail)
}

func TestValidateSuite(t *testing.T) {
	suite.Run(t, new(ValidateSuite))
}
