FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN go build -o /cmd/serve/js-centralized-wallet ./cmd/serve

CMD ["/cmd/serve/js-centralized-wallet"]
