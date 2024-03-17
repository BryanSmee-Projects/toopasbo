FROM golang:1.22 AS builder

WORKDIR /app

COPY . .

RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o toopasbo .

FROM gcr.io/distroless/static

COPY --from=builder /app/toopasbo /

ENTRYPOINT ["/toopasbo"]