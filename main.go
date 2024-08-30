package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"smee.ovh/toopasbo/config"
	"smee.ovh/toopasbo/endpoints"
)

// TODO: Routing for telegram

// TODO: Think of rewrite of the package structure

func serverMode() {
	fmt.Println("Running in server mode")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	is_running_in_aws := os.Getenv("RUNTIME_ENVIRONMENT") == "aws"
	var appConfig config.Config
	if is_running_in_aws {
		fmt.Println("Running in AWS")
		appConfig = config.NewConfigFromAWS()
	} else {
		fmt.Println("Running in local")
		appConfig = config.NewConfigFromEnv()
	}

	ctx = context.WithValue(ctx, config.AppConfigContextKey, appConfig)

	wc, err := endpoints.NewWebhookClient(ctx)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", wc.HttpHandler())
	if is_running_in_aws {
		lambda.Start(httpadapter.New(http.DefaultServeMux).ProxyWithContext)
	} else {
		fmt.Println("Running local server")
		http.ListenAndServe("0.0.0.0:8080", nil)
	}
}

func main() {
	serverMode()
}
