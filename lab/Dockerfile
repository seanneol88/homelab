FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod main.go ./
COPY static ./static/
RUN CGO_ENABLED=0 go build -o lab .
FROM alpine:3.20
WORKDIR /app
COPY --from=build /src/lab /app/lab
EXPOSE 8080
CMD ["/app/lab"]
