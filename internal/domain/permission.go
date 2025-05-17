package domain

import "time"

type PermissionID string

type Permission struct {
	ID          PermissionID `json:"id"`
	DisplayName string       `json:"displayName"`
	Description string       `json:"description,omitempty"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}
