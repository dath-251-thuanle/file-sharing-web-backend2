package test

import (
	"fmt"
	"testing"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/infrastructure/database"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/infrastructure/jwt"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/repository"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/service"
)

func TestCreateUser(t *testing.T) {
	if err := database.InitDB(); err != nil {
		t.Errorf("Failed: %v", err)
		return
	}

	userrepo := repository.NewSQLUserRepository(database.DB)
	auth := service.NewAuthService(userrepo, repository.NewAuthRepository(database.DB), jwt.NewJWTService())
	serv := service.NewUserService(userrepo)

	if _, err := auth.CreateUser("kdm", "kdm12345", "kdm@gmail.com"); err != nil {
		t.Errorf(`FAILED: %v`, err)
		return
	}

	user, err := serv.GetUserByEmail("kdm@gmail.com")

	if err != nil {
		t.Errorf(`FAILED: %v`, err)
		return
	}

	fmt.Println(user)
}
