FROM golang:1.21-alpine
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -C ./cmd/kafka/order -o /app/order
RUN CGO_ENABLED=0 GOOS=linux go build -C ./cmd/kafka/warehouse -o /app/warehouse
RUN CGO_ENABLED=0 GOOS=linux go build -C ./cmd/kafka/payment -o /app/payment

