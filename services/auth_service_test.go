package services

import (
	"errors"
	"testing"

	"gobase-app/models"
	"gobase-app/repositories"

	"golang.org/x/crypto/bcrypt"
)

type authAccountFinderStub struct {
	account models.AuthAccount
	err     error
}

func (s authAccountFinderStub) FindActiveByUsername(string) (models.AuthAccount, error) {
	return s.account, s.err
}

func TestAuthServiceAuthenticate(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("generate password hash: %v", err)
	}

	tests := []struct {
		name       string
		repo       AuthAccountFinder
		username   string
		password   string
		wantErr    error
		wantUserID int
	}{
		{
			name: "valid credentials",
			repo: authAccountFinderStub{account: models.AuthAccount{
				UserID: 7, Username: "requester", Name: "Demo Requester", PasswordHash: string(passwordHash),
			}},
			username: " requester ", password: "correct-password", wantUserID: 7,
		},
		{name: "username required", repo: authAccountFinderStub{}, password: "secret", wantErr: ErrUsernameRequired},
		{name: "password required", repo: authAccountFinderStub{}, username: "requester", wantErr: ErrPasswordRequired},
		{
			name: "account missing", repo: authAccountFinderStub{err: repositories.ErrAuthAccountNotFound},
			username: "missing", password: "secret", wantErr: ErrInactiveOrMissing,
		},
		{
			name: "invalid password", repo: authAccountFinderStub{account: models.AuthAccount{PasswordHash: string(passwordHash)}},
			username: "requester", password: "wrong-password", wantErr: ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := AuthService{Repo: tt.repo}
			user, err := service.Authenticate(tt.username, tt.password)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Authenticate() error = %v, want %v", err, tt.wantErr)
			}
			if user.UserID != tt.wantUserID {
				t.Fatalf("Authenticate() user ID = %d, want %d", user.UserID, tt.wantUserID)
			}
		})
	}
}
