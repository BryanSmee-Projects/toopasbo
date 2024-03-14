package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"smee.ovh/toopasbo/endpoints"
	"smee.ovh/toopasbo/gatherers"
	"smee.ovh/toopasbo/transformers"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	b := endpoints.StartTelegram(ctx)
	var zipCode = "14700"
	var countryCode = "CZ"

	var position, positionErr = gatherers.GetPosition(zipCode, countryCode)
	if positionErr != nil {
		fmt.Println(positionErr)
		os.Exit(1)
	}

	var weather, weatherErr = gatherers.GetWeather(position)

	if weatherErr != nil {
		fmt.Println(weatherErr)
		os.Exit(1)
	}

	var imagePath, err = transformers.GenerateDallEPicture(weather)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(imagePath)
	endpoints.SendImageToTelegram(imagePath, weather, b, ctx)

}
