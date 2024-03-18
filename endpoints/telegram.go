package endpoints

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"smee.ovh/toopasbo/gatherers"
	"smee.ovh/toopasbo/transformers"
)

var telegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
var chatIDsFile = "./persistentdata/chat_ids.txt"

func loadChatIDs() ([]string, error) {
	if _, err := os.Stat(chatIDsFile); os.IsNotExist(err) {
		return []string{}, nil
	}
	file, err := os.Open(chatIDsFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	return lines, nil
}

func registerChatID(chatID string) error {

	chatIDs, err := loadChatIDs()
	if err != nil {
		return err
	}
	if contains(chatIDs, chatID) {
		return nil
	}
	chatIDs = append(chatIDs, chatID)
	file, err := os.Create(chatIDsFile)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, id := range chatIDs {
		_, err := writer.WriteString(id + "\n")
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}

func StartTelegram(ctx context.Context) *bot.Bot {
	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(telegramBotToken, opts...)
	if err != nil {
		panic(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/register", bot.MatchTypeExact, registerHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/meteo", bot.MatchTypePrefix, meteoHandler)

	go b.Start(ctx)

	return b
}

var captionTemplate = `
%s.
Current temperature is %d°C. Expect a minimum of %d°C and a maximum of %d°C.
`

func weatherToTelegramText(weather gatherers.Weather) string {
	return fmt.Sprintf(captionTemplate, weather.Summary, weather.CurrentTemperature, weather.MinTemperature, weather.MaxTemperature)
}

func SendImageToTelegram(imagePath string, weather gatherers.Weather, b *bot.Bot, ctx context.Context) {
	chatIDs, err := loadChatIDs()
	if err != nil {
		panic(err)
	}

	imageData, errRead := os.ReadFile(imagePath)
	if errRead != nil {
		panic(errRead)
	}

	for _, chatID := range chatIDs {
		_, errTelegram := b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:  chatID,
			Photo:   &models.InputFileUpload{Filename: "dalle.png", Data: bytes.NewReader(imageData)},
			Caption: weatherToTelegramText(weather),
		})

		if errTelegram != nil {
			panic(errTelegram)
		}
	}
}

func parseCommand(command string) (string, string) {
	parts := strings.SplitN(command, " ", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func meteoHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, city := parseCommand(update.Message.Text)
	position, posErr := gatherers.GetCityPosition(city)
	if posErr != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("City '%s' not found", city),
		})
		fmt.Printf("City '%s' not found: %v", city, posErr)
	}

	weather, weatherErr := gatherers.GetWeather(position)
	if weatherErr != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Error getting weather for city '%s'", city),
		})
		fmt.Printf("Error getting weather: %v", weatherErr)
	}

	imagePath, dallErr := transformers.GenerateDallEPicture(weather)
	if dallErr != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Error generating image for city '%s'", city),
		})
		fmt.Printf("Error generating image: %v", dallErr)
	}

	imageData, errRead := os.ReadFile(imagePath)
	if errRead != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Error reading image for city '%s'", city),
		})
		fmt.Printf("Error reading image: %v", errRead)
	}

	_, errTelegram := b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:  update.Message.Chat.ID,
		Photo:   &models.InputFileUpload{Filename: "dalle.png", Data: bytes.NewReader(imageData)},
		Caption: weatherToTelegramText(weather),
	})

	if errTelegram != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Error sending image for city '%s'", city),
		})
		fmt.Printf("Error sending image: %v", errTelegram)
	}

}

func registerHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := strconv.FormatInt(update.Message.Chat.ID, 10)
	err := registerChatID(chatID)
	if err != nil {
		panic(err)
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "The chat ID has been registered!",
	})

}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatIDs, err := loadChatIDs()
	if err != nil {
		panic(err)
	}

	for _, chatID := range chatIDs {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Hello, just booted!",
		})
	}

}
