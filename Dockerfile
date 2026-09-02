FROM golang:1.25-alpine AS build

RUN apk add --no-cache build-base pkgconfig vips-dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o ./.bin/go-store ./cmd/app/

FROM alpine:latest

RUN apk add --no-cache vips ca-certificates

COPY --from=build /app/.bin/go-store /app/

EXPOSE 8080

WORKDIR /app/

CMD ["./go-store", "start"]
