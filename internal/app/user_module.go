package app

import (
	userhttp "github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/adapter/inbound/http"
	users "github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/adapter/outbound/db/users"
	"github.com/ryanadi/go-be-template-hexagonal-arch/internal/user/application"
	"gopkg.in/go-playground/validator.v9"
	"gorm.io/gorm"
)

func buildUserModule(db *gorm.DB, validate *validator.Validate) *userhttp.UserModule {
	service := application.NewUserService(application.Config{
		UserRepository: users.NewUsersDB(users.Config{Client: db}),
	})

	return userhttp.NewModule(userhttp.Dependencies{
		UserService: service,
		Validator:   validate,
	})
}
