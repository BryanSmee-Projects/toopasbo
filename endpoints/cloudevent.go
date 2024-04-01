package endpoints

import (
	"context"
	"log"
	"os"
	"strconv"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/go-telegram/bot"
)

type CloudEventContext struct {
	Bot *bot.Bot
	Ctx context.Context
}

type CloudEventRouter struct {
	Handlers map[string]func(cloudevents.Event)
}

func (router *CloudEventRouter) DispatchEvent(event cloudevents.Event) {
	if handler, ok := router.Handlers[event.Type()]; ok {
		handler(event)
	} else {
		log.Printf("no handler for event type %s", event.Type())
	}
}

func NewCloudEventRouter() *CloudEventRouter {
	return &CloudEventRouter{
		Handlers: make(map[string]func(cloudevents.Event)),
	}
}

func RegisterHandler(router *CloudEventRouter, eventType string, handler func(cloudevents.Event)) {
	router.Handlers[eventType] = handler
}

func ToCloudEventHandler(handler func(event cloudevents.Event, eventCtx CloudEventContext), eventCtx CloudEventContext) interface{} {
	return func(event cloudevents.Event) {
		handler(event, eventCtx)
	}
}

var requiredEnvVars = []string{"CLOUDEVENT_PORT"}

func StartCloudEventHandler(callback interface{}, ctx context.Context) cloudevents.Client {
	EnsureEnvVars(requiredEnvVars)
	port, err := strconv.Atoi(os.Getenv("CLOUDEVENT_PORT"))
	if err != nil {
		log.Fatalf("failed to convert CLOUDEVENT_PORT to int: %v", err)
	}
	options := []cloudevents.HTTPOption{
		http.WithPort(port),
	}
	c, err := cloudevents.NewClientHTTP(options...)
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}
	if err = c.StartReceiver(ctx, callback); err != nil {
		log.Fatalf("failed to start receiver: %v", err)
	}

	return c
}

func SendCloudEvent(event cloudevents.Event) {
	requiredEnvVars = append(requiredEnvVars, "CLOUDEVENT_TARGET")
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	// Set a target.
	ctx := cloudevents.ContextWithTarget(context.Background(), os.Getenv("CLOUDEVENT_TARGET"))

	// Send that Event.
	if result := c.Send(ctx, event); cloudevents.IsUndelivered(result) {
		log.Fatalf("failed to send, %v", result)
	} else {
		log.Printf("sent: %v", event)
		log.Printf("result: %v", result)
	}
}
