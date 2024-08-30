build:
	rm -rf dist
	mkdir -p dist/webhook
	mkdir -p dist/daily
	mkdir -p dist/weekly
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/webhook/bootstrap
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/daily/bootstrap jobs/daily/daily.go
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/weekly/bootstrap jobs/weekly/weekly.go

deploy:
	make build
	cd terraform && terraform apply -auto-approve