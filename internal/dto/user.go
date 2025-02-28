package dto

import "github.com/samcharles93/cinea/internal/entity"

type UserDTO struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

func UserToDTO(user *entity.User) *UserDTO {
	return &UserDTO{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}
}
