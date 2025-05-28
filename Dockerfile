FROM golang:1.24.2 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ./server ./cmd/server

CMD ["./server"]