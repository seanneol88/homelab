FROM golang:1.23-bookworm AS build
WORKDIR /app
COPY go.mod main.go ./
COPY static ./static
RUN CGO_ENABLED=0 go build -o lab .

FROM python:3.12-slim-bookworm
RUN apt-get update \
 && apt-get install -y --no-install-recommends poppler-utils \
 && rm -rf /var/lib/apt/lists/* \
 && pip install --no-cache-dir pdf2docx python-pptx
WORKDIR /app
COPY --from=build /app/lab .
COPY tools ./tools
EXPOSE 8080
CMD ["./lab"]
