package user

import (
	"context"
	"errors"
	"strings"

	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/errorsx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Create(ctx context.Context, u *User) (*User, error) {
	if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch {
			case strings.Contains(pgErr.ConstraintName, "email"):
				return nil, errorsx.ErrEmailAlreadyExists
			// other error cases
			default:
				return nil, errorsx.ErrDuplicateKey
			}
		}

		return nil, err
	}

	return u, nil
}

func (r *repository) Update(ctx context.Context, u *User) (*User, error) {
	tx := r.db.WithContext(ctx).Select("name", "email").Save(u)

	if err := tx.Error; err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch {
			case strings.Contains(pgErr.ConstraintName, "email"):
				return nil, errorsx.ErrEmailAlreadyExists
			// other cases
			default:
				return nil, errorsx.ErrDuplicateKey
			}
		}

		return nil, err
	}

	if tx.RowsAffected == 0 {
		return nil, errorsx.ErrNotFound
	}

	return u, nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&User{}, id).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User

	if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorsx.ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var u User

	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorsx.ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *repository) GetMany(ctx context.Context) ([]User, error) {
	var users []User
	err := r.db.WithContext(ctx).Find(&users).Error
	return users, err
}
