package domain

type User struct {
	Id         string `json:"id"`
	Username   string `json:"username" `
	Password   string `json:"password" `
	Email      string `json:"email" `
	Role       string `json:"role"`
	EnableTOTP bool   `json:"enableTOTP"`
	SecretTOTP string `json:"secretTOTP"`
}

type UserCreate struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	Id         string `json:"id"`
	Username   string `json:"username" `
	Email      string `json:"email" `
	Role       string `json:"role"`
	EnableTOTP bool   `json:"totpEnabled"`
}

type UsersLoginSession struct {
	Id			string `json:"id"`
	Cid			string `json:"cid"`
}
