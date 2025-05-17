package dynamodb

import (
	"aws-dynamodb-store/internal/config"
	"aws-dynamodb-store/internal/domain"
	"aws-dynamodb-store/internal/repository"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type userItem struct {
	baseItem
	ID          domain.UserID `dynamodbav:"EntityID"` // Store the raw ID too
	DisplayName string        `dynamodbav:"DisplayName"`
	Email       string        `dynamodbav:"Email"`
	CreatedAt   time.Time     `dynamodbav:"CreatedAt"`
	UpdatedAt   time.Time     `dynamodbav:"UpdatedAt"`
}

type DynamoDBUserRepository struct {
	client *dynamodb.Client
	config config.DynamoDBConfig
}

func NewDynamoDBUserRepository(client *dynamodb.Client, config config.DynamoDBConfig) repository.UserRepository {
	return &DynamoDBUserRepository{client: client, config: config}
}

func userToItem(user *domain.User) *userItem {
	pk := UserPrefix + string(user.ID)
	return &userItem{
		baseItem: baseItem{
			PK:         pk,
			SK:         MetadataPrefix + string(user.ID),
			EntityType: EntityTypeUser,
		},
		ID:          user.ID,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

func itemToUser(item *userItem) *domain.User {
	return &domain.User{
		ID:          item.ID,
		DisplayName: item.DisplayName,
		Email:       item.Email,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func (r *DynamoDBUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = user.CreatedAt
	item := userToItem(user)

	return createItem(ctx, r.client, r.config.TableName, item)

}

func (r *DynamoDBUserRepository) GetUserByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	pk := UserPrefix + string(id)
	sk := MetadataPrefix + string(id)

	out, err := getItemById(ctx, r.client, r.config.TableName, pk, sk)
	if err != nil {
		return nil, err
	}

	var userItem userItem
	if err := attributevalue.UnmarshalMap(out.Item, &userItem); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user item: %w", err)
	}
	return itemToUser(&userItem), nil
}

func (r *DynamoDBUserRepository) ListAllUsers(ctx context.Context) ([]*domain.User, error) {
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(r.config.TableName),
		IndexName:              aws.String(r.config.EntityTypeIndex),
		KeyConditionExpression: aws.String("EntityType = :entityTypeVal"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":entityTypeVal": &types.AttributeValueMemberS{Value: EntityTypeUser},
		},
	}

	paginator := dynamodb.NewQueryPaginator(r.client, queryInput)
	var users []*domain.User

	// roleRepo := NewDynamoDBRoleRepository(r.client, r.tableName) // Or inject it

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to query user roles: %w", err)
		}
		for _, item := range page.Items {
			var partialUser userItem
			if err := attributevalue.UnmarshalMap(item, &partialUser); err != nil {
				log.Print(err.Error())
				continue
			}

			user, err := r.GetUserByID(ctx, partialUser.ID)
			if err != nil {
				log.Print(err.Error())
				continue
			}

			users = append(users, user)
		}
	}
	return users, nil

}

func (r *DynamoDBUserRepository) AssignRoleToUser(ctx context.Context, userID domain.UserID, roleID domain.RoleID) error {
	item := map[string]types.AttributeValue{
		"PK":         &types.AttributeValueMemberS{Value: UserPrefix + string(userID)},
		"SK":         &types.AttributeValueMemberS{Value: RolePrefix + string(roleID)},
		"EntityType": &types.AttributeValueMemberS{Value: "UserRoleAssignment"}, // Optional but good for clarity
		"AssignedAt": &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
	}
	_, err := r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.config.TableName),
		Item:      item,
	})
	// Consider adding checks: Does user exist? Does role exist? (Usually done in service layer)
	if err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}
	return nil
}

func (r *DynamoDBUserRepository) RemoveRoleFromUser(ctx context.Context, userID domain.UserID, roleID domain.RoleID) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.config.TableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: UserPrefix + string(userID)},
			"SK": &types.AttributeValueMemberS{Value: RolePrefix + string(roleID)},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to remove role from user: %w", err)
	}
	return nil
}

func (r *DynamoDBUserRepository) GetUserRoles(ctx context.Context, userID domain.UserID) ([]*domain.Role, error) {
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(r.config.TableName),
		KeyConditionExpression: aws.String("PK = :pkVal AND begins_with(SK, :skPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pkVal":    &types.AttributeValueMemberS{Value: UserPrefix + string(userID)},
			":skPrefix": &types.AttributeValueMemberS{Value: RolePrefix},
		},
	}

	paginator := dynamodb.NewQueryPaginator(r.client, queryInput)
	var roles []*domain.Role

	// Need a separate RoleRepository or a helper to fetch Role details
	// For simplicity here, we'll assume the SK contains enough info or we do a subsequent GetItem for role details.
	// A better approach would be to have a GetRoleByID in RoleRepository and call it.
	// Let's assume for now SK for User-Role assignment is just ROLE#roleID.
	// We need to fetch the actual role metadata.

	roleRepo := NewDynamoDBRoleRepository(r.client, r.config) // Or inject it

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to query user roles: %w", err)
		}
		for _, item := range page.Items {
			var base baseItem
			if err := attributevalue.UnmarshalMap(item, &base); err != nil {
				// log error and continue or return
				continue
			}
			// SK should be like ROLE#role-id
			roleIDStr := base.SK[len(RolePrefix):]
			role, err := roleRepo.GetRoleByID(ctx, domain.RoleID(roleIDStr))
			if err != nil {
				// Handle error, maybe role was deleted, log and continue
				fmt.Printf("Warning: could not fetch role %s for user %s: %v\n", roleIDStr, userID, err)
				continue
			}
			roles = append(roles, role)
		}
	}
	return roles, nil
}

func (r *DynamoDBUserRepository) ListUsersInRole(ctx context.Context, roleID domain.RoleID) ([]*domain.User, error) {
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(r.config.TableName),
		IndexName:              aws.String(GSI1Name), // GSI1PK = SK, GSI1SK = PK
		KeyConditionExpression: aws.String("SK = :skVal AND begins_with(PK, :pkPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":skVal":    &types.AttributeValueMemberS{Value: RolePrefix + string(roleID)},
			":pkPrefix": &types.AttributeValueMemberS{Value: UserPrefix},
		},
	}

	paginator := dynamodb.NewQueryPaginator(r.client, queryInput)
	var users []*domain.User

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to query users in role using GSI1: %w", err)
		}
		for _, itemMap := range page.Items {
			var base baseItem
			if err := attributevalue.UnmarshalMap(itemMap, &base); err != nil {
				// log and continue
				continue
			}
			// PK should be USER#userID
			userIDStr := base.PK[len(UserPrefix):]
			user, err := r.GetUserByID(ctx, domain.UserID(userIDStr)) // Fetch full user details
			if err != nil {
				// Handle error, maybe user was deleted
				fmt.Printf("Warning: could not fetch user %s for role %s: %v\n", userIDStr, roleID, err)
				continue
			}
			users = append(users, user)
		}
	}
	return users, nil
}
