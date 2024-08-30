package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"smee.ovh/toopasbo/config"
	"smee.ovh/toopasbo/gatherers"
	"smee.ovh/toopasbo/transformers"
)

type ChatConfigClientKey string

const chatConfigClientKey ChatConfigClientKey = "chatConfigClient"

type WebhookClient struct {
	ChatConfigClient  *config.ChatConfigClient
	OpenWeatherClient *gatherers.OpenWeatherClient
	Bot               *bot.Bot
	ctx               context.Context
}

func NewWebhookClient(ctx context.Context) (*WebhookClient, error) {
	chatConfigClient, err := config.NewChatConfigClient(ctx, false)
	if err != nil {
		return nil, err
	}

	opts := []bot.Option{
		// bot.WithDefaultHandler(handler),
	}
	appConfig := config.GetAppConfig(ctx)

	openWeatherClient, err := gatherers.NewOpenWeatherClient(ctx)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, chatConfigClientKey, chatConfigClient)

	b, err := bot.New(appConfig.TelegramBotToken, opts...)
	if err != nil {
		panic(err)
	}

	// b.RegisterHandler(bot.HandlerTypeMessageText, "/register", bot.MatchTypeExact, registerHandler)
	// b.RegisterHandler(bot.HandlerTypeMessageText, "/delete", bot.MatchTypeExact, deleteHandler)
	// b.RegisterHandler(bot.HandlerTypeMessageText, "/disable", bot.MatchTypeExact, disableHandler)
	// b.RegisterHandler(bot.HandlerTypeMessageText, "/meteo", bot.MatchTypeExact, meteoHandler)

	// go b.StartWebhook(ctx)

	return &WebhookClient{ChatConfigClient: chatConfigClient, Bot: b, OpenWeatherClient: openWeatherClient, ctx: ctx}, nil
}

func (wc *WebhookClient) HttpHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		body, errReadBody := io.ReadAll(req.Body)
		if errReadBody != nil {
			log.Fatal("error read request body, %w", errReadBody)
			return
		}

		update := &models.Update{}

		errDecode := json.Unmarshal(body, update)
		if errDecode != nil {
			log.Fatal("error decode request body, %w", body, errDecode)
			return
		}

		message := "Unprocessed"
		if update.Message == nil {
			io.WriteString(w, message)
			return
		}

		if update.Message.Text == "/register" {
			message = wc.registerHandler(update)
		} else if update.Message.Text == "/delete" {
			message = wc.deleteHandler(update)
		} else if update.Message.Text == "/disable" {
			message = wc.disableHandler(update)
		} else if strings.HasPrefix(update.Message.Text, "/meteo") {
			message = wc.meteoHandler(update)
		} else {
			message = ""
		}

		io.WriteString(w, message)

	}
}

func (wc *WebhookClient) registerHandler(update *models.Update) string {
	chatConfig := config.ChatConfig{
		ID:      strconv.FormatInt(update.Message.Chat.ID, 10),
		Enabled: true,
	}
	err := wc.ChatConfigClient.WriteChatConfig(wc.ctx, chatConfig)
	if err != nil {
		log.Printf("failed to set chat config: %v", err)
		return "Failed to register"
	}
	wc.Bot.SendMessage(wc.ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Registered",
	})

	return "Registered"
}

func (wc *WebhookClient) deleteHandler(update *models.Update) string {
	chatConfigClient := wc.ctx.Value(chatConfigClientKey).(*config.ChatConfigClient)
	chatConfig := config.ChatConfig{
		ID:      strconv.FormatInt(update.Message.Chat.ID, 10),
		Enabled: false,
	}
	err := chatConfigClient.DeleteChatConfig(wc.ctx, chatConfig)
	if err != nil {
		log.Printf("failed to set chat config: %v", err)
		return "Failed to delete"
	}
	wc.Bot.SendMessage(wc.ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Deleted",
	})

	return "Deleted"
}

func (wc *WebhookClient) disableHandler(update *models.Update) string {
	chatConfigClient := wc.ctx.Value(chatConfigClientKey).(*config.ChatConfigClient)
	chatConfig := config.ChatConfig{
		ID:      strconv.FormatInt(update.Message.Chat.ID, 10),
		Enabled: true,
	}
	err := chatConfigClient.DeleteChatConfig(wc.ctx, chatConfig)
	if err != nil {
		log.Printf("failed to set chat config: %v", err)
		return "Failed to disable"
	}
	wc.Bot.SendMessage(wc.ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Disabled",
	})

	return "Disabled"
}

