package domain

type User struct {
	Id         string `json:"user_id"`
	Username   string `json:"user_name" `
	Password   string `json:"password" `
	Email      string `json:"email" `
	Role       string `json:"role"`
	EnableTOTP bool   `json:"enableTOTP"`
	SecretTOTP string `json:"secretTOTP"`
}

type UserCreate struct {
	Username   string `json:"user_name" binding:"required"`
	Email      string `json:"email" binding:"required"`
	Password   string `json:"password" binding:"required"`
	EnableTOTP bool   `json:"enableTOTP"`
	Role       string `json:"role" binding:"required"`
}
