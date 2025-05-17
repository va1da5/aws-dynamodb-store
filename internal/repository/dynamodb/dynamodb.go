// internal/repository/dynamodb/dynamodb_rbac.go
package dynamodb

import (
	"aws-dynamodb-store/internal/config"
	"aws-dynamodb-store/internal/repository"
	"context"
	"errors"
	"fmt"
	"log"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	EntityTypeUser       = "USER"
	EntityTypeRole       = "ROLE"
	EntityTypePermission = "PERMISSION"
	MetadataPrefix       = "METADATA#"
	UserPrefix           = "USER#"
	RolePrefix           = "ROLE#"
	PermissionPrefix     = "PERMISSION#"
	GSI1Name             = "GSI1" // Name of your GSI (SK-PK)

)

// Helper struct for DynamoDB items
type baseItem struct {
	PK         string `dynamodbav:"PK"`
	SK         string `dynamodbav:"SK"`
	EntityType string `dynamodbav:"EntityType,omitempty"` // For filtering and clarity
}

func NewDynamoDBRepository(cfg config.DynamoDBConfig) repository.Repository {

	var client *dynamodb.Client

	sdkConfig, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion(cfg.AWSRegion),
	)

	if err != nil {
		log.Fatal(err.Error())
	}

	client = dynamodb.NewFromConfig(sdkConfig)

	if cfg.UseDynamoDBLocal {
		log.Printf("DynamoDB endpoint: %s", cfg.DynamoDBLocalURL)
		client = dynamodb.NewFromConfig(sdkConfig,
			func(o *dynamodb.Options) {
				o.BaseEndpoint = aws.String(cfg.DynamoDBLocalURL)
				o.Region = cfg.AWSRegion
			},
		)
	}

	return repository.Repository{
		User:       NewDynamoDBUserRepository(client, cfg),
		Role:       NewDynamoDBRoleRepository(client, cfg),
		Permission: NewDynamoDBPermissionRepository(client, cfg),
	}
}

func createItem(ctx context.Context, client *dynamodb.Client, tableName string, item interface{}) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(tableName),
		Item:                av,
		ConditionExpression: aws.String("attribute_not_exists(PK)"), // Ensure permission doesn't already exist
	})
	if err != nil {
		var condCheckFailed *types.ConditionalCheckFailedException
		if errors.As(err, &condCheckFailed) {
			return repository.ErrAlreadyExists
		}
		return fmt.Errorf("failed to put item: %w", err)
	}
	return nil
}

func getItemById(ctx context.Context, client *dynamodb.Client, tableName string, pk string, sk string) (*dynamodb.GetItemOutput, error) {
	out, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
			"SK": &types.AttributeValueMemberS{Value: sk},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user item: %w", err)
	}
	if out.Item == nil {
		return nil, repository.ErrNotFound
	}

	return out, nil
}

// Implement UpdateUser and DeleteUser similarly. DeleteUser will need to:
// 1. Find all USER#id / ROLE#roleID items and delete them.
// 2. Delete the USER#id / METADATA#id item.
// This might be better done with BatchWriteItem or TransactWriteItems for atomicity.

// For brevity, full implementation of Role and Permission repos are omitted but follow the same pattern.
// Ensure proper marshalling/unmarshalling for roleItem and permissionItem.
// For example, GetRolePermissions would query PK=ROLE#roleID and SK begins_with PERMISSION#,
// then for each result, extract the permissionID from SK and call GetPermissionByID on PermissionRepository.
