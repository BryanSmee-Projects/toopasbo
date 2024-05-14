package jobs

import (
	"fmt"
	"os"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"smee.ovh/toopasbo/endpoints"
	"smee.ovh/toopasbo/gatherers"
	"smee.ovh/toopasbo/transformers"
)

var weekDays = []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}

func weathersToText(weathers []gatherers.Weather) string {
	var result string = "Weather for the week:\n"
	for day, weather := range weathers {
		result += "- " + weekDays[day] + ": " + endpoints.WeatherToShortText(weather) + "\n"
	}
	return result
}

func weeklyJob() {
	fmt.Println("Running in job mode")
	var zipCode = "14700"
	var countryCode = "CZ"

	var position, positionErr = gatherers.GetPosition(zipCode, countryCode)
	if positionErr != nil {
		fmt.Println(positionErr)
		os.Exit(1)
	}

	var weather, weatherErr = gatherers.GetWeatherForWeek(position)

	if weatherErr != nil {
		fmt.Println(weatherErr)
		os.Exit(1)
	}

	var imageUrl, err = transformers.GenerateWeeklyMidjourneyPicture(weather)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(imageUrl)
	weatherText := weathersToText(weather)

	event := cloudevents.NewEvent()
	eventData := endpoints.TelegramImageEventData{
		ImageLink: imageUrl,
		Weather:   weatherText,
	}
	event.SetData(cloudevents.ApplicationJSON, eventData)
	event.SetSource("job/sendall")
	event.SetType("telegram.sendall")
	endpoints.SendCloudEvent(event)
}
