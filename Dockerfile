FROM golang:1.22-alpine AS build

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the Go binary
RUN go build -o todo-api ./main.go

# Final image (minimal)
FROM alpine:3.20

WORKDIR /app

# Copy the compiled binary from build stage
COPY --from=build /app/todo-api .

# Expose the app port
EXPOSE 8080

# Run the binary
CMD ["./todo-api"]
