package config

import "context"

type Config struct {
	TelegramBotToken  string
	ChatConfigFolder  string
	OpenaiAPIKey      string
	FreepikAPIKey     string
	FalAIAPIKey       string
	OpenWeatherAPIKey string
	MidjourneyApiUrl  string
	Runtime           string
	DBConfig          DBConfig
}

type DBConfig struct {
	TableName string
	Region    string
}

type appConfigKey string

const AppConfigContextKey appConfigKey = "appConfig"

func GetAppConfig(ctx context.Context) Config {
	return ctx.Value(AppConfigContextKey).(Config)
}
