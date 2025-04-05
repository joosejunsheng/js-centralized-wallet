FROM golang:1.24

WORKDIR /app

RUN apt-get update && apt-get install -y postgresql-client && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

RUN go build -o /cmd/serve/js-centralized-wallet ./cmd/serve

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/cmd/serve/js-centralized-wallet"]
