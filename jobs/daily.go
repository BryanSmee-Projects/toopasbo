package jobs

import (
	"fmt"
	"os"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"smee.ovh/toopasbo/endpoints"
	"smee.ovh/toopasbo/gatherers"
	"smee.ovh/toopasbo/transformers"
)

func dailyMode() {
	fmt.Println("Running in job mode")
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

	var imageUrl, err = transformers.GenerateDallEPicture(weather)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(imageUrl)

	event := cloudevents.NewEvent()
	eventData := endpoints.TelegramImageEventData{
		ImageLink: imageUrl,
		Weather:   endpoints.WeatherToTelegramText(weather),
	}
	event.SetData(cloudevents.ApplicationJSON, eventData)
	event.SetSource("job/sendall")
	event.SetType("telegram.sendall")
	endpoints.SendCloudEvent(event)
}
