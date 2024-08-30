package transformers

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"

	"smee.ovh/toopasbo/config"
	"smee.ovh/toopasbo/gatherers"
)

var dallEPromptTemplate = `Photo of a humanoid %s dressed with %s.`

func getDallEPrompt(ctx context.Context, weather gatherers.Weather) (string, error) {
	animal := GetAnimalsByTemperature(weather.MaxTemperature)
	clothes, err := GetClothesForWeather(ctx, weather)
	if err != nil {
		fmt.Printf("Error getting clothes: %v\n", err)
		return "", err
	}
	return fmt.Sprintf(dallEPromptTemplate, animal, clothes), nil
}

func GenerateDallEPicture(ctx context.Context, weather gatherers.Weather) (string, error) {
	prompt, promptErr := getDallEPrompt(ctx, weather)
	if promptErr != nil {
		fmt.Printf("Error getting prompt: %v\n", promptErr)
		return "", promptErr
	}

	appConfig := config.GetAppConfig(ctx)
	client := openai.NewClient(appConfig.OpenaiAPIKey)

	fmt.Println("Creating image...")
	fmt.Println(prompt)

	reqURL := openai.ImageRequest{
		Prompt:         prompt,
		Size:           openai.CreateImageSize1024x1024,
		Model:          openai.CreateImageModelDallE3,
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
	}

	imageResponse, err := client.CreateImage(ctx, reqURL)
	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return "", err
	}

	url := imageResponse.Data[0].URL

	fmt.Println("Got image URL:")
	fmt.Println(url)

	return url, nil

}

func GenerateWeeklyDallEPicture(ctx context.Context, weathers []gatherers.Weather) (string, error) {
	prompt := "Generate an image of the following animals, side by side and from left to right. Don't add any other, they should be 7.\n"
	for _, weather := range weathers {
		p, err := getDallEPrompt(ctx, weather)
		if err != nil {
			fmt.Printf("Error getting prompt: %v\n", err)
			return "", err
		}
		prompt += " - " + strings.TrimSpace(p) + "\n"
	}

	appConfig := config.GetAppConfig(ctx)
	client := openai.NewClient(appConfig.OpenaiAPIKey)

	fmt.Println("Creating image...")
	fmt.Println(prompt)

	reqURL := openai.ImageRequest{
		Prompt:         prompt,
		Size:           openai.CreateImageSize1792x1024,
		Model:          openai.CreateImageModelDallE3,
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
	}

	imageResponse, err := client.CreateImage(ctx, reqURL)
	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return "", err
	}

	url := imageResponse.Data[0].URL

	fmt.Println("Got image URL:")
	fmt.Println(url)

	return url, nil
}
