package transformers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"smee.ovh/toopasbo/config"
	"smee.ovh/toopasbo/gatherers"
)

type MidJourneyResponse struct {
	ID    string `json:"id"`
	Flags int    `json:"flags"`
	Hash  string `json:"hash"`
	URI   string `json:"uri"`
}

type MidJourneyRequest struct {
	Prompt string `json:"prompt"`
}

var midjourneyPromptTemplate = `Full body portrait of a humanoid %s dressed with %s, standing in %s. The weather is %s. --no human`

func getMidjourneyPrompt(ctx context.Context, weather gatherers.Weather) (string, error) {
	animal := GetAnimalsByTemperature(weather.MaxTemperature)
	clothes, err := GetClothesForWeather(ctx, weather)
	if err != nil {
		fmt.Printf("Error getting clothes: %v\n", err)
		return "", err
	}
	return fmt.Sprintf(midjourneyPromptTemplate, animal, clothes, weather.Location, weather.Description), nil
}

func generateSimpleImage(ctx context.Context, prompt string) (string, error) {
	fmt.Println("Creating image...")
	fmt.Println(prompt)

	// do http request to midjourney api
	requestBody := MidJourneyRequest{
		Prompt: prompt,
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Printf("Error marshalling request body: %v\n", err)
		return "", err
	}

	appConfig := config.GetAppConfig(ctx)

	resp, err := http.Post(appConfig.MidjourneyApiUrl+"/simpleimage", "application/json", strings.NewReader(string(requestBodyBytes)))
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()
	// unmarshal response
	var midJourneyResponse MidJourneyResponse
	err = json.NewDecoder(resp.Body).Decode(&midJourneyResponse)
	if err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		return "", err
	}

	url := midJourneyResponse.URI
	return url, nil
}

func GenerateMidjourneyPicture(ctx context.Context, weather gatherers.Weather) (string, error) {
	prompt, promptErr := getMidjourneyPrompt(ctx, weather)
	if promptErr != nil {
		fmt.Printf("Error getting prompt: %v\n", promptErr)
		return "", promptErr
	}

	url, err := generateSimpleImage(ctx, prompt)
	if err != nil {
		fmt.Printf("Error generating image: %v\n", err)
		return "", err
	}

	return url, nil
}

func GenerateWeeklyMidjourneyPicture(ctx context.Context, weathers []gatherers.Weather) (string, error) {
	prompt := "Generate an image of the following animals, side by side and from left to right. Don't add any other, they should be 7.\n"
	for _, weather := range weathers {
		p, err := getDallEPrompt(ctx, weather)
		if err != nil {
			fmt.Printf("Error getting prompt: %v\n", err)
			return "", err
		}
		prompt += " - " + strings.TrimSpace(p) + "\n"
	}

	prompt += "--no human --ar 7:4"

	url, err := generateSimpleImage(ctx, prompt)
	if err != nil {
		fmt.Printf("Error generating image: %v\n", err)
		return "", err
	}

	return url, nil
}
