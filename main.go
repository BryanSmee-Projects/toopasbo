package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"smee.ovh/toopasbo/endpoints"
	"smee.ovh/toopasbo/gatherers"
	"smee.ovh/toopasbo/transformers"
)

// TODO: Routing for telegram

// TODO: Think of rewrite of the package structure

func serverMode() {
	fmt.Println("Running in server mode")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	wg := &sync.WaitGroup{}
	wg.Add(2)
	telegramBot := endpoints.StartTelegram(ctx)

	cloudEventContext := endpoints.CloudEventContext{
		Bot: telegramBot,
		Ctx: ctx,
	}

	cloudeventRouter := endpoints.NewCloudEventRouter()
	telegramHandler := endpoints.ToCloudEventHandler(endpoints.SendImageAllCloudEventHandler, cloudEventContext)
	endpoints.RegisterHandler(cloudeventRouter, "telegram.sendall", telegramHandler.(func(event.Event)))
	endpoints.StartCloudEventHandler(cloudeventRouter.DispatchEvent, ctx)

	wg.Wait()

}

func jobMode() {
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

func main() {

	boolFlag := flag.Bool("server", false, "Run it in server mode")
	flag.Parse()
	if *boolFlag {
		serverMode()
	}
	jobMode()

}
