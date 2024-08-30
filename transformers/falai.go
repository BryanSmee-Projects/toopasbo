package transformers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"smee.ovh/toopasbo/config"
	"smee.ovh/toopasbo/gatherers"
)

type FalAIClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	ctx        context.Context
}

type FalImageResponse struct {
	Images []struct {
		URL         string `json:"url"`
		ContentType string `json:"content_type"`
	} `json:"images"`
	Prompt string `json:"prompt"`
}

func NewFalAIClient(ctx context.Context, model string) (*FalAIClient, error) {
	appConfig := config.GetAppConfig(ctx)
	if appConfig.FalAIAPIKey == "" {
		return nil, fmt.Errorf("FalAIAPIKey is not set")
	}

	return &FalAIClient{
		BaseURL:    "https://fal.run/" + model,
		APIKey:     appConfig.FalAIAPIKey,
		HTTPClient: &http.Client{},
		ctx:        ctx,
	}, nil
}

var falPromptTemplate = `Full body portrait of a humanoid %s dressed with %s, standing in %s. The weather is %s. %d° is written on it's clothes.`

func (c *FalAIClient) getFalPrompt(weather gatherers.Weather) (string, error) {
	animal := GetAnimalsByTemperature(weather.MaxTemperature)
	clothes, err := GetClothesForWeather(c.ctx, weather)
	if err != nil {
		fmt.Printf("Error getting clothes: %v\n", err)
		return "", err
	}
	return fmt.Sprintf(falPromptTemplate, animal, clothes, weather.Location, weather.Description, weather.MaxTemperature), nil
}

func (c *FalAIClient) GenerateWeatherImage(weather gatherers.Weather) (string, error) {
	prompt, err := c.getFalPrompt(weather)
	if err != nil {
		fmt.Printf("Error getting prompt: %v\n", err)
		return "", err
	}

	image, err := c.GenerateImage(prompt, "portrait_4_3")
	if err != nil {
		fmt.Printf("Error generating image: %v\n", err)
		return "", err
	}

	return image, nil
}

func (c *FalAIClient) GenerateImage(prompt string, image_size string) (string, error) {
	requestBody, err := json.Marshal(map[string]string{
		"prompt":     prompt,
		"image_size": image_size,
	})
	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Authorization", "Key "+c.APIKey)
	req.Header.Add("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	var imageResponse FalImageResponse
	if err := json.Unmarshal(body, &imageResponse); err != nil {
		return "", fmt.Errorf("error unmarshalling response JSON: %w", err)
	}

	if len(imageResponse.Images) == 0 {
		return "", fmt.Errorf("no images found in response")
	}

	return imageResponse.Images[0].URL, nil
}

func (c *FalAIClient) GenerateWeeklyPicture(weathers []gatherers.Weather) (string, error) {
	prompt := "Generate an image of the following animals, side by side and from left to right. Don't add any other, they should be 7.\n"
	for _, weather := range weathers {
		p, err := getDallEPrompt(c.ctx, weather)
		if err != nil {
			fmt.Printf("Error getting prompt: %v\n", err)
			return "", err
		}
		prompt += " - " + strings.TrimSpace(p) + "\n"
	}

	url, err := c.GenerateImage(prompt, "landscape_16_9")
	if err != nil {
		fmt.Printf("Error generating image: %v\n", err)
		return "", err
	}

	return url, nil

}
