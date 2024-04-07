package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/cloudevents/sdk-go/v2/event"
	"smee.ovh/toopasbo/endpoints"
	"smee.ovh/toopasbo/jobs"
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

func main() {

	boolFlag := flag.Bool("server", false, "Run it in server mode")
	jobFlag := flag.String("job", "", "Run this job")
	flag.Parse()
	if *boolFlag {
		serverMode()
	}
	if *jobFlag != "" {
		jobs.RunJob(*jobFlag)
	} else {
		panic("No mode selected for the application. Please select either -server or -job flag.")
	}
}
