package transformers

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"

	"smee.ovh/toopasbo/gatherers"
)

var assistantClothesMessage = `
You are an assistant that select cothes for the weather
The user will provide the temperature, feels like, wind speed and general description
You will provide a suggestion for the clothes
You don't provide multiple suggestions, just one
You don't add any extra information to the suggestion, just the clothes
You don't add any verb, just the clothes
`

var clothesPromptTemplate = `
The current temperature is %d
The minimum temperature is %d and the maximum temperature is %d
Wind speed is %d
General description is %s
You could summarize the weather as %s
`

func GetClothesForWeather(weather gatherers.Weather) (string, error) {
	fmt.Println("Generating clothes suggestion...")
	prompt := fmt.Sprintf(clothesPromptTemplate, weather.CurrentTemperature, weather.MinTemperature, weather.MaxTemperature, weather.WindSpeed, weather.Description, weather.Summary)

	c := openai.NewClient(openaiAPIKey)
	ctx := context.Background()

	req := openai.ChatCompletionRequest{
		Model:     openai.GPT4,
		MaxTokens: 128,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: assistantClothesMessage,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Stream: false,
	}
	resp, err := c.CreateChatCompletion(ctx, req)
	if err != nil {
		fmt.Printf("Completion error: %v\n", err)
		return "", err
	}

	suggestion := strings.TrimSpace(resp.Choices[0].Message.Content)

	fmt.Println("Got suggestion:")
	fmt.Println(suggestion)

	return suggestion, nil
}
