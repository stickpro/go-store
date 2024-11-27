FROM golang:1.23-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o ./.bin/go-store ./cmd/app/

FROM alpine:latest

COPY --from=build /app/.bin/go-store /app/

EXPOSE 8080

WORKDIR /app/

CMD ["./go-store", "start"]
