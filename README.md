# Project aws-dynamodb-store

One Paragraph of project description goes here

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

## AWS DynamoDB

```bash
# create table
aws dynamodb create-table \
    --endpoint-url http://localhost:8000 \
    --table-name rbac \
    --attribute-definitions \
        AttributeName=PK,AttributeType=S \
        AttributeName=SK,AttributeType=S \
        AttributeName=EntityType,AttributeType=S \
        AttributeName=EntityID,AttributeType=S \
    --key-schema \
        AttributeName=PK,KeyType=HASH \
        AttributeName=SK,KeyType=RANGE \
    --billing-mode=PAY_PER_REQUEST \
    --provisioned-throughput ReadCapacityUnits=10,WriteCapacityUnits=5 \
    --global-secondary-indexes \
        "[
            {
                \"IndexName\": \"EntityTypeIndex\",
                \"KeySchema\": [
                    {\"AttributeName\":\"EntityType\",\"KeyType\":\"HASH\"},
                    {\"AttributeName\":\"EntityID\",\"KeyType\":\"RANGE\"}
                ],
                \"Projection\": {
                    \"ProjectionType\":\"KEYS_ONLY\"
                },
                \"ProvisionedThroughput\": {
                    \"ReadCapacityUnits\": 10,
                    \"WriteCapacityUnits\": 5
                }
            }
        ]"

```

## MakeFile

Run build make command with tests

```bash
make all
```

Build the application

```bash
make build
```

Run the application

```bash
make run
```

Live reload the application:

```bash
make watch
```

Run the test suite:

```bash
make test
```

Clean up binary from the last build:

```bash
make clean
```

## References

- [Amazon DynamoDB Documentation](https://docs.aws.amazon.com/dynamodb/)
- [AWS DynamoDB Guides - Everything you need to know about DynamoDB](https://www.youtube.com/playlist?list=PL9nWRykSBSFi5QD8ssI0W5odL9S0309E2)
- [AWS re:Invent 2018: Amazon DynamoDB Deep Dive: Advanced Design Patterns for DynamoDB (DAT401)](https://www.youtube.com/watch?v=HaEPXoXVf2k)
- [AWS re:Invent 2021 - DynamoDB deep dive: Advanced design patterns](https://www.youtube.com/watch?v=xfxBhvGpoa0)
- [Common Single-Table design modeling mistakes with DynamoDB](https://www.youtube.com/watch?v=XMEkNZby95M)
- [Single-Table Design with DynamoDB - Alex DeBrie, AWS Data Hero](https://www.youtube.com/watch?v=BnDKD_Zv0og)
- [Deploying DynamoDB locally on your computer](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.DownloadingAndRunning.html)
