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

type FreepikImage struct {
	Size string `json:"size"`
}

type FreepikRequest struct {
	Image     FreepikImage `json:"image"`
	NumImages int          `json:"num_images"`
	Mode      string       `json:"mode"`
	Prompt    string       `json:"prompt"`
}

type FreepikResponse struct {
	Data []struct {
		Base64  string `json:"base64"`
		HasNSFW bool   `json:"has_nsfw"`
	} `json:"data"`
	Meta map[string]interface{} `json:"meta"`
}

type FreepickErrorResp struct {
	Message string `json:"message"`
}

type FreepikClient struct {
	APIKey string
	APIUrl string
	ctx    context.Context
}

func NewFreepikClient(ctx context.Context) *FreepikClient {
	appConfig := config.GetAppConfig(ctx)
	return &FreepikClient{APIKey: appConfig.FreepikAPIKey, APIUrl: "https://api.freepik.com/v1/ai/text-to-image", ctx: ctx}
}

var freepikPromptTemplate = `Full body portrait of a humanoid %s dressed with %s, standing in %s. The weather is %s.`

func (f *FreepikClient) getFreepikPrompt(weather gatherers.Weather) (string, error) {
	animal := GetAnimalsByTemperature(weather.MaxTemperature)
	clothes, err := GetClothesForWeather(f.ctx, weather)
	if err != nil {
		fmt.Printf("Error getting clothes: %v\n", err)
		return "", err
	}
	return fmt.Sprintf(freepikPromptTemplate, animal, clothes, weather.Location, weather.Description), nil
}

func (f *FreepikClient) GenerateWeatherImage(weather gatherers.Weather) (*FreepikResponse, error) {
	prompt, err := f.getFreepikPrompt(weather)
	if err != nil {
		fmt.Printf("Error getting prompt: %v\n", err)
		return nil, err
	}

	image, err := f.GenerateImage(prompt)
	if err != nil {
		fmt.Printf("Error generating image: %v\n", err)
		return nil, err
	}

	return image, nil

}

func (f *FreepikClient) GenerateImage(prompt string) (*FreepikResponse, error) {
	requestBody := FreepikRequest{
		Image:     FreepikImage{Size: "portrait"},
		NumImages: 1,
		Mode:      "flux-realism",
		Prompt:    prompt,
	}

	httpClient := &http.Client{}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", f.APIUrl, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Accept-Language", "en-US")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-freepik-api-key", f.APIKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errorMsg := FreepickErrorResp{}
		err = json.NewDecoder(resp.Body).Decode(&errorMsg)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error: %v", errorMsg.Message)

	}

	var freepikResponse FreepikResponse
	err = json.NewDecoder(resp.Body).Decode(&freepikResponse)
	if err != nil {
		return nil, err
	}

	return &freepikResponse, nil
}
