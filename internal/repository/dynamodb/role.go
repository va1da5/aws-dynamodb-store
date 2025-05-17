package dynamodb

import (
	"aws-dynamodb-store/internal/config"
	"aws-dynamodb-store/internal/domain"
	"aws-dynamodb-store/internal/repository"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type roleItem struct {
	baseItem
	ID          domain.RoleID `dynamodbav:"EntityID"`
	DisplayName string        `dynamodbav:"DisplayName"`
	Description string        `dynamodbav:"Description,omitempty"`
	CreatedAt   time.Time     `dynamodbav:"CreatedAt"`
	UpdatedAt   time.Time     `dynamodbav:"UpdatedAt"`
}

type DynamoDBRoleRepository struct {
	client *dynamodb.Client
	config config.DynamoDBConfig
}

func NewDynamoDBRoleRepository(client *dynamodb.Client, config config.DynamoDBConfig) repository.RoleRepository {
	return &DynamoDBRoleRepository{client: client, config: config}
}

func roleToItem(role *domain.Role) *roleItem {
	pk := RolePrefix + string(role.ID)
	return &roleItem{
		baseItem: baseItem{
			PK:         pk,
			SK:         MetadataPrefix + string(role.ID),
			EntityType: EntityTypeUser,
		},
		ID:          role.ID,
		DisplayName: role.DisplayName,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

func itemToRole(item *roleItem) *domain.Role {
	return &domain.Role{
		ID:          item.ID,
		DisplayName: item.DisplayName,
		Description: item.Description,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func (r *DynamoDBRoleRepository) CreateRole(ctx context.Context, role *domain.Role) error {
	role.CreatedAt = time.Now().UTC()
	role.UpdatedAt = role.CreatedAt
	item := roleToItem(role)

	return createItem(ctx, r.client, r.config.TableName, item)

}

func (r *DynamoDBRoleRepository) GetRoleByID(ctx context.Context, id domain.RoleID) (*domain.Role, error) {
	pk := RolePrefix + string(id)
	sk := MetadataPrefix + string(id)

	out, err := getItemById(ctx, r.client, r.config.TableName, pk, sk)
	if err != nil {
		return nil, err
	}

	var roleItem roleItem
	if err := attributevalue.UnmarshalMap(out.Item, &roleItem); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user item: %w", err)
	}
	return itemToRole(&roleItem), nil
}

// TODO
func (r *DynamoDBRoleRepository) ListAllRoles(ctx context.Context) ([]*domain.Role, error) {
	return []*domain.Role{}, nil
}

// TODO
func (r *DynamoDBRoleRepository) AssignPermissionToRole(ctx context.Context, roleID domain.RoleID, permissionID domain.PermissionID) error {
	return nil
}

// TODO
func (r *DynamoDBRoleRepository) GetRolePermissions(ctx context.Context, roleID domain.RoleID) ([]*domain.Permission, error) {
	return []*domain.Permission{}, nil
}

// --- DynamoDBRoleRepository ---
// (Similar structure to DynamoDBUserRepository)
// - CreateRole, GetRoleByID, UpdateRole, DeleteRole
// - AssignPermissionToRole (PK: ROLE#roleID, SK: PERMISSION#permID)
// - RemovePermissionFromRole
// - GetRolePermissions (Query PK=ROLE#roleID, SK begins_with PERMISSION#)
// - ListRolesWithPermission (Query GSI1 GSI1PK=PERMISSION#permID, GSI1SK begins_with ROLE#)
// - ListAllRoles (Query for EntityType=ROLE on GSI or do a scan if few roles and infrequent)
