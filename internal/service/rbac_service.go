package service

import (
	"aws-dynamodb-store/internal/domain"
	"aws-dynamodb-store/internal/repository"
	"context"
	"errors"
	"fmt"
)

// RBACService provides methods for managing users, roles, permissions, and checking access.
type RBACService interface {
	// User Management
	CreateUser(ctx context.Context, displayName, email string) (*domain.User, error)
	GetUser(ctx context.Context, userID domain.UserID) (*domain.User, error)
	AssignRoleToUser(ctx context.Context, userID domain.UserID, roleID domain.RoleID) error
	RemoveRoleFromUser(ctx context.Context, userID domain.UserID, roleID domain.RoleID) error
	GetUserRoles(ctx context.Context, userID domain.UserID) ([]*domain.Role, error)

	// // Role Management
	CreateRole(ctx context.Context, displayName, description string) (*domain.Role, error)
	// GetRole(ctx context.Context, roleID domain.RoleID) (*domain.Role, error)
	AssignPermissionToRole(ctx context.Context, roleID domain.RoleID, permissionID domain.PermissionID) error
	// RemovePermissionFromRole(ctx context.Context, roleID domain.RoleID, permissionID domain.PermissionID) error
	// GetRolePermissions(ctx context.Context, roleID domain.RoleID) ([]*domain.Permission, error)

	// // Permission Management
	CreatePermission(ctx context.Context, id domain.PermissionID, displayName, description string) (*domain.Permission, error) // ID is often predefined string
	// GetPermission(ctx context.Context, permissionID domain.PermissionID) (*domain.Permission, error)
	GetAllPermissions(ctx context.Context) ([]*domain.Permission, error)
	GetAllRoles(ctx context.Context) ([]*domain.Role, error)

	// // Authorization
	UserHasPermission(ctx context.Context, userID domain.UserID, permissionID domain.PermissionID) (bool, error)
}

type rbacServiceImpl struct {
	repository repository.Repository
	// idGenerator func() string // For generating IDs if not client-provided
}

func NewRBACService(repository repository.Repository) RBACService {
	return &rbacServiceImpl{
		repository: repository,
	}
}

// --- User Management Methods ---
func (s *rbacServiceImpl) CreateUser(ctx context.Context, displayName, email string) (*domain.User, error) {
	// In a real app, you'd generate a unique ID (e.g., UUID)
	// For simplicity, let's assume email could be part of ID or a separate unique ID is generated
	userID := domain.UserID("user-" + email) // Simplistic ID generation
	user := &domain.User{
		ID:          userID,
		DisplayName: displayName,
		Email:       email,
	}
	if err := s.repository.User.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("service.CreateUser: %w", err)
	}
	return user, nil
}

func (s *rbacServiceImpl) GetUser(ctx context.Context, userID domain.UserID) (*domain.User, error) {
	user, err := s.repository.User.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service.GetUser: %w", err)
	}
	return user, nil
}

func (s *rbacServiceImpl) AssignRoleToUser(ctx context.Context, userID domain.UserID, roleID domain.RoleID) error {
	// Optional: Check if user and role exist before assigning
	_, err := s.repository.User.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("service.AssignRoleToUser: user not found: %w", err)
	}
	_, err = s.repository.Role.GetRoleByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("service.AssignRoleToUser: role not found: %w", err)
	}

	if err := s.repository.User.AssignRoleToUser(ctx, userID, roleID); err != nil {
		return fmt.Errorf("service.AssignRoleToUser: %w", err)
	}
	return nil
}

func (s *rbacServiceImpl) RemoveRoleFromUser(ctx context.Context, userID domain.UserID, roleID domain.RoleID) error {
	if err := s.repository.User.RemoveRoleFromUser(ctx, userID, roleID); err != nil {
		return fmt.Errorf("service.RemoveRoleFromUser: %w", err)
	}
	return nil
}

func (s *rbacServiceImpl) GetUserRoles(ctx context.Context, userID domain.UserID) ([]*domain.Role, error) {
	roles, err := s.repository.User.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service.GetUserRoles: %w", err)
	}
	return roles, nil
}

// --- Role Management Methods (implement similarly) ---
func (s *rbacServiceImpl) CreateRole(ctx context.Context, displayName, description string) (*domain.Role, error) {
	roleID := domain.RoleID("role-" + displayName) // Simplistic
	role := &domain.Role{
		ID:          roleID,
		DisplayName: displayName,
		Description: description,
	}
	if err := s.repository.Role.CreateRole(ctx, role); err != nil {
		return nil, fmt.Errorf("service.CreateRole: %w", err)
	}
	return role, nil
}

// ... other role methods

func (s *rbacServiceImpl) AssignPermissionToRole(ctx context.Context, roleID domain.RoleID, permissionID domain.PermissionID) error {
	// Optional: Check if role and permission exist
	_, err := s.repository.Role.GetRoleByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("service.AssignPermissionToRole: role not found: %w", err)
	}
	_, err = s.repository.Permission.GetPermissionByID(ctx, permissionID)
	if err != nil {
		return fmt.Errorf("service.AssignPermissionToRole: permission not found: %w", err)
	}

	if err := s.repository.Role.AssignPermissionToRole(ctx, roleID, permissionID); err != nil {
		return fmt.Errorf("service.AssignPermissionToRole: %w", err)
	}
	return nil
}

// ... other role methods

// --- Permission Management Methods (implement similarly) ---
func (s *rbacServiceImpl) CreatePermission(ctx context.Context, id domain.PermissionID, displayName, description string) (*domain.Permission, error) {
	permission := &domain.Permission{
		ID:          id, // Permissions often have predefined string IDs like "document:create"
		DisplayName: displayName,
		Description: description,
	}
	if err := s.repository.Permission.CreatePermission(ctx, permission); err != nil {
		return nil, fmt.Errorf("service.CreatePermission: %w", err)
	}
	return permission, nil
}

// ... other permission methods

func (s *rbacServiceImpl) GetAllPermissions(ctx context.Context) ([]*domain.Permission, error) {
	return s.repository.Permission.ListAllPermissions(ctx)
}

func (s *rbacServiceImpl) GetAllRoles(ctx context.Context) ([]*domain.Role, error) {
	return s.repository.Role.ListAllRoles(ctx)
}

// --- Authorization Method ---
func (s *rbacServiceImpl) UserHasPermission(ctx context.Context, userID domain.UserID, targetPermissionID domain.PermissionID) (bool, error) {
	roles, err := s.repository.User.GetUserRoles(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) { // User might not exist or have no roles
			return false, nil
		}
		return false, fmt.Errorf("service.UserHasPermission: failed to get user roles: %w", err)
	}

	if len(roles) == 0 {
		return false, nil // No roles, no permissions
	}

	userPermissions := make(map[domain.PermissionID]bool)

	for _, role := range roles {
		permissions, err := s.repository.Role.GetRolePermissions(ctx, role.ID)
		if err != nil {
			// Log this error, but continue checking other roles might be appropriate
			fmt.Printf("Warning: Could not get permissions for role %s: %v\n", role.ID, err)
			continue
		}
		for _, p := range permissions {
			userPermissions[p.ID] = true
		}
	}

	return userPermissions[targetPermissionID], nil
}
