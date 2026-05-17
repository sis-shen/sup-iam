package service

import (
	"golang.org/x/crypto/bcrypt"
)

type PasswordHasherInterface interface {
	HashPassword(password string) (string, error)
	VerifyPassword(plainPassword string, hashPassword string) error
}

type InnerBcryptPasswordHasher struct {
	cost int
}

func NewInnerBcryptPasswordHasher(cost int) *InnerBcryptPasswordHasher {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return &InnerBcryptPasswordHasher{
		cost: cost,
	}
}

func (h *InnerBcryptPasswordHasher) HashPassword(plainPassword string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), h.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (h *InnerBcryptPasswordHasher) VerifyPassword(plainPassword string, hashPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(plainPassword))
}
