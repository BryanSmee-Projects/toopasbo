package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func NewConfigFromEnv() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	config := &Config{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		ChatConfigFolder: os.Getenv("CHAT_CONFIG_FOLDER"),
		OpenaiAPIKey:     os.Getenv("OPENAI_API_KEY"),
		FreepikAPIKey:    os.Getenv("FREEPIK_API_KEY"),
		Runtime:          "local",
	}
	if config.TelegramBotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}
	if config.ChatConfigFolder == "" {
		log.Fatal("CHAT_CONFIG_FOLDER is not set")
	}
	if config.OpenaiAPIKey == "" {
		log.Fatal("OPENAI_API_KEY is not set")
	}
	if config.FreepikAPIKey == "" {
		log.Fatal("FREEPIK_API_KEY is not set")
	}
	return *config
}
