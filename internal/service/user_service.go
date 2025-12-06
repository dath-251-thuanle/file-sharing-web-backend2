package service

import (
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/domain"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/repository"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
)

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		userRepo: repo,
	}
}

func (us *userService) GetUserById(id string) (*domain.UserResponse, *utils.ReturnStatus) {
	user := &domain.User{}
	err := us.userRepo.FindById(id, user)
	if err != nil {
		return nil, err
	}
	resp := &domain.UserResponse{
		Id:         user.Id,
		Username:   user.Username,
		Email:      user.Email,
		Role:       user.Role,
		EnableTOTP: user.EnableTOTP,
	}
	return resp, nil
}

func (us *userService) GetUserByEmail(email string) (*domain.UserResponse, *utils.ReturnStatus) {
	user := &domain.User{}
	err := us.userRepo.FindByEmail(email, user)
	if err != nil {
		return nil, err
	}
	resp := &domain.UserResponse{
		Id:         user.Id,
		Username:   user.Username,
		Email:      user.Email,
		Role:       user.Role,
		EnableTOTP: user.EnableTOTP,
	}
	return resp, nil
}
