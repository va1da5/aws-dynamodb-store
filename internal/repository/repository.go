package repository

import (
	"aws-dynamodb-store/internal/domain"
	"context"
	"errors"
)

var (
	ErrNotFound      = errors.New("entity not found")
	ErrAlreadyExists = errors.New("entity already exists")
	// Add other common repository errors
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	// UpdateUser(ctx context.Context, user *domain.User) error // For user metadata
	// DeleteUser(ctx context.Context, id domain.UserID) error  // Deletes user and their role assignments
	ListAllUsers(ctx context.Context) ([]*domain.User, error)

	AssignRoleToUser(ctx context.Context, userID domain.UserID, roleID domain.RoleID) error
	RemoveRoleFromUser(ctx context.Context, userID domain.UserID, roleID domain.RoleID) error
	GetUserRoles(ctx context.Context, userID domain.UserID) ([]*domain.Role, error)
	ListUsersInRole(ctx context.Context, roleID domain.RoleID) ([]*domain.User, error)
}

type RoleRepository interface {
	CreateRole(ctx context.Context, role *domain.Role) error
	GetRoleByID(ctx context.Context, id domain.RoleID) (*domain.Role, error)
	// UpdateRole(ctx context.Context, role *domain.Role) error // For role metadata
	// DeleteRole(ctx context.Context, id domain.RoleID) error  // Deletes role and its permission assignments, and unassigns from users

	AssignPermissionToRole(ctx context.Context, roleID domain.RoleID, permissionID domain.PermissionID) error
	// RemovePermissionFromRole(ctx context.Context, roleID domain.RoleID, permissionID domain.PermissionID) error
	GetRolePermissions(ctx context.Context, roleID domain.RoleID) ([]*domain.Permission, error)
	// ListRolesWithPermission(ctx context.Context, permissionID domain.PermissionID) ([]*domain.Role, error)
	ListAllRoles(ctx context.Context) ([]*domain.Role, error)
}

type PermissionRepository interface {
	CreatePermission(ctx context.Context, permission *domain.Permission) error
	GetPermissionByID(ctx context.Context, id domain.PermissionID) (*domain.Permission, error)
	// UpdatePermission(ctx context.Context, permission *domain.Permission) error // For permission metadata
	// DeletePermission(ctx context.Context, id domain.PermissionID) error        // Deletes permission and unassigns from roles
	ListAllPermissions(ctx context.Context) ([]*domain.Permission, error)
}

type Repository struct {
	User       UserRepository
	Role       RoleRepository
	Permission PermissionRepository
}

// type Repository interface {
// 	// CreateUser(context.Context, model.User) error
// 	GetUser(context.Context, string) (*model.User, error)
// 	AssignRoleToUser(ctx context.Context, userID, roleID string) error
// 	GetUserRoles(ctx context.Context, userID string) ([]string, error)
// 	GetRolePermissions(ctx context.Context, userID string) ([]string, error)
// }