func parseCommand(command string) (string, string) {
	parts := strings.SplitN(command, " ", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

var captionTemplate = `
%s.
Current temperature is %d°C. Expect a minimum of %d°C and a maximum of %d°C.
`

func WeatherToTelegramText(weather gatherers.Weather) string {
	return fmt.Sprintf(captionTemplate, weather.Summary, weather.CurrentTemperature, weather.MinTemperature, weather.MaxTemperature)
}

func WeatherToShortText(weather gatherers.Weather) string {
	return fmt.Sprintf("Min: %d°C |  Max: %d°C - %s", weather.MinTemperature, weather.MaxTemperature, weather.Description)
}

func (wc *WebhookClient) meteoHandler(update *models.Update) string {
	_, city := parseCommand(update.Message.Text)
	position, posErr := wc.OpenWeatherClient.GetCityPosition(city)
	if posErr != nil {
		wc.Bot.SendMessage(wc.ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("City '%s' not found", city),
		})
		fmt.Printf("City '%s' not found: %v", city, posErr)
		return "City not found"
	}

	weather, weatherErr := wc.OpenWeatherClient.GetWeather(position)
	if weatherErr != nil {
		wc.Bot.SendMessage(wc.ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Error getting weather for city '%s'", city),
		})
		fmt.Printf("Error getting weather: %v", weatherErr)
		return "Error getting weather"
	}

	// imageUrl, err := transformers.GenerateMidjourneyPicture(weather)
	imageClient, err := transformers.NewFalAIClient(wc.ctx, "fal-ai/flux-pro")
	if err != nil {
		wc.Bot.SendMessage(wc.ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Error creating image client for city '%s'", city),
		})
		fmt.Printf("Error creating image client: %v", err)
		return "Error creating image client"
	}

	image, err := imageClient.GenerateWeatherImage(weather)
	if err != nil {
		wc.Bot.SendMessage(wc.ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Error generating image for city '%s'", city),
		})
		fmt.Printf("Error generating image: %v", err)
		return "Error generating image"
	}

	imagePath, err := DownloadFile(image)
	if err != nil {
		wc.Bot.SendMessage(wc.ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Error downloading image for city '%s'", city),
		})
		fmt.Printf("Error downloading image: %v", err)
		return "Error downloading image"
	}

	imageData, errRead := os.ReadFile(imagePath)
	if errRead != nil {
		wc.Bot.SendMessage(wc.ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Error decoding image for city '%s'", city),
		})
		fmt.Printf("Error decoding image: %v", errRead)
		return "Error decoding image"
	}

	_, errTelegram := wc.Bot.SendPhoto(wc.ctx, &bot.SendPhotoParams{
		ChatID:  update.Message.Chat.ID,
		Photo:   &models.InputFileUpload{Filename: "picture.png", Data: bytes.NewReader(imageData)},
		Caption: WeatherToTelegramText(weather),
	})

	if errTelegram != nil {
		wc.Bot.SendMessage(wc.ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Error sending image for city '%s'", city),
		})
		fmt.Printf("Error sending image: %v", errTelegram)
		return "Error sending image"
	}

	return "Weather sent"
}

func (wc *WebhookClient) SendToAll(ctx context.Context, imagePath string, weather string) {
	chatConfigs, err := wc.ChatConfigClient.GetAllChatConfigs(ctx)
	if err != nil {
		log.Printf("failed to get chat configs: %v", err)
		return
	}

	imageData, errRead := os.ReadFile(imagePath)
	if errRead != nil {
		panic(errRead)
	}

	for _, chatConfig := range chatConfigs {
		if !chatConfig.Enabled {
			continue
		}

		_, errTelegram := wc.Bot.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:  chatConfig.ID,
			Photo:   &models.InputFileUpload{Filename: "picture.png", Data: bytes.NewReader(imageData)},
			Caption: weather,
		})

		if errTelegram != nil {
			wc.Bot.SendMessage(wc.ctx, &bot.SendMessageParams{
				ChatID: chatConfig.ID,
				Text:   "Error sending recurring image",
			})
			log.Fatal("Error sending image: %w", errTelegram)
		}
	}

}
