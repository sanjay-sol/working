# FROM golang:1.21-alpine

# WORKDIR /app

# COPY . .

# RUN go mod download
# RUN go build -o uploader ./cmd/uploader/main.go

# CMD ["/app/uploader"]

FROM golang:1.21-alpine

WORKDIR /app

COPY . .

RUN go mod download

# Optional: install git if your project depends on it
# RUN apk add --no-cache git

CMD ["go", "run", "./cmd/uploader/main.go"]
