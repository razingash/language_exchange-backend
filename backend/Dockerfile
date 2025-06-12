FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
COPY . .

RUN go mod download

CMD ["go", "run", "./main/app/main.go"]
