package config

import (
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type AWSConfig struct {
	TelegramBotToken  string `json:"TELEGRAM_BOT_TOKEN"`
	OpenaiAPIKey      string `json:"OPENAI_API_KEY"`
	FalAIAPIKey       string `json:"FAL_AI_API_KEY"`
	OpenWeatherAPIKey string `json:"OPENWEATHER_API_KEY"`
}

func LoadDBConfig() DBConfig {
	dbConfig := DBConfig{
		TableName: os.Getenv("DYNAMODB_TABLE_NAME"),
		Region:    os.Getenv("AWS_REGION"),
	}
	if dbConfig.TableName == "" {
		log.Fatal("DYNAMODB_TABLE_NAME is not set")
	}
	if dbConfig.Region == "" {
		log.Fatal("AWS_REGION is not set")
	}
	return dbConfig
}

func NewConfigFromAWS() Config {
	// Replace with your AWS Secret Manager secret name
	secretName := os.Getenv("AWS_SECRET_NAME")
	region := os.Getenv("AWS_REGION")

	if secretName == "" {
		log.Fatal("AWS_SECRET_NAME is not set")
	}
	if region == "" {
		log.Fatal("AWS_REGION is not set")
	}

	// Create a new session with the default config
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		log.Fatalf("failed to create session: %v", err)
	}

	// Create a Secrets Manager client
	svc := secretsmanager.New(sess)

	// Retrieve the secret from AWS Secrets Manager
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := svc.GetSecretValue(input)
	if err != nil {
		log.Fatalf("failed to get secret: %v", err)
	}

	awsConfig := AWSConfig{}
	err = json.Unmarshal([]byte(*result.SecretString), &awsConfig)
	if err != nil {
		log.Fatalf("failed to unmarshal secret: %v", err)
	}

	config := &Config{
		TelegramBotToken:  awsConfig.TelegramBotToken,
		OpenaiAPIKey:      awsConfig.OpenaiAPIKey,
		OpenWeatherAPIKey: awsConfig.OpenWeatherAPIKey,
		FalAIAPIKey:       awsConfig.FalAIAPIKey,
		Runtime:           "aws",
		DBConfig:          LoadDBConfig(),
	}

	if config.TelegramBotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}
	if config.OpenaiAPIKey == "" {
		log.Fatal("OPENAI_API_KEY is not set")
	}
	if config.FalAIAPIKey == "" {
		log.Fatal("FAL_AI_API_KEY is not set")
	}
	if config.OpenWeatherAPIKey == "" {
		log.Fatal("OPENWEATHER_API_KEY is not set")
	}

	return *config
}
