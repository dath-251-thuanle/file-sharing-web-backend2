package test

import (
	"fmt"
	"testing"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/infrastructure/database"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/repository"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/service"
)

func TestCreateUser(t *testing.T) {
	name := "kdm"
	password := "bitch133"
	email := "kdm@gmail.com"
	role := "user"

	database.InitDB()

	userrepo := repository.NewSQLUserRepository(database.DB)
	service := service.NewUserService(userrepo)

	if err := service.CreateUser(name, password, email, role); err != nil {
		t.Errorf(`FAILED: %v`, err)
		return
	}

	user, err := service.GetUserByEmail(email)

	if err != nil {
		t.Errorf(`FAILED: %v`, err)
		return
	}

	fmt.Println(user)
}
