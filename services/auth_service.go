package services

import (
	"errors"
	"strings"

	helpers "gobase-app/helper"
	"gobase-app/models"
	"gobase-app/repositories"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameRequired  = errors.New("username wajib diisi")
	ErrPasswordRequired  = errors.New("password wajib diisi")
	ErrInactiveOrMissing = errors.New("username tidak ditemukan / atau mungkin user tidak aktif")
	ErrInvalidPassword   = errors.New("password salah")
)

type AuthAccountFinder interface {
	FindActiveByUsername(username string) (models.AuthAccount, error)
}

type AuthService struct {
	Repo AuthAccountFinder
}

func (s *AuthService) Authenticate(username, password string) (models.SessionUser, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return models.SessionUser{}, ErrUsernameRequired
	}
	if password == "" {
		return models.SessionUser{}, ErrPasswordRequired
	}

	account, err := s.Repo.FindActiveByUsername(username)
	if errors.Is(err, repositories.ErrAuthAccountNotFound) {
		return models.SessionUser{}, ErrInactiveOrMissing
	}
	if err != nil {
		return models.SessionUser{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)); err != nil {
		return models.SessionUser{}, ErrInvalidPassword
	}

	return models.SessionUser{
		UserID:          account.UserID,
		NIP:             account.NIP,
		Name:            account.Name,
		Initials:        helpers.Initials(account.Name),
		Username:        account.Username,
		Role:            account.Role,
		StoreID:         account.StoreID,
		IsAuthenticated: true,
	}, nil
}
