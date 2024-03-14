package transformers

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image/png"
	"os"

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

	reqBase64 := openai.ImageRequest{
		Prompt:         prompt,
		Size:           openai.CreateImageSize1024x1024,
		Model:          openai.CreateImageModelDallE3,
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		N:              1,
	}

	respBase64, err := client.CreateImage(ctx, reqBase64)
	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return "", err
	}

	imgBytes, err := base64.StdEncoding.DecodeString(respBase64.Data[0].B64JSON)
	if err != nil {
		fmt.Printf("Base64 decode error: %v\n", err)
		return "", err
	}

	r := bytes.NewReader(imgBytes)
	imgData, err := png.Decode(r)
	if err != nil {
		fmt.Printf("PNG decode error: %v\n", err)
		return "", err
	}

	filedir, err := os.MkdirTemp("", "toopasbo-")
	if err != nil {
		fmt.Printf("Temp dir creation error: %v\n", err)
		return "", err
	}

	filepath := filedir + "/dalle.png"

	file, err := os.Create(filepath)
	if err != nil {
		fmt.Printf("File creation error: %v\n", err)
		return "", err
	}
	defer file.Close()

	if err := png.Encode(file, imgData); err != nil {
		fmt.Printf("PNG encode error: %v\n", err)
		return "", err
	}

	return filepath, nil

}
