package model

type UserCreateInput struct {
	DisplayName string `json:"name"`
	Email       string `json:"email" validate:"required,email"`
	// Password    string `json:"password" validate:"required,min=8"` // Plain text password from client
}

type UserResponse struct {
	ID          string `json:"id"` // EntityID
	DisplayName string `json:"displayName,omitempty"`
	Email       string `json:"email"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}
