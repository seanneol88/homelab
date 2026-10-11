# Multi-stage: build with Go toolchain, ship a tiny runtime image
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server .

FROM alpine:3.20
RUN adduser -D -H appuser
USER appuser
WORKDIR /app
COPY --from=build /server /app/server
EXPOSE 8090
CMD ["/app/server"]