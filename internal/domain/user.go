package domain

type User struct {
	Id         string `json:"userid"`
	Username   string `json:"username" `
	Password   string `json:"password" `
	Email      string `json:"email" `
	Role       string `json:"role"`
	EnableTOTP bool   `json:"enableTOTP"`
	SecretTOTP string `json:"secretTOTP"`
}

type UserCreate struct {
	Username   string `json:"username" binding:"required"`
	Email      string `json:"email" binding:"required"`
	Password   string `json:"password" binding:"required"`
}
