package domain

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginTOTPInput struct {
	ID       string `json:"id" binding:"required"`
	TOTPCode string `json:"code" binding:"required"`
}

type LoginResponse struct {
	AccessToken string    `json:"accessToken,omitempty"`
	ExpiresIn   int       `json:"expiresIn,omitempty"`
	User        *UserInfo `json:"user,omitempty"`
	RequireTOTP bool      `json:"requireTOTP,omitempty"`
	Message     string    `json:"message,omitempty"`
}

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
