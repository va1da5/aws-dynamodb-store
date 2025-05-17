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

type permissionItem struct {
	baseItem
	ID          domain.PermissionID `dynamodbav:"EntityID"`
	DisplayName string              `dynamodbav:"DisplayName"`
	Description string              `dynamodbav:"Description,omitempty"`
	CreatedAt   time.Time           `dynamodbav:"CreatedAt"`
	UpdatedAt   time.Time           `dynamodbav:"UpdatedAt"`
}

type DynamoDBPermissionRepository struct {
	client *dynamodb.Client
	config config.DynamoDBConfig
}

func NewDynamoDBPermissionRepository(client *dynamodb.Client, config config.DynamoDBConfig) repository.PermissionRepository {
	return &DynamoDBPermissionRepository{client: client, config: config}
}

func permissionToItem(permission *domain.Permission) *permissionItem {
	pk := RolePrefix + string(permission.ID)
	return &permissionItem{
		baseItem: baseItem{
			PK:         pk,
			SK:         MetadataPrefix + string(permission.ID),
			EntityType: EntityTypeUser,
		},
		ID:          permission.ID,
		DisplayName: permission.DisplayName,
		Description: permission.Description,
		CreatedAt:   permission.CreatedAt,
		UpdatedAt:   permission.UpdatedAt,
	}
}

func itemToPermission(item *permissionItem) *domain.Permission {
	return &domain.Permission{
		ID:          item.ID,
		DisplayName: item.DisplayName,
		Description: item.Description,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func (r *DynamoDBPermissionRepository) CreatePermission(ctx context.Context, permission *domain.Permission) error {
	permission.CreatedAt = time.Now().UTC()
	permission.UpdatedAt = permission.CreatedAt
	item := permissionToItem(permission)

	return createItem(ctx, r.client, r.config.TableName, item)
}

func (r *DynamoDBPermissionRepository) GetPermissionByID(ctx context.Context, id domain.PermissionID) (*domain.Permission, error) {
	pk := PermissionPrefix + string(id)
	sk := MetadataPrefix + string(id)

	out, err := getItemById(ctx, r.client, r.config.TableName, pk, sk)
	if err != nil {
		return nil, err
	}

	var permissionItem permissionItem
	if err := attributevalue.UnmarshalMap(out.Item, &permissionItem); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user item: %w", err)
	}
	return itemToPermission(&permissionItem), nil
}

// TODO
func (r *DynamoDBPermissionRepository) ListAllPermissions(ctx context.Context) ([]*domain.Permission, error) {
	return []*domain.Permission{}, nil
}

// --- DynamoDBPermissionRepository ---
// (Similar structure)
// - CreatePermission, GetPermissionByID, UpdatePermission, DeletePermission
// - ListAllPermissions (Query for EntityType=PERMISSION on GSI or scan)
