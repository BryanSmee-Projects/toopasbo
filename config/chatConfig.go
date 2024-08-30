package config

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func NewDynamoDBClient(region string) (*dynamodb.DynamoDB, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		return nil, err
	}

	return dynamodb.New(sess), nil
}

type ChatConfig struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

type ChatConfigClient struct {
	DynamoDBClient *dynamodb.DynamoDB
	TableName      string
	Local          bool
}

func NewChatConfigClient(ctx context.Context, local bool) (*ChatConfigClient, error) {
	if local {
		return &ChatConfigClient{Local: true}, nil
	}

	appConfig := GetAppConfig(ctx)

	dynamoDBClient, err := NewDynamoDBClient(appConfig.DBConfig.Region)
	if err != nil {
		return nil, err
	}

	return &ChatConfigClient{DynamoDBClient: dynamoDBClient, TableName: appConfig.DBConfig.TableName}, nil
}

func writeLocalChatConfig(ctx context.Context, chatConfig ChatConfig) error {
	appConfig := GetAppConfig(ctx)
	chatConfigFile := appConfig.ChatConfigFolder + "/" + chatConfig.ID + ".json"
	item, err := json.Marshal(chatConfig)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(appConfig.ChatConfigFolder, 0755); err != nil {
		return err
	}

	err = os.WriteFile(chatConfigFile, item, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *ChatConfigClient) WriteChatConfig(ctx context.Context, chatConfig ChatConfig) error {
	if c.Local {
		return writeLocalChatConfig(ctx, chatConfig)
	}

	item, err := dynamodbattribute.MarshalMap(chatConfig)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(c.TableName),
		Item:      item,
	}

	_, err = c.DynamoDBClient.PutItem(input)
	return err
}

func deleteLocalChatConfig(ctx context.Context, chatConfig ChatConfig) error {
	appConfig := GetAppConfig(ctx)
	chatConfigFile := appConfig.ChatConfigFolder + "/" + chatConfig.ID + ".json"
	return os.Remove(chatConfigFile)
}

func (c *ChatConfigClient) DeleteChatConfig(ctx context.Context, chatConfig ChatConfig) error {
	if c.Local {
		return deleteLocalChatConfig(ctx, chatConfig)
	}

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(c.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(chatConfig.ID),
			},
		},
	}

	_, err := c.DynamoDBClient.DeleteItem(input)
	return err
}

func getChatConfigLocal(ctx context.Context, chatConfig ChatConfig) (ChatConfig, error) {
	appConfig := GetAppConfig(ctx)
	chatConfigFile := appConfig.ChatConfigFolder + "/" + chatConfig.ID + ".json"
	item, err := os.ReadFile(chatConfigFile)
	if err != nil {
		return chatConfig, err
	}

	err = json.Unmarshal(item, &chatConfig)
	return chatConfig, err
}

func (c *ChatConfigClient) GetChatConfig(ctx context.Context, chatConfig ChatConfig) (ChatConfig, error) {
	if c.Local {
		return getChatConfigLocal(ctx, chatConfig)
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(c.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(chatConfig.ID),
			},
		},
	}

	result, err := c.DynamoDBClient.GetItem(input)
	if err != nil {
		return chatConfig, err
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, &chatConfig)
	return chatConfig, err
}

func getAllChatConfigsLocal(ctx context.Context) ([]ChatConfig, error) {
	appConfig := GetAppConfig(ctx)
	files, err := os.ReadDir(appConfig.ChatConfigFolder)
	if err != nil {
		return nil, err
	}

	var chatConfigs []ChatConfig
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		chatConfigFile := appConfig.ChatConfigFolder + "/" + file.Name()
		item, err := os.ReadFile(chatConfigFile)
		if err != nil {
			return nil, err
		}

		var chatConfig ChatConfig
		err = json.Unmarshal(item, &chatConfig)
		if err != nil {
			return nil, err
		}

		chatConfigs = append(chatConfigs, chatConfig)
	}
	return chatConfigs, nil
}

func (c *ChatConfigClient) GetAllChatConfigs(ctx context.Context) ([]ChatConfig, error) {
	if c.Local {
		return getAllChatConfigsLocal(ctx)
	}

	var chatConfigs []ChatConfig
	input := &dynamodb.ScanInput{
		TableName: aws.String(c.TableName),
	}

	result, err := c.DynamoDBClient.Scan(input)
	if err != nil {
		return nil, err
	}

	for _, item := range result.Items {
		var chatConfig ChatConfig
		err = dynamodbattribute.UnmarshalMap(item, &chatConfig)
		if err != nil {
			return nil, err
		}
		chatConfigs = append(chatConfigs, chatConfig)
	}
	return chatConfigs, nil
}
