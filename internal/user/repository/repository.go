package repository

import (
	"context"
	"lumiere/internal/models"
)

type UserRepo interface {
	Create(ctx context.Context, u *models.User) error
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByID(ctx context.Context, id uint) (*models.User, error)
}
