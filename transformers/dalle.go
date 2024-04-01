package transformers

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"

	"smee.ovh/toopasbo/gatherers"
)

var dallEPromptTemplate = `
Photo of a %s dressed with %s in a %s weather.
`

func getDallEPrompt(weather gatherers.Weather) (string, error) {
	animal := GetAnimalsByTemperature(weather.MaxTemperature)
	clothes, err := GetClothesForWeather(weather)
	if err != nil {
		fmt.Printf("Error getting clothes: %v\n", err)
		return "", err
	}
	return fmt.Sprintf(dallEPromptTemplate, animal, clothes, weather.Description), nil
}

func GenerateDallEPicture(weather gatherers.Weather) (string, error) {
	prompt, promptErr := getDallEPrompt(weather)
	if promptErr != nil {
		fmt.Printf("Error getting prompt: %v\n", promptErr)
		return "", promptErr
	}

	client := openai.NewClient(openaiAPIKey)
	ctx := context.Background()

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
