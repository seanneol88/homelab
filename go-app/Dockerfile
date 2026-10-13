FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod main.go ./
COPY static ./static/
RUN go mod tidy && CGO_ENABLED=0 go build -o go-app .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /src/go-app /app/go-app
EXPOSE 8080
CMD ["/app/go-app"]
