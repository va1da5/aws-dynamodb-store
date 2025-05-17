package domain

import "time"

type RoleID string

type Role struct {
	ID          RoleID    `json:"id"`
	DisplayName string    `json:"displayName"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
