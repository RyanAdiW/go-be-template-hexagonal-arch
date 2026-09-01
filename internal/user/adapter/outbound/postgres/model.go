package postgres

import (
	"github.com/google/uuid"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/domain"
	"time"
)

// userModel belongs exclusively to the PostgreSQL adapter. Domain entities remain ORM-free.
type userModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email     string    `gorm:"not null"`
	Name      string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (userModel) TableName() string { return "users" }
func toModel(user domain.User) userModel {
	return userModel{ID: user.ID, Email: user.Email, Name: user.Name, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt}
}
func toDomain(model userModel) domain.User {
	return domain.User{ID: model.ID, Email: model.Email, Name: model.Name, CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt}
}
